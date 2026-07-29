package portal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// User is the authenticated portal user. Sub is the Okta subject (user id) and is what
// entitlements and agent ownership are keyed on.
type User struct {
	Sub   string
	Email string
	Admin bool
}

// errNoIdentity is returned when no authenticated user can be determined.
var errNoIdentity = errors.New("no authenticated user")

// IdentityResolver extracts the current user from a request. In prod the user comes from the
// Okta ID token that the Envoy gateway OIDC proxy forwards. A dev override and trusted-header
// path exist for local and rdev testing.
type IdentityResolver struct {
	devSub        string
	devEmail      string
	idTokenCookie string
	subHeader     string
	emailHeader   string
	groupsHeader  string
	adminGroups   map[string]bool
}

// NewIdentityResolver reads its configuration from the environment:
//   - PORTAL_DEV_SUB / PORTAL_DEV_EMAIL: act as a fixed user (treated as admin). Handy in rdev
//     to impersonate any user for testing without a real login.
//   - PORTAL_ID_TOKEN_COOKIE: name of the cookie the Envoy gateway OIDC proxy stores the Okta
//     ID token in. When set, identity comes from that token.
//   - PORTAL_SUB_HEADER / PORTAL_EMAIL_HEADER / PORTAL_GROUPS_HEADER: headers an auth proxy
//     injects. Defaults match oauth2-proxy's X-Auth-Request-* headers.
//   - PORTAL_ADMIN_GROUPS: comma-separated groups that grant the admin view.
func NewIdentityResolver() *IdentityResolver {
	return &IdentityResolver{
		devSub:        os.Getenv("PORTAL_DEV_SUB"),
		devEmail:      os.Getenv("PORTAL_DEV_EMAIL"),
		idTokenCookie: os.Getenv("PORTAL_ID_TOKEN_COOKIE"),
		subHeader:     envOr("PORTAL_SUB_HEADER", "X-Auth-Request-User"),
		emailHeader:   envOr("PORTAL_EMAIL_HEADER", "X-Auth-Request-Email"),
		groupsHeader:  envOr("PORTAL_GROUPS_HEADER", "X-Auth-Request-Groups"),
		adminGroups:   parseAdminGroups(os.Getenv("PORTAL_ADMIN_GROUPS")),
	}
}

// Resolve returns the current user, or errNoIdentity if none can be determined.
func (ir *IdentityResolver) Resolve(r *http.Request) (*User, error) {
	if ir.devSub != "" {
		return &User{Sub: ir.devSub, Email: ir.devEmail, Admin: true}, nil
	}

	if ir.idTokenCookie != "" {
		return ir.userFromIDToken(r)
	}

	return ir.userFromHeaders(r)
}

// userFromIDToken reads the Okta ID token from the cookie the gateway forwards. The gateway
// has already authenticated the session, so the token is decoded and trusted rather than
// re-verified (the same trust the header path places in the proxy).
func (ir *IdentityResolver) userFromIDToken(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(ir.idTokenCookie)
	if err != nil {
		return nil, errNoIdentity
	}
	claims, err := parseIDTokenClaims(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("parsing id token: %w", err)
	}
	if claims.Sub == "" {
		return nil, errNoIdentity
	}

	user := &User{Sub: claims.Sub, Email: claims.Email}
	for _, group := range claims.Groups {
		if ir.adminGroups[group] {
			user.Admin = true
			break
		}
	}
	return user, nil
}

func (ir *IdentityResolver) userFromHeaders(r *http.Request) (*User, error) {
	sub := strings.TrimSpace(r.Header.Get(ir.subHeader))
	if sub == "" {
		return nil, errNoIdentity
	}

	user := &User{Sub: sub, Email: strings.TrimSpace(r.Header.Get(ir.emailHeader))}
	for _, group := range strings.Split(r.Header.Get(ir.groupsHeader), ",") {
		if ir.adminGroups[strings.TrimSpace(group)] {
			user.Admin = true
			break
		}
	}
	return user, nil
}

type idTokenClaims struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

// parseIDTokenClaims decodes the claims from a JWT without verifying its signature.
func parseIDTokenClaims(token string) (*idTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("token is not a JWT")
	}
	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding payload: %w", err)
	}
	claims := &idTokenClaims{}
	err = json.Unmarshal(payload, claims)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling claims: %w", err)
	}
	return claims, nil
}

// decodeJWTSegment base64url-decodes a JWT segment, tolerating missing padding.
func decodeJWTSegment(seg string) ([]byte, error) {
	if m := len(seg) % 4; m != 0 {
		seg += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(seg)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
