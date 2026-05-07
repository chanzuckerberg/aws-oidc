package workload_oidc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultMaxAttempts  = 3
	defaultInitialDelay = 500 * time.Millisecond
	defaultMaxDelay     = 5 * time.Second
)

// FetchToken performs an OAuth 2.0 client_credentials grant against the
// configured Custom Authorization Server and returns the issued access_token
// JWT. The token's Expiry is also returned for the caller to inform any
// downstream credential cache.
//
// FetchToken sends client credentials in the request body (client_secret_post,
// equivalent to oauth2.AuthStyleInParams). It does NOT send an `audience`
// parameter — Okta auto-stamps the `aud` claim from its Custom AS
// configuration.
//
// FetchToken retries up to defaultMaxAttempts times with exponential backoff
// on transient failures (HTTP 5xx, network errors). 4xx responses are treated
// as permanent configuration errors and not retried.
func (c Config) FetchToken(ctx context.Context) (jwt string, expiresAt time.Time, err error) {
	if err := c.validate(); err != nil {
		return "", time.Time{}, err
	}

	tokenURL := strings.TrimRight(c.IssuerURL, "/") + "/v1/token"

	cfg := &clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       c.Scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	var token *oauth2.Token
	err = retryWithBackoff(ctx, defaultMaxAttempts, func(ctx context.Context) error {
		tok, tokErr := cfg.Token(ctx)
		if tokErr != nil {
			return tokErr
		}
		token = tok
		return nil
	})
	if err != nil {
		return "", time.Time{}, classifyTokenError(err)
	}
	if token == nil || token.AccessToken == "" {
		return "", time.Time{}, ErrTokenMissing
	}

	return token.AccessToken, token.Expiry, nil
}

func (c Config) validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("workload_oidc: %w: ClientID is empty", ErrInvalidClient)
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("workload_oidc: %w: ClientSecret is empty", ErrInvalidClient)
	}
	if c.IssuerURL == "" {
		return fmt.Errorf("workload_oidc: %w: IssuerURL is empty", ErrIssuerUnreachable)
	}
	if !strings.HasPrefix(c.IssuerURL, "https://") && !strings.HasPrefix(c.IssuerURL, "http://") {
		return fmt.Errorf("workload_oidc: %w: IssuerURL must start with http:// or https://", ErrIssuerUnreachable)
	}
	return nil
}

// classifyTokenError converts oauth2 errors into the package's sentinel errors
// where appropriate. 401/403 from the issuer becomes ErrInvalidClient.
func classifyTokenError(err error) error {
	var oauthErr *oauth2.RetrieveError
	if errors.As(err, &oauthErr) && oauthErr.Response != nil {
		switch oauthErr.Response.StatusCode {
		case 401, 403:
			return fmt.Errorf("workload_oidc: %w: %v", ErrInvalidClient, err)
		}
	}
	return fmt.Errorf("workload_oidc: token request failed: %w", err)
}

// retryWithBackoff retries a function up to maxAttempts times with exponential
// backoff and jitter. Only transient errors trigger a retry — see
// isTransientError for the classification.
func retryWithBackoff(ctx context.Context, maxAttempts int, fn func(ctx context.Context) error) error {
	var lastErr error
	delay := defaultInitialDelay
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if !isTransientError(err) {
			return err
		}
		if attempt == maxAttempts {
			break
		}

		// Jittered exponential backoff: between 50% and 100% of the delay.
		jittered := delay/2 + time.Duration(rand.Int64N(int64(delay/2)+1))
		slog.Debug("workload_oidc: transient error, retrying",
			slog.Int("attempt", attempt),
			slog.Duration("backoff", jittered),
			slog.String("error", err.Error()),
		)
		select {
		case <-time.After(jittered):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay *= 2
		if delay > defaultMaxDelay {
			delay = defaultMaxDelay
		}
	}
	return fmt.Errorf("retry exhausted after %d attempts: %w", maxAttempts, lastErr)
}

// isTransientError returns true for errors that are likely to succeed on
// retry: HTTP 5xx from the issuer, and network-layer errors. HTTP 4xx
// responses are treated as permanent configuration errors and not retried.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var oauthErr *oauth2.RetrieveError
	if errors.As(err, &oauthErr) {
		if oauthErr.Response != nil {
			return oauthErr.Response.StatusCode >= 500
		}
		// No HTTP response attached: assume transport-level failure → retryable.
		return true
	}
	// Anything else (DNS, TLS, connection refused, EOF) is transient.
	return true
}
