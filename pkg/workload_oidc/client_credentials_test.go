package workload_oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope,omitempty"`
}

func TestFetchToken_HappyPath(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.Equal("/v1/token", req.URL.Path)
		r.Equal(http.MethodPost, req.Method)
		r.NoError(req.ParseForm())
		r.Equal("client_credentials", req.Form.Get("grant_type"))
		// AuthStyleInParams puts client_id and client_secret in the body.
		r.Equal("test-client", req.Form.Get("client_id"))
		r.Equal("test-secret", req.Form.Get("client_secret"))
		r.Equal("aws-m2m-access", req.Form.Get("scope"))
		// We must NOT send an audience parameter — that's an Auth0-ism.
		r.Empty(req.Form.Get("audience"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockTokenResponse{
			AccessToken: "fake.jwt.token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
			Scope:       "aws-m2m-access",
		})
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
		Scopes:       []string{"aws-m2m-access"},
	}

	jwt, expiresAt, err := cfg.FetchToken(context.Background())
	r.NoError(err)
	r.Equal("fake.jwt.token", jwt)
	r.WithinDuration(time.Now().Add(time.Hour), expiresAt, 5*time.Second)
}

func TestFetchToken_BadCredentials_NoRetry(t *testing.T) {
	r := require.New(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"error":"invalid_client","error_description":"Client authentication failed."}`)
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "bad",
		ClientSecret: "wrong",
		IssuerURL:    server.URL,
	}

	_, _, err := cfg.FetchToken(context.Background())
	r.Error(err)
	r.True(errors.Is(err, ErrInvalidClient), "expected ErrInvalidClient, got %v", err)
	// 4xx must NOT trigger retries — exactly one call.
	r.Equal(int32(1), calls.Load(), "401 should not be retried")
}

func TestFetchToken_TransientThenSuccess_Retries(t *testing.T) {
	r := require.New(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"error":"server_error"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockTokenResponse{
			AccessToken: "after.retry.jwt",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
	}

	jwt, _, err := cfg.FetchToken(context.Background())
	r.NoError(err)
	r.Equal("after.retry.jwt", jwt)
	r.Equal(int32(3), calls.Load())
}

func TestFetchToken_Persistent5xx_RetriesExhausted(t *testing.T) {
	r := require.New(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
	}

	_, _, err := cfg.FetchToken(context.Background())
	r.Error(err)
	r.Equal(int32(defaultMaxAttempts), calls.Load(), "should attempt all retries on persistent 5xx")
}

func TestFetchToken_NetworkError_Retries(t *testing.T) {
	r := require.New(t)

	// Bind a port, close immediately so connect refuses.
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    url,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, _, err := cfg.FetchToken(ctx)
	r.Error(err, "expected error from refused connection")
}

func TestFetchToken_MalformedJSON(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{not-json}`)
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
	}

	_, _, err := cfg.FetchToken(context.Background())
	r.Error(err)
}

func TestFetchToken_MissingAccessToken(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 200 with empty body / no access_token field.
		_, _ = fmt.Fprint(w, `{"token_type":"Bearer","expires_in":3600}`)
	}))
	defer server.Close()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
	}

	_, _, err := cfg.FetchToken(context.Background())
	r.Error(err)
}

func TestFetchToken_ContextCanceled_NoRetry(t *testing.T) {
	r := require.New(t)

	var calls atomic.Int32
	cleanup := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		// Block until either the client disconnects or test cleanup signals.
		// We use cleanup so server.Close() never blocks even if the client
		// disconnect doesn't promptly cancel req.Context().
		select {
		case <-req.Context().Done():
		case <-cleanup:
		}
	}))
	defer func() {
		close(cleanup)
		server.Close()
	}()

	cfg := Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    server.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _, err := cfg.FetchToken(ctx)
	r.Error(err)
	// Context cancellation must not retry.
	r.LessOrEqual(calls.Load(), int32(1), "context cancellation should not retry")
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		cfg     Config
		wantErr error
	}{
		"missing client_id": {
			cfg:     Config{ClientSecret: "x", IssuerURL: "https://example.okta.com/oauth2/aus123"},
			wantErr: ErrInvalidClient,
		},
		"missing client_secret": {
			cfg:     Config{ClientID: "x", IssuerURL: "https://example.okta.com/oauth2/aus123"},
			wantErr: ErrInvalidClient,
		},
		"missing issuer_url": {
			cfg:     Config{ClientID: "x", ClientSecret: "x"},
			wantErr: ErrIssuerUnreachable,
		},
		"non-http issuer_url": {
			cfg:     Config{ClientID: "x", ClientSecret: "x", IssuerURL: "ftp://example.com"},
			wantErr: ErrIssuerUnreachable,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			err := tc.cfg.validate()
			r.Error(err)
			r.True(errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)
		})
	}
}

func TestConfig_RedactsSecret(t *testing.T) {
	r := require.New(t)

	cfg := Config{
		ClientID:     "client-id",
		ClientSecret: "super-secret-value",
		IssuerURL:    "https://example.okta.com/oauth2/aus123",
		Scopes:       []string{"foo"},
	}

	// fmt.Stringer
	got := cfg.String()
	r.NotContains(got, "super-secret-value")
	r.Contains(got, "[REDACTED]")
	r.Contains(got, "client-id")

	// %v / %+v formatters use Stringer
	r.NotContains(fmt.Sprintf("%v", cfg), "super-secret-value")
	r.NotContains(fmt.Sprintf("%+v", cfg), "super-secret-value")

	// slog LogValue
	v := cfg.LogValue()
	rendered := v.String()
	r.NotContains(rendered, "super-secret-value")
	r.Contains(rendered, "[REDACTED]")
}

func TestFetchToken_TrimsTrailingSlashOnIssuerURL(t *testing.T) {
	r := require.New(t)

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		receivedPath = req.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockTokenResponse{
			AccessToken: "x",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer server.Close()

	// Trailing slash on the issuer URL should not produce a //v1/token URL.
	cfg := Config{
		ClientID:     "x",
		ClientSecret: "x",
		IssuerURL:    server.URL + "/",
	}
	_, _, err := cfg.FetchToken(context.Background())
	r.NoError(err)
	r.Equal("/v1/token", receivedPath)
	r.False(strings.Contains(receivedPath, "//"), "should not produce double-slash path")
}
