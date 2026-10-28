package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/gorilla/handlers"
)

type AWSConfigGenerationParams struct {
	OIDCProvider string
	Concurrency  int
}

// From https://github.com/dgrijalva/jwt-go/blob/master/request/oauth2.go
// Strips 'Bearer ' prefix from bearer token string
func stripBearerPrefixFromTokenString(token string) string {
	// Should be a bearer token
	if len(token) > 6 && strings.ToUpper(token[0:7]) == "BEARER " {
		return token[7:]
	}
	return token
}

func Index(
	awsGenerationParams *AWSConfigGenerationParams,
	oktaClient okta.AppLister,
	mappingsProvider MappingsProvider,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		principal, ok := principalFromContext(ctx)
		if !ok {
			slog.Error("no principal in context")
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
		if principal.Issuer != normalizeIssuer(awsGenerationParams.OIDCProvider) ||
			principal.Email == "" {
			http.Error(w, fmt.Sprintf("%v:%s", 403, http.StatusText(403)), 403)
			return
		}

		// Re-read the rolemap on every request so a freshly generated mapping is picked up
		// without restarting the server.
		clientMappingsByKey, err := mappingsProvider(ctx)
		if err != nil {
			slog.Error("reading rolemap", "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		clientIDs, err := okta.GetClientIDs(ctx, principal.Subject, oktaClient)
		if err != nil {
			slog.Error(fmt.Sprintf("getting list of ClientIDs for sub %s", principal.Subject), "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		slog.Debug("creating aws config", "email", principal.Email, "clientIDsLen", len(clientIDs))
		awsConfig, err := createAWSConfig(awsGenerationParams.OIDCProvider, clientMappingsByKey, clientIDs)
		if err != nil {
			slog.Error("getting AWS Config File", "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(awsConfig)
		if err != nil {
			slog.Error("writing config to http.ResponseWriter", "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
	})
}

// MappingsProvider returns the current rolemap, grouped by client ID. It is called once
// per request so the server always serves the latest rolemap.
type MappingsProvider func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error)

type RouterConfig struct {
	Verifier            IdentityVerifier
	AwsGenerationParams *AWSConfigGenerationParams
	OktaAppClient       okta.AppLister
	MappingsProvider    MappingsProvider
	EntitlementHandler  http.Handler
}

type SlogRecoveryLogger slog.Logger

func (l SlogRecoveryLogger) Println(v ...interface{}) {
	slog.Error(fmt.Sprintf("%v", v))
}

func GetRouter(
	ctx context.Context,
	config *RouterConfig,
) http.Handler {
	mux := http.NewServeMux()
	handle := NewAuthMiddleware(Index(
		config.AwsGenerationParams,
		config.OktaAppClient,
		config.MappingsProvider,
	), config.Verifier)

	mux.Handle("/", handle)
	if config.EntitlementHandler != nil {
		mux.Handle(
			"/api/v1/entitlements",
			NewAuthMiddleware(config.EntitlementHandler, config.Verifier),
		)
	}
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	logger := slog.Default()
	handler := handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(SlogRecoveryLogger(*logger)),
	)(mux)
	return handler
}
