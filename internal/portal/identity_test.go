package portal

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeJWT builds an unsigned JWT with the given claims payload. The signature is a placeholder
// because the portal trusts the gateway and does not verify it.
func fakeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	seg := base64.RawURLEncoding.EncodeToString(payload)
	return "header." + seg + ".sig"
}

func TestResolveDevOverride(t *testing.T) {
	ir := &IdentityResolver{devSub: "00udev", devEmail: "dev@example.com"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	user, err := ir.Resolve(req)
	require.NoError(t, err)
	require.Equal(t, "00udev", user.Sub)
	require.Equal(t, "dev@example.com", user.Email)
	require.True(t, user.Admin, "dev override should be admin")
}

func TestResolveIDTokenCookie(t *testing.T) {
	ir := &IdentityResolver{
		idTokenCookie: "IdToken-portal",
		adminGroups:   map[string]bool{"aws-oidc-admins": true},
	}

	token := fakeJWT(t, map[string]any{
		"sub":    "00ureal",
		"email":  "real@example.com",
		"groups": []string{"everyone", "aws-oidc-admins"},
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "IdToken-portal", Value: token})

	user, err := ir.Resolve(req)
	require.NoError(t, err)
	require.Equal(t, "00ureal", user.Sub)
	require.Equal(t, "real@example.com", user.Email)
	require.True(t, user.Admin, "membership in an admin group should grant admin")
}

func TestResolveIDTokenCookieNonAdmin(t *testing.T) {
	ir := &IdentityResolver{
		idTokenCookie: "IdToken-portal",
		adminGroups:   map[string]bool{"aws-oidc-admins": true},
	}

	token := fakeJWT(t, map[string]any{"sub": "00ureal", "groups": []string{"everyone"}})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "IdToken-portal", Value: token})

	user, err := ir.Resolve(req)
	require.NoError(t, err)
	require.Equal(t, "00ureal", user.Sub)
	require.False(t, user.Admin)
}

func TestResolveIDTokenCookieMissing(t *testing.T) {
	ir := &IdentityResolver{idTokenCookie: "IdToken-portal"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := ir.Resolve(req)
	require.ErrorIs(t, err, errNoIdentity)
}

func TestResolveHeaders(t *testing.T) {
	ir := &IdentityResolver{
		subHeader:    "X-Auth-Request-User",
		emailHeader:  "X-Auth-Request-Email",
		groupsHeader: "X-Auth-Request-Groups",
		adminGroups:  map[string]bool{"admins": true},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Auth-Request-User", "00uheader")
	req.Header.Set("X-Auth-Request-Email", "h@example.com")
	req.Header.Set("X-Auth-Request-Groups", "everyone, admins")

	user, err := ir.Resolve(req)
	require.NoError(t, err)
	require.Equal(t, "00uheader", user.Sub)
	require.Equal(t, "h@example.com", user.Email)
	require.True(t, user.Admin)
}
