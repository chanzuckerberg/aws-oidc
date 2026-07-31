package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
)

type oidcVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type claims struct {
	Email   string `json:"email"`
	Subject string `json:"sub"`
}

type AWSConfigGenerationParams struct {
	OIDCProvider string
	Concurrency  int
}

type AuthMiddleware struct {
	handler  http.Handler
	verifier oidcVerifier
}

func NewAuthMiddleware(handler http.Handler, verifier oidcVerifier) *AuthMiddleware {
	return &AuthMiddleware{
		handler:  handler,
		verifier: verifier,
	}
}

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	ctx := r.Context()
	if len(authHeader) <= 0 {
		slog.Debug(`no "Authorization" header found`)
		http.Error(w, fmt.Sprintf("%v:%s", 407, http.StatusText(407)), 407)
		return
	}
	rawIDToken := identity.StripBearer(authHeader)

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		slog.Error("verifying idToken", "error", err)
		http.Error(w, fmt.Sprintf("%v:%s", 401, http.StatusText(401)), 401)
		return
	}

	claims := &claims{}
	err = idToken.Claims(claims)
	if err != nil {
		slog.Error("parsing email from id token", "error", err)
		http.Error(w, fmt.Sprintf("%v:%s", 400, http.StatusText(400)), 400)
		return
	}
	user := &identity.User{Sub: claims.Subject, Email: claims.Email}
	rWithValues := r.WithContext(identity.NewContext(r.Context(), user))

	a.handler.ServeHTTP(w, rWithValues)
}

func getEmailFromCtx(ctx context.Context) *string {
	user := identity.FromContext(ctx)
	if user == nil {
		return nil
	}
	return &user.Email
}

func getSubFromCtx(ctx context.Context) *string {
	user := identity.FromContext(ctx)
	if user == nil {
		return nil
	}
	return &user.Sub
}

func Index(
	awsGenerationParams *AWSConfigGenerationParams,
	oktaClient okta.AppLister,
	mappingsProvider awsaccess.MappingsProvider,
	agentsProvider AgentsProvider,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		email := getEmailFromCtx(ctx)
		if email == nil {
			slog.Error("no email in context")
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		sub := getSubFromCtx(ctx)
		if sub == nil {
			slog.Error(fmt.Sprintf("getting subject ID for %s", *email))
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
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

		access, err := awsaccess.Resolve(ctx, *sub, oktaClient, clientMappingsByKey)
		if err != nil {
			slog.Error(fmt.Sprintf("resolving access for sub %s", *sub), "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		slog.Debug("creating aws config", "email", *email, "accountsLen", len(access.Accounts))
		awsConfig := createAWSConfig(awsGenerationParams.OIDCProvider, access)

		// Agent profiles are additive and optional. A nil provider (the feature disabled, as
		// in prod today) leaves the response with only the human profiles.
		if agentsProvider != nil {
			agents, agentsErr := agentsProvider(ctx, *sub)
			if agentsErr != nil {
				// Degrade rather than fail: a registry hiccup must not break human config,
				// which is the primary purpose of this endpoint.
				slog.Error("building agent configs", "sub", *sub, "error", agentsErr)
			} else {
				awsConfig.Agents = agents
			}
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

// AgentsProvider returns the caller's agent configs, keyed on the caller's Okta subject. It
// is optional: a nil provider disables agent config generation and the response carries only
// the human profiles. It is a seam so this package stays free of the Kubernetes client that
// reads the Agent custom resources.
type AgentsProvider func(ctx context.Context, ownerSub string) ([]AgentConfig, error)

type RouterConfig struct {
	Verifier            oidcVerifier
	AwsGenerationParams *AWSConfigGenerationParams
	OktaAppClient       okta.AppLister
	MappingsProvider    awsaccess.MappingsProvider
	AgentsProvider      AgentsProvider
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
		config.AgentsProvider,
	), config.Verifier)

	mux.Handle("/", handle)
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
