package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
)

type oidcVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type contextKey int

const contextKeyEmail contextKey = 0
const contextKeySub contextKey = 1

type claims struct {
	Email   string `json:"email"`
	Subject string `json:"sub"`
}

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
	rawIDToken := stripBearerPrefixFromTokenString(authHeader)

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
	ctxWithValues := context.WithValue(r.Context(), contextKeyEmail, claims.Email)
	ctxWithValues = context.WithValue(ctxWithValues, contextKeySub, claims.Subject)
	rWithValues := r.WithContext(ctxWithValues)

	a.handler.ServeHTTP(w, rWithValues)
}

func getEmailFromCtx(ctx context.Context) *string {
	email, ok := ctx.Value(contextKeyEmail).(string)
	if !ok {
		return nil
	}
	return &email
}

func getSubFromCtx(ctx context.Context) *string {
	sub, ok := ctx.Value(contextKeySub).(string)
	if !ok {
		return nil
	}
	return &sub
}

func Index(
	awsGenerationParams *AWSConfigGenerationParams,
	oktaClient okta.AppResource,
	clientMappingsByKey okta.OIDCRoleMappingsByKey,
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

		clientIDs, err := okta.GetClientIDs(ctx, *sub, oktaClient)
		if err != nil {
			slog.Error(fmt.Sprintf("getting list of ClientIDs for %s", *email), "error", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		slog.Debug("creating aws config", "email", *email, "clientIDsLen", len(clientIDs))
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

type RouterConfig struct {
	Verifier            oidcVerifier
	AwsGenerationParams *AWSConfigGenerationParams
	OktaAppClient       okta.AppResource
	ClientMappings      okta.OIDCRoleMappingsByKey
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
		config.ClientMappings,
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
