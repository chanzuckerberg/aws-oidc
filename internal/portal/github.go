package portal

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	githubDefaultAPI           = "https://api.github.com"
	githubReposRefreshInterval = 10 * time.Minute
)

// repoSuggester suggests and validates the repositories an agent may clone. The GitHub App
// client implements it; the Config field is left nil where the portal has no GitHub
// credentials, which hides the Repositories tab rather than offering a search that cannot work.
type repoSuggester interface {
	// Suggest returns up to limit repositories whose "owner/repo" contains query, drawn from
	// the installations the fleet's GitHub App can reach.
	Suggest(ctx context.Context, query string, limit int) ([]string, error)
	// Reachable reports whether repo ("owner/repo") is accessible to one of those installations.
	Reachable(ctx context.Context, repo string) (bool, error)
}

var _ repoSuggester = (*GitHubApp)(nil)

// GitHubApp lists and validates the repositories the shared GitHub App can reach across every
// installation the fleet routes to (the default installation plus the ones in the operator's
// installation map). The portal uses it to power the Repositories tab's type-ahead and to
// reject repositories a workspace pod could not actually clone. A background loop started by Start
// holds the accessible set in memory; the type-ahead reads only that cache and never calls the
// GitHub API on the request path.
type GitHubApp struct {
	appID           string
	key             *rsa.PrivateKey
	apiURL          string
	installations   []string
	httpClient      *http.Client
	refreshInterval time.Duration

	mu        sync.Mutex
	tokens    map[string]cachedToken
	repos     []string
	reposByID map[string]bool
	loaded    bool
	lastErr   error
}

type cachedToken struct {
	token  string
	expiry time.Time
}

// NewGitHubApp builds a client for the shared App. privateKeyPEM is the App's PEM-encoded RSA
// key. defaultInstallation is the installation every owner falls back to; installationMap is
// the operator's "owner=id,owner2=id2" routing, whose ids are added to the reachable set.
// Returns nil, nil when appID or the key is absent, so callers can treat missing credentials
// as "feature off" without special-casing.
func NewGitHubApp(appID string, privateKeyPEM []byte, defaultInstallation, installationMap, apiURL string) (*GitHubApp, error) {
	if strings.TrimSpace(appID) == "" || len(privateKeyPEM) == 0 {
		return nil, nil
	}
	key, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing github app private key: %w", err)
	}
	if apiURL == "" {
		apiURL = githubDefaultAPI
	}
	seen := map[string]bool{}
	var installations []string
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		installations = append(installations, id)
	}
	add(defaultInstallation)
	for _, entry := range strings.FieldsFunc(installationMap, func(r rune) bool { return r == ',' || r == ' ' }) {
		_, id, ok := strings.Cut(entry, "=")
		if ok {
			add(id)
		}
	}
	if len(installations) == 0 {
		return nil, fmt.Errorf("github app has no installations: set a default installation id")
	}
	return &GitHubApp{
		appID:           appID,
		key:             key,
		apiURL:          strings.TrimRight(apiURL, "/"),
		installations:   installations,
		httpClient:      &http.Client{Timeout: 15 * time.Second},
		refreshInterval: githubReposRefreshInterval,
		tokens:          map[string]cachedToken{},
	}, nil
}

func parseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("key is neither PKCS1 nor PKCS8: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return key, nil
}

func (g *GitHubApp) Suggest(_ context.Context, query string, limit int) ([]string, error) {
	all, _, _, _ := g.accessible()
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]string, 0, limit)
	for _, repo := range all {
		if q == "" || strings.Contains(strings.ToLower(repo), q) {
			out = append(out, repo)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (g *GitHubApp) Reachable(_ context.Context, repo string) (bool, error) {
	_, byID, loaded, lastErr := g.accessible()
	if !loaded {
		if lastErr != nil {
			return false, fmt.Errorf("repository list not ready: %w", lastErr)
		}
		return false, fmt.Errorf("repository list is still loading; try again in a moment")
	}
	return byID[strings.ToLower(strings.TrimSpace(repo))], nil
}

// Start refreshes the accessible-repository cache once now and then every refreshInterval,
// until ctx is done. Enumerating an org's repositories takes seconds and spans many pages, and
// repositories rarely change, so the type-ahead reads only this cache and leaves the GitHub API
// calls to the background loop.
func (g *GitHubApp) Start(ctx context.Context) {
	go g.refreshLoop(ctx)
}

func (g *GitHubApp) refreshLoop(ctx context.Context) {
	err := g.refresh(ctx)
	if err != nil {
		slog.Warn("initial repository cache refresh failed", "error", err)
	}
	ticker := time.NewTicker(g.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := g.refresh(ctx)
			if err != nil {
				slog.Warn("repository cache refresh failed", "error", err)
			}
		}
	}
}

// accessible returns the cached union of every installation's repositories as sorted
// "owner/repo" strings, a lowercase membership set, whether a refresh has ever succeeded, and
// the last refresh error. It reads only memory; the background loop keeps the cache current.
func (g *GitHubApp) accessible() ([]string, map[string]bool, bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.repos, g.reposByID, g.loaded, g.lastErr
}

// refresh enumerates every installation's repositories and replaces the cache. On error it
// keeps the previous cache, so a transient GitHub failure does not empty the type-ahead.
func (g *GitHubApp) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	byID := map[string]bool{}
	var all []string
	for _, id := range g.installations {
		repos, err := g.listInstallationRepos(ctx, id)
		if err != nil {
			g.mu.Lock()
			g.lastErr = err
			g.mu.Unlock()
			return err
		}
		for _, repo := range repos {
			key := strings.ToLower(repo)
			if !byID[key] {
				byID[key] = true
				all = append(all, repo)
			}
		}
	}
	sort.Strings(all)

	g.mu.Lock()
	g.repos, g.reposByID, g.loaded, g.lastErr = all, byID, true, nil
	g.mu.Unlock()
	return nil
}

func (g *GitHubApp) listInstallationRepos(ctx context.Context, installationID string) ([]string, error) {
	token, err := g.installationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	var repos []string
	for page := 1; page <= 100; page++ {
		u := fmt.Sprintf("%s/installation/repositories?per_page=100&page=%d", g.apiURL, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, fmt.Errorf("building repositories request: %w", err)
		}
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		resp, err := g.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("listing installation %s repositories: %w", installationID, err)
		}
		names, more, err := decodeRepoPage(resp)
		if err != nil {
			return nil, err
		}
		repos = append(repos, names...)
		if !more {
			break
		}
	}
	return repos, nil
}

func decodeRepoPage(resp *http.Response) (repos []string, more bool, err error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, false, fmt.Errorf("reading repositories response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("listing repositories: github returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out struct {
		Repositories []struct {
			FullName string `json:"full_name"`
		} `json:"repositories"`
	}
	err = json.Unmarshal(body, &out)
	if err != nil {
		return nil, false, fmt.Errorf("decoding repositories response: %w", err)
	}
	for _, r := range out.Repositories {
		if r.FullName != "" {
			repos = append(repos, r.FullName)
		}
	}
	return repos, len(out.Repositories) >= 100, nil
}

func (g *GitHubApp) installationToken(ctx context.Context, installationID string) (string, error) {
	g.mu.Lock()
	cached, ok := g.tokens[installationID]
	g.mu.Unlock()
	if ok && time.Until(cached.expiry) > time.Minute {
		return cached.token, nil
	}

	jwt, err := g.appJWT()
	if err != nil {
		return "", err
	}
	u := fmt.Sprintf("%s/app/installations/%s/access_tokens", g.apiURL, installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return "", fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("minting installation token: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("minting installation token: github returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	err = json.Unmarshal(body, &out)
	if err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}

	g.mu.Lock()
	g.tokens[installationID] = cachedToken{token: out.Token, expiry: out.ExpiresAt}
	g.mu.Unlock()
	return out.Token, nil
}

func (g *GitHubApp) appJWT() (string, error) {
	now := time.Now()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := fmt.Sprintf(`{"iat":%d,"exp":%d,"iss":%q}`, now.Add(-60*time.Second).Unix(), now.Add(9*time.Minute).Unix(), g.appID)
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	signingInput := header + "." + payload
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, g.key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("signing app jwt: %w", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}
