// Package identity models the authenticated caller once, so the config server and the portal
// answer "who is this person" the same way even though they verify different tokens. The
// config server verifies an Okta ID token; the portal verifies a gateway-forwarded access
// token. Both resolve to a User keyed on the Okta subject, which is the same value the AWS
// trust policy conditions on and the portal stamps as an Agent's owner.
package identity

import (
	"context"
	"strings"
)

// User is the authenticated caller. Sub is the Okta user id (the "00u" value) that
// entitlements, agent ownership, and the AWS trust policy all key on.
type User struct {
	Sub    string
	Email  string
	Groups []string
	Admin  bool
}

type ctxKey struct{}

// NewContext returns a context carrying the user.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, ctxKey{}, u)
}

// FromContext returns the user carried by the context, or nil.
func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(ctxKey{}).(*User)
	return u
}

// StripBearer removes an optional, case-insensitive "Bearer " prefix from an Authorization
// header value and trims surrounding whitespace.
func StripBearer(header string) string {
	const prefix = "Bearer "
	if len(header) >= len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return strings.TrimSpace(header)
}
