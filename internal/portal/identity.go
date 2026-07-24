package portal

import (
	"errors"
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

// IdentityResolver extracts the current user from a request. Real Okta browser login is a
// follow-up; for now identity comes from a dev override (for rdev) or from trusted headers
// set by an auth proxy in front of the portal.
type IdentityResolver struct {
	devSub       string
	devEmail     string
	subHeader    string
	emailHeader  string
	groupsHeader string
	adminGroups  map[string]bool
}

// NewIdentityResolver reads its configuration from the environment:
//   - PORTAL_DEV_SUB / PORTAL_DEV_EMAIL: hardcode the current user (rdev/dev). When set,
//     the user is treated as admin so the whole flow can be exercised.
//   - PORTAL_SUB_HEADER / PORTAL_EMAIL_HEADER / PORTAL_GROUPS_HEADER: headers an auth proxy
//     injects. Defaults match oauth2-proxy's X-Auth-Request-* headers.
//   - PORTAL_ADMIN_GROUPS: comma-separated groups that grant the admin view.
func NewIdentityResolver() *IdentityResolver {
	return &IdentityResolver{
		devSub:       os.Getenv("PORTAL_DEV_SUB"),
		devEmail:     os.Getenv("PORTAL_DEV_EMAIL"),
		subHeader:    envOr("PORTAL_SUB_HEADER", "X-Auth-Request-User"),
		emailHeader:  envOr("PORTAL_EMAIL_HEADER", "X-Auth-Request-Email"),
		groupsHeader: envOr("PORTAL_GROUPS_HEADER", "X-Auth-Request-Groups"),
		adminGroups:  parseAdminGroups(os.Getenv("PORTAL_ADMIN_GROUPS")),
	}
}

// Resolve returns the current user, or errNoIdentity if none can be determined.
func (ir *IdentityResolver) Resolve(r *http.Request) (*User, error) {
	if ir.devSub != "" {
		return &User{Sub: ir.devSub, Email: ir.devEmail, Admin: true}, nil
	}

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
