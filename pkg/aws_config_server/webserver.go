package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/wrappers/hnyhttprouter"
	"github.com/julienschmidt/httprouter"
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

func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
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

func requireAuthentication(next httprouter.Handle, verifier oidcVerifier) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authHeader := r.Header.Get("Authorization")
		ctx := r.Context()
		if len(authHeader) <= 0 {
			slog.Debug(`no "Authorization" header found`)
			http.Error(w, fmt.Sprintf("%v:%s", 407, http.StatusText(407)), 407)
			return
		}
		rawIDToken := stripBearerPrefixFromTokenString(authHeader)

		idToken, err := verifier.Verify(ctx, rawIDToken)
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

		next(w, rWithValues, ps)
	}
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
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		email := getEmailFromCtx(ctx)
		if email == nil {
			slog.Error("no email in context")
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
		beeline.AddField(ctx, "email", *email)

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

		slog.Debug(fmt.Sprintf("%s's clientIDs: %s", *email, clientIDs))

		clientMapping := okta.FromContext(ctx)
		slog.Debug(fmt.Sprintf("%s's client mapping: %#v", *email, clientMapping))

		awsConfig, err := createAWSConfig(awsGenerationParams.OIDCProvider, clientMapping)
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
	}
}

type RouterConfig struct {
	Verifier            oidcVerifier
	AwsGenerationParams *AWSConfigGenerationParams
	OktaAppClient       okta.AppResource
}

func GetRouter(
	ctx context.Context,
	config *RouterConfig,
) http.Handler {
	router := httprouter.New()
	handle := requireAuthentication(
		Index(
			config.AwsGenerationParams,
			config.OktaAppClient,
		),
		config.Verifier,
	)
	handle = hnyhttprouter.Middleware(handle)
	router.GET("/", handle)
	router.GET("/health", Health)

	handler := handlers.CombinedLoggingHandler(os.Stdout, router)
	handler = handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
	)(handler)
	return handler
}
