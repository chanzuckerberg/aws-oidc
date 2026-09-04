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
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	githubDefaultAPI    = "https://api.github.com"
	githubReposCacheTTL = 5 * time.Minute
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
// reject repositories a thread pod could not actually clone. It caches the accessible set so a
// burst of keystrokes does not hammer the GitHub API.
type GitHubApp struct {
	appID         string
	key           *rsa.PrivateKey
	apiURL        string
	installations []string
	httpClient    *http.Client
	cacheTTL      time.Duration

	mu         sync.Mutex
	tokens     map[string]cachedToken
	repos      []string
	reposByID  map[string]bool
	fetchedAt  time.Time
	refreshing chan struct{}
	lastErr    error
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
	app := &GitHubApp{
		appID:         appID,
		key:           key,
		apiURL:        strings.TrimRight(apiURL, "/"),
		installations: installations,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		cacheTTL:      githubReposCacheTTL,
		tokens:        map[string]cachedToken{},
	}
	// Warm the cache now, so the first search hits it instead of enumerating thousands of
	// repositories while the user waits.
	app.mu.Lock()
	app.ensureRefreshLocked()
	app.mu.Unlock()
	return app, nil
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

func (g *GitHubApp) Suggest(ctx context.Context, query string, limit int) ([]string, error) {
	all, _, err := g.accessible(ctx)
	if err != nil {
		return nil, err
	}
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

func (g *GitHubApp) Reachable(ctx context.Context, repo string) (bool, error) {
	_, byID, err := g.accessible(ctx)
	if err != nil {
		return false, err
	}
	return byID[strings.ToLower(strings.TrimSpace(repo))], nil
}

// accessible returns the union of every installation's repositories as sorted "owner/repo"
// strings, plus a lowercase membership set. The set is cached for cacheTTL and refreshed in
// the background. Enumerating an org's repositories takes seconds and spans many pages, so
// doing it on the request's context once per keystroke made every search cancel the previous
// one and never finish. Detaching the refresh means a keystroke that gives up still leaves the
// enumeration running to warm the cache, and concurrent searches share one refresh.
func (g *GitHubApp) accessible(ctx context.Context) ([]string, map[string]bool, error) {
	g.mu.Lock()
	fresh := g.repos != nil && time.Since(g.fetchedAt) < g.cacheTTL
	if fresh {
		repos, byID := g.repos, g.reposByID
		g.mu.Unlock()
		return repos, byID, nil
	}
	wait := g.ensureRefreshLocked()
	// Serve a stale set immediately rather than blocking a keystroke on a full refresh.
	if g.repos != nil {
		repos, byID := g.repos, g.reposByID
		g.mu.Unlock()
		return repos, byID, nil
	}
	g.mu.Unlock()

	// No cache yet: wait for the in-flight refresh, but give up if the caller does. The refresh
	// runs on its own context, so giving up here does not abort it.
	select {
	case <-wait:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	g.mu.Lock()
	repos, byID, err := g.repos, g.reposByID, g.lastErr
	g.mu.Unlock()
	if repos == nil {
		if err == nil {
			err = fmt.Errorf("no repositories available")
		}
		return nil, nil, err
	}
	return repos, byID, nil
}

// ensureRefreshLocked starts a background refresh unless one is already running, and returns a
// channel closed when it finishes. The caller must hold g.mu.
func (g *GitHubApp) ensureRefreshLocked() chan struct{} {
	if g.refreshing != nil {
		return g.refreshing
	}
	done := make(chan struct{})
	g.refreshing = done
	go g.refresh(done)
	return done
}

func (g *GitHubApp) refresh(done chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	byID := map[string]bool{}
	var all []string
	var refreshErr error
	for _, id := range g.installations {
		repos, err := g.listInstallationRepos(ctx, id)
		if err != nil {
			refreshErr = err
			break
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
	if refreshErr == nil {
		g.repos, g.reposByID, g.fetchedAt, g.lastErr = all, byID, time.Now(), nil
	} else {
		g.lastErr = refreshErr
	}
	g.refreshing = nil
	g.mu.Unlock()
	close(done)
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
