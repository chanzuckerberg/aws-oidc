// Package workload_oidc implements an OAuth 2.0 client_credentials grant
// against an Okta Custom Authorization Server, intended for unattended
// (non-human) workloads. The resulting access_token JWT can be exchanged for
// AWS STS credentials via sts:AssumeRoleWithWebIdentity.
//
// IMPORTANT: the IssuerURL on Config MUST point to an Okta Custom Authorization
// Server (e.g. https://example.okta.com/oauth2/<auth_server_id>). The Okta Org
// (default) Authorization Server at https://example.okta.com/oauth2/v1 issues
// opaque access tokens whose structure is "subject to change at any time
// without notice" per Okta documentation, and AWS STS cannot validate them.
package workload_oidc

import (
	"errors"
	"fmt"
	"log/slog"
)

// Sentinel errors used by the package. Wrap with fmt.Errorf("...: %w", err).
var (
	// ErrInvalidClient indicates that the configured client credentials were
	// rejected by the issuer (HTTP 401 / invalid_client). Generally not
	// transient — fix the configuration.
	ErrInvalidClient = errors.New("invalid client credentials")

	// ErrIssuerUnreachable indicates the OIDC issuer URL is missing,
	// malformed, or unreachable.
	ErrIssuerUnreachable = errors.New("issuer URL unreachable or invalid")

	// ErrTokenMissing indicates the issuer responded successfully but did not
	// include an access_token.
	ErrTokenMissing = errors.New("issuer response did not contain access_token")
)

// Config holds OAuth 2.0 client_credentials configuration for a workload
// OIDC flow.
//
// Audience is intentionally NOT a field on Config. Okta Custom Authorization
// Servers stamp the `aud` claim from their own configuration and ignore any
// audience parameter sent on the /v1/token request — that's an Auth0
// convention, not an Okta one. The audience must be configured at the Okta
// Custom AS level and matched in the AWS IAM OIDC identity provider's
// client_id_list and the IAM role's trust policy.
type Config struct {
	// ClientID is the Okta Service app's client_id.
	ClientID string

	// ClientSecret is the Okta Service app's client secret. Treat as sensitive;
	// the String() and LogValue() methods on Config redact this field.
	ClientSecret string

	// IssuerURL is the Okta Custom Authorization Server URL, e.g.
	// "https://example.okta.com/oauth2/aus123abc". The /v1/token suffix is
	// appended internally. Org / default authorization servers MUST NOT be
	// used here.
	IssuerURL string

	// Scopes are the OAuth scopes to request. Typically a custom scope
	// configured on the Okta Custom AS (e.g. "aws-m2m-access"). May be empty.
	Scopes []string
}

// String returns a redacted representation of the Config suitable for logs and
// error messages. The ClientSecret is replaced with "[REDACTED]".
func (c Config) String() string {
	return fmt.Sprintf(
		"workload_oidc.Config{ClientID:%q IssuerURL:%q Scopes:%v ClientSecret:[REDACTED]}",
		c.ClientID, c.IssuerURL, c.Scopes,
	)
}

// LogValue implements slog.LogValuer so structured logs emit a redacted
// representation. The ClientSecret never appears in any slog output.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("client_id", c.ClientID),
		slog.String("issuer_url", c.IssuerURL),
		slog.Any("scopes", c.Scopes),
		slog.String("client_secret", "[REDACTED]"),
	)
}
