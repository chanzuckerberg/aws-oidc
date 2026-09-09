package portal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
)

// errNoIdentity is returned when no authenticated user can be determined.
var errNoIdentity = errors.New("no authenticated user")

const idTokenCookiePrefix = "IdToken-"

// IdentityResolver extracts the current user from the gateway's verified OIDC ID token.
type IdentityResolver struct {
	devSub      string
	devEmail    string
	adminGroups map[string]bool

	// verifyIDToken verifies the X-ID-Token JWT and returns the user's sub, email, and groups.
	// It is a struct field so tests can stub it without a live issuer.
	verifyIDToken func(ctx context.Context, rawToken string) (sub, email string, groups []string, err error)
}

// idTokenClaims are the claims the portal reads from the forwarded OIDC ID token.
type idTokenClaims struct {
	Email  string   `json:"email"`
	Groups []string `json:"teamGroups"`
}

// NewIdentityResolver builds a resolver that trusts the Envoy gateway OIDC proxy. Following the
// convention argus uses, it creates an OIDC provider from the issuer at boot (which discovers
// the JWKS endpoint) and verifies every forwarded token's signature against those keys.
//
// Config:
//   - issuerURL: the Okta issuer (e.g. https://czi.okta.com). Its JWKS signs the tokens.
//   - clientID: the OAuth client the gateway authenticates with. The ID token audience must
//     match it, which binds the token to our gateway app.
//
// Env:
//   - PORTAL_DEV_SUB / PORTAL_DEV_EMAIL: act as a fixed user (treated as admin) without a real
//     login. Handy in rdev to impersonate any user for testing.
//   - PORTAL_ADMIN_GROUPS: comma-separated Okta groups that grant the admin view.
func NewIdentityResolver(ctx context.Context, issuerURL, clientID string) (*IdentityResolver, error) {
	if issuerURL == "" {
		return nil, errors.New("issuer URL is required")
	}
	if clientID == "" {
		return nil, errors.New("client id is required")
	}

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("creating oidc provider: %w", err)
	}

	// ID token verifier: the gateway's OIDC client ID is the audience.
	idVerifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	devSub := os.Getenv("PORTAL_DEV_SUB")
	adminGroups := parseAdminGroups(os.Getenv("PORTAL_ADMIN_GROUPS"))
	slog.Info("portal identity resolver ready",
		"issuer", issuerURL,
		"expected_token_client_id", clientID,
		"admin_groups", sortedKeys(adminGroups),
		"dev_override", devSub != "",
	)

	return &IdentityResolver{
		devSub:      devSub,
		devEmail:    os.Getenv("PORTAL_DEV_EMAIL"),
		adminGroups: adminGroups,
		verifyIDToken: func(ctx context.Context, raw string) (string, string, []string, error) {
			tok, err := idVerifier.Verify(ctx, raw)
			if err != nil {
				return "", "", nil, fmt.Errorf("verifying id token: %w", err)
			}
			var claims idTokenClaims
			err = tok.Claims(&claims)
			if err != nil {
				return "", "", nil, fmt.Errorf("reading id token claims: %w", err)
			}
			return tok.Subject, claims.Email, claims.Groups, nil
		},
	}, nil
}

// Resolve returns the current user, or errNoIdentity if none can be determined.
func (ir *IdentityResolver) Resolve(ctx context.Context, r *http.Request) (*identity.User, error) {
	if ir.devSub != "" {
		slog.Info("portal identity taken from PORTAL_DEV_SUB override", "sub", ir.devSub, "email", ir.devEmail)
		return &identity.User{
			Sub:         ir.devSub,
			Email:       ir.devEmail,
			Admin:       true,
			AdminReason: "Admin access granted by PORTAL_DEV_SUB",
		}, nil
	}

	idTokens := idTokenCandidates(r)
	if len(idTokens) == 0 {
		slog.Info("portal found no ID token",
			"header_names", headerNames(r),
			"cookie_names", requestCookieNames(r),
		)
	}

	for _, idToken := range idTokens {
		if ir.verifyIDToken == nil {
			break
		}
		sub, email, groups, err := ir.verifyIDToken(ctx, idToken.raw)
		if err != nil {
			slog.Warn("portal rejected an ID token", "source", idToken.source, "error", err, describeToken(idToken.raw))
		} else {
			user := &identity.User{Sub: sub, Email: email, Groups: groups}
			if group := matchingAdminGroup(groups, ir.adminGroups); group != "" {
				user.Admin = true
				user.AdminReason = "Admin through Okta group " + group
			}
			slog.Info("portal resolved user from ID token", "source", idToken.source, "sub", sub, "email", email, "groups", groups, "admin", user.Admin)
			return user, nil
		}
	}

	slog.Warn("portal request has no valid ID token",
		"header_names", headerNames(r),
		"cookie_names", requestCookieNames(r),
	)
	return nil, fmt.Errorf("%w: no valid ID token on request (headers: %s; cookies: %s)",
		errNoIdentity,
		strings.Join(headerNames(r), ", "),
		strings.Join(requestCookieNames(r), ", "),
	)
}

func matchingAdminGroup(groups []string, adminGroups map[string]bool) string {
	for _, g := range groups {
		group := strings.TrimSpace(g)
		if adminGroups[group] {
			return group
		}
	}
	return ""
}

func describeToken(raw string) slog.Attr {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return slog.String("token", "not a jwt, so the gateway may be forwarding an opaque token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return slog.String("token", "jwt payload is not base64url")
	}
	claims := struct {
		Issuer    string          `json:"iss"`
		ClientID  string          `json:"cid"`
		UserID    string          `json:"uid"`
		Subject   string          `json:"sub"`
		Audience  json.RawMessage `json:"aud"`
		ExpiresAt int64           `json:"exp"`
	}{}
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return slog.String("token", "jwt payload is not json")
	}

	attrs := []any{
		slog.String("iss", claims.Issuer),
		slog.String("cid", claims.ClientID),
		slog.String("uid", claims.UserID),
		slog.String("sub", claims.Subject),
		slog.String("aud", string(claims.Audience)),
	}
	if claims.ExpiresAt != 0 {
		expiry := time.Unix(claims.ExpiresAt, 0).UTC()
		attrs = append(attrs,
			slog.Time("exp", expiry),
			slog.Bool("expired", time.Now().After(expiry)),
		)
	}
	return slog.Group("unverified_token_claims", attrs...)
}

func headerNames(r *http.Request) []string {
	names := make([]string, 0, len(r.Header))
	for name := range r.Header {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type idTokenCandidate struct {
	raw    string
	source string
}

func idTokenCandidates(r *http.Request) []idTokenCandidate {
	cookies := r.Cookies()
	candidates := make([]idTokenCandidate, 0, len(cookies)+1)
	if raw := r.Header.Get("X-Id-Token"); raw != "" {
		candidates = append(candidates, idTokenCandidate{raw: raw, source: "header"})
	}
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, idTokenCookiePrefix) && cookie.Value != "" {
			candidates = append(candidates, idTokenCandidate{raw: cookie.Value, source: "cookie:" + cookie.Name})
		}
	}
	return candidates
}

func requestCookieNames(r *http.Request) []string {
	cookies := r.Cookies()
	names := make([]string, 0, len(cookies))
	for _, c := range cookies {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	return names
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func parseAdminGroups(raw string) map[string]bool {
	groups := map[string]bool{}
	for _, g := range strings.Split(raw, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			groups[g] = true
		}
	}
	return groups
}
