package portal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDTokenClaimsTeamGroups(t *testing.T) {
	var claims idTokenClaims
	err := json.Unmarshal([]byte(`{"email":"user@example.com","teamGroups":["team-central-infra-eng"]}`), &claims)
	require.NoError(t, err)
	require.Equal(t, "user@example.com", claims.Email)
	require.Equal(t, []string{"team-central-infra-eng"}, claims.Groups)
}

func TestResolveDevOverride(t *testing.T) {
	ir := &IdentityResolver{devSub: "00udev", devEmail: "dev@example.com"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00udev", user.Sub)
	require.Equal(t, "dev@example.com", user.Email)
	require.True(t, user.Admin, "dev override should be admin")
	require.Equal(t, "Admin access granted by PORTAL_DEV_SUB", user.AdminReason)
}

func TestResolveIDToken(t *testing.T) {
	ir := &IdentityResolver{
		adminGroups: map[string]bool{"infra-eng": true},
		verifyIDToken: func(_ context.Context, raw string) (string, string, []string, error) {
			require.Equal(t, "idtok456", raw)
			return "00uid", "user@example.com", []string{"everyone", "infra-eng"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Id-Token", "idtok456")

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "00uid", user.Sub)
	require.Equal(t, "user@example.com", user.Email)
	require.True(t, user.Admin, "membership in an admin group should grant admin")
	require.Equal(t, "Admin through Okta group infra-eng", user.AdminReason)
}

func TestResolveIDTokenNonAdmin(t *testing.T) {
	ir := &IdentityResolver{
		adminGroups: map[string]bool{"infra-eng": true},
		verifyIDToken: func(_ context.Context, _ string) (string, string, []string, error) {
			return "00uid", "user@example.com", []string{"everyone"}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Id-Token", "idtok")

	user, err := ir.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.False(t, user.Admin)
	require.Empty(t, user.AdminReason)
}

func TestResolveInvalidIDToken(t *testing.T) {
	ir := &IdentityResolver{
		verifyIDToken: func(_ context.Context, _ string) (string, string, []string, error) {
			return "", "", nil, errors.New("bad id token")
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Id-Token", "bad")

	_, err := ir.Resolve(context.Background(), req)
	require.ErrorIs(t, err, errNoIdentity)
}

func TestResolveMissingAuthHeader(t *testing.T) {
	ir := &IdentityResolver{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := ir.Resolve(context.Background(), req)
	require.ErrorIs(t, err, errNoIdentity)
}

func TestDescribeTokenReadsClaimsWithoutVerifying(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(
		`{"iss":"https://czi.okta.com","cid":"client-b","uid":"00ureal","aud":"api://default","exp":1}`))
	attr := describeToken("header." + payload + ".signature")

	require.Equal(t, "unverified_token_claims", attr.Key)
	rendered := attr.Value.String()
	require.Contains(t, rendered, "client-b", "the cid must be visible to explain a client mismatch")
	require.Contains(t, rendered, "https://czi.okta.com")
	require.Contains(t, rendered, "expired=true")
}

func TestDescribeTokenOpaque(t *testing.T) {
	attr := describeToken("not-a-jwt")
	require.Equal(t, "token", attr.Key)
	require.Contains(t, attr.Value.String(), "opaque")
}

func TestResolveMissingHeaderNamesTheHeadersPresent(t *testing.T) {
	ir := &IdentityResolver{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	_, err := ir.Resolve(context.Background(), req)
	require.ErrorIs(t, err, errNoIdentity)
	require.Contains(t, err.Error(), "X-Forwarded-For", "the log must show what the gateway did send")
}
