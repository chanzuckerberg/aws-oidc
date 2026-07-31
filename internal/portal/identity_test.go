package portal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserIDFromClaims(t *testing.T) {
	uid, err := userIDFromClaims(accessTokenClaims{UID: "00ureal", CID: "client-a"}, "client-a")
	require.NoError(t, err)
	require.Equal(t, "00ureal", uid)
}

func TestUserIDFromClaimsWrongClient(t *testing.T) {
	_, err := userIDFromClaims(accessTokenClaims{UID: "00ureal", CID: "client-b"}, "client-a")
	require.Error(t, err, "a token minted for another client must be rejected")
}

func TestUserIDFromClaimsMissingUID(t *testing.T) {
	_, err := userIDFromClaims(accessTokenClaims{CID: "client-a"}, "client-a")
	require.Error(t, err)
}

func TestResolveDevOverride(t *testing.T) {
	ir := &IdentityResolver{devSub: "00udev", devEmail: "dev@example.com"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00udev", user.Sub)
	require.Equal(t, "dev@example.com", user.Email)
	require.True(t, user.Admin, "dev override should be admin")
}

func TestResolveAccessToken(t *testing.T) {
	ir := &IdentityResolver{
		adminGroups: map[string]bool{"aws-oidc-admins": true},
		verifyToken: func(_ context.Context, raw string) (string, error) {
			require.Equal(t, "tok123", raw)
			return "00ureal", nil
		},
		fetchUserInfo: func(_ context.Context, _ string) (string, []string, error) {
			return "real@example.com", []string{"everyone", "aws-oidc-admins"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok123")

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00ureal", user.Sub, "user id comes from the token's uid claim")
	require.Equal(t, "real@example.com", user.Email)
	require.True(t, user.Admin, "membership in an admin group should grant admin")
}

func TestResolveAccessTokenNonAdmin(t *testing.T) {
	ir := &IdentityResolver{
		adminGroups: map[string]bool{"aws-oidc-admins": true},
		verifyToken: func(_ context.Context, _ string) (string, error) { return "00ureal", nil },
		fetchUserInfo: func(_ context.Context, _ string) (string, []string, error) {
			return "real@example.com", []string{"everyone"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00ureal", user.Sub)
	require.False(t, user.Admin)
}

func TestResolveMissingAuthHeader(t *testing.T) {
	ir := &IdentityResolver{
		verifyToken: func(_ context.Context, _ string) (string, error) {
			t.Fatal("verifyToken should not be called without an Authorization header")
			return "", nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := ir.Resolve(context.Background(), req)
	require.ErrorIs(t, err, errNoIdentity)
}

func TestResolveVerifyFails(t *testing.T) {
	ir := &IdentityResolver{
		verifyToken: func(_ context.Context, _ string) (string, error) {
			return "", errors.New("bad signature")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")

	_, err := ir.Resolve(context.Background(), req)
	require.Error(t, err)
}

// A userinfo failure must not lock the user out: identity is already established from the
// verified token, so the request degrades to a non-admin view.
func TestResolveUserInfoFailureDegradesToNonAdmin(t *testing.T) {
	ir := &IdentityResolver{
		adminGroups: map[string]bool{"aws-oidc-admins": true},
		verifyToken: func(_ context.Context, _ string) (string, error) { return "00ureal", nil },
		fetchUserInfo: func(_ context.Context, _ string) (string, []string, error) {
			return "", nil, errors.New("userinfo unavailable")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00ureal", user.Sub)
	require.Empty(t, user.Email)
	require.False(t, user.Admin)
}
