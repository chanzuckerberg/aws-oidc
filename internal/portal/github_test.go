package portal

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testAppKeyPEM(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

// A configured client suggests and validates against the union of every installation's
// repositories, matches case-insensitively, and caches the set so a burst of keystrokes mints
// each installation token only once.
func TestGitHubAppSuggestAndReachable(t *testing.T) {
	ctx := context.Background()

	reposByToken := map[string][]string{
		"tok-111": {"chanzuckerberg/aws-oidc", "chanzuckerberg/shared-infra"},
		"tok-222": {"evolutionaryscale/foo"},
	}
	var tokenCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("POST /app/installations/{id}/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token":      "tok-" + r.PathValue("id"),
			"expires_at": time.Now().Add(time.Hour).UTC(),
		})
	})
	mux.HandleFunc("GET /installation/repositories", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "token ")
		var payload struct {
			Repositories []map[string]string `json:"repositories"`
		}
		for _, full := range reposByToken[token] {
			payload.Repositories = append(payload.Repositories, map[string]string{"full_name": full})
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	app, err := NewGitHubApp("4783816", testAppKeyPEM(t), "111", "evolutionaryscale=222", server.URL)
	require.NoError(t, err)
	require.NotNil(t, app)

	got, err := app.Suggest(ctx, "aws", 20)
	require.NoError(t, err)
	require.Contains(t, got, "chanzuckerberg/aws-oidc")

	got, err = app.Suggest(ctx, "foo", 20)
	require.NoError(t, err)
	require.Equal(t, []string{"evolutionaryscale/foo"}, got)

	reachable, err := app.Reachable(ctx, "Chanzuckerberg/AWS-OIDC")
	require.NoError(t, err)
	require.True(t, reachable, "reachability is case-insensitive")

	reachable, err = app.Reachable(ctx, "someorg/private")
	require.NoError(t, err)
	require.False(t, reachable)

	require.Equal(t, int32(2), atomic.LoadInt32(&tokenCalls), "the accessible set is cached across calls")
}

func TestNewGitHubAppNoCredentials(t *testing.T) {
	app, err := NewGitHubApp("", nil, "111", "", "")
	require.NoError(t, err)
	require.Nil(t, app, "no credentials leaves the feature off")
}
