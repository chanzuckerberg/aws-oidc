package portal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// User is the authenticated portal user. Sub is the Okta user id (the "00u" value) and is what
// entitlements and agent ownership are keyed on.
type User struct {
	Sub   string
	Email string
	Admin bool
}

// errNoIdentity is returned when no authenticated user can be determined.
var errNoIdentity = errors.New("no authenticated user")

// IdentityResolver extracts the current user from a request. The Envoy gateway OIDC proxy
// authenticates the browser and forwards the user's Okta access token in the Authorization
// header (oidcProxyGateway.forwardAccessToken). That token carries the Okta user id in its
// uid claim but not group membership, so groups are read from Okta's userinfo endpoint. A dev
// override (PORTAL_DEV_SUB) short-circuits the whole flow for rdev testing.
type IdentityResolver struct {
	devSub      string
	devEmail    string
	adminGroups map[string]bool

	// verifyToken checks the forwarded access token and returns the Okta user id (its uid
	// claim). It is a struct field so tests can stub it without a live issuer.
	verifyToken func(ctx context.Context, rawToken string) (userID string, err error)
	// fetchUserInfo returns the email and group memberships for the token's user from Okta's
	// userinfo endpoint. It is a struct field so tests can stub it without a live issuer.
	fetchUserInfo func(ctx context.Context, rawToken string) (email string, groups []string, err error)
}

// accessTokenClaims are the claims the portal reads from the forwarded Okta access token.
// uid is the Okta user id (the "00u" value); cid is the OAuth client the token was issued to.
type accessTokenClaims struct {
	UID string `json:"uid"`
	CID string `json:"cid"`
}

// NewIdentityResolver builds a resolver that trusts the Envoy gateway OIDC proxy. Following the
// convention argus uses, it creates an OIDC provider from the issuer at boot (which discovers
// the JWKS endpoint) and verifies every forwarded token's signature against those keys. Groups
// come from Okta userinfo.
//
// Config:
//   - issuerURL: the Okta issuer (e.g. https://czi.okta.com). Its JWKS signs the tokens.
//   - clientID: the OAuth client the gateway authenticates with. The forwarded access token's
//     cid claim must match it, which binds the token to our gateway app.
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

	// The forwarded token is an access token whose audience is the Okta org, not our client
	// id, so skip the audience check here and bind to the client via the cid claim below.
	// Signature, issuer, and expiry are still verified, which stops a request that bypassed
	// the gateway from spoofing a user.
	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	return &IdentityResolver{
		devSub:      os.Getenv("PORTAL_DEV_SUB"),
		devEmail:    os.Getenv("PORTAL_DEV_EMAIL"),
		adminGroups: parseAdminGroups(os.Getenv("PORTAL_ADMIN_GROUPS")),
		verifyToken: func(ctx context.Context, raw string) (string, error) {
			token, err := verifier.Verify(ctx, raw)
			if err != nil {
				return "", fmt.Errorf("verifying access token: %w", err)
			}
			claims := accessTokenClaims{}
			err = token.Claims(&claims)
			if err != nil {
				return "", fmt.Errorf("reading token claims: %w", err)
			}
			return userIDFromClaims(claims, clientID)
		},
		fetchUserInfo: func(ctx context.Context, raw string) (string, []string, error) {
			info, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: raw}))
			if err != nil {
				return "", nil, fmt.Errorf("calling userinfo: %w", err)
			}
			var claims struct {
				Email  string   `json:"email"`
				Groups []string `json:"groups"`
			}
			err = info.Claims(&claims)
			if err != nil {
				return "", nil, fmt.Errorf("reading userinfo claims: %w", err)
			}
			return claims.Email, claims.Groups, nil
		},
	}, nil
}

// Resolve returns the current user, or errNoIdentity if none can be determined.
func (ir *IdentityResolver) Resolve(ctx context.Context, r *http.Request) (*User, error) {
	if ir.devSub != "" {
		return &User{Sub: ir.devSub, Email: ir.devEmail, Admin: true}, nil
	}

	raw := stripBearer(r.Header.Get("Authorization"))
	if raw == "" {
		return nil, errNoIdentity
	}

	userID, err := ir.verifyToken(ctx, raw)
	if err != nil {
		return nil, err
	}
	if userID == "" {
		return nil, errNoIdentity
	}
	user := &User{Sub: userID}

	// The verified token establishes who the user is. Email and groups come from userinfo; if
	// that call fails we still know the user, so degrade to a non-admin view rather than
	// locking them out.
	email, groups, err := ir.fetchUserInfo(ctx, raw)
	if err != nil {
		slog.Warn("fetching userinfo for portal user", "sub", userID, "error", err)
		return user, nil
	}
	user.Email = email
	user.Admin = isAdmin(groups, ir.adminGroups)
	return user, nil
}

// userIDFromClaims binds the access token to our client via the cid claim and returns the
// Okta user id. Rejecting a mismatched cid stops a valid token minted for a different Okta app
// from being replayed to the portal.
func userIDFromClaims(claims accessTokenClaims, clientID string) (string, error) {
	if claims.CID != clientID {
		return "", fmt.Errorf("token client id %q does not match expected %q", claims.CID, clientID)
	}
	if claims.UID == "" {
		return "", errors.New("token has no uid claim")
	}
	return claims.UID, nil
}

func isAdmin(groups []string, adminGroups map[string]bool) bool {
	for _, g := range groups {
		if adminGroups[strings.TrimSpace(g)] {
			return true
		}
	}
	return false
}

// stripBearer removes an optional "Bearer " prefix from an Authorization header value.
func stripBearer(header string) string {
	const prefix = "Bearer "
	if len(header) >= len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return strings.TrimSpace(header)
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
