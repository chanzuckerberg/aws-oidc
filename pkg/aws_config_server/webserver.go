package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/honeycombio/beeline-go/wrappers/hnyhttprouter"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
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
	OIDCProvider   string
	AWSWorkerRole  string
	AWSMasterRoles []string
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
		AddBeelineFields(r.Context(), BeelineField{
			Key:   "authHeader",
			Value: authHeader,
		})
		ctx := r.Context()
		if len(authHeader) <= 0 {
			logrus.Debugf("error: No Authorization header found.")
			http.Error(w, fmt.Sprintf("%v:%s", 407, http.StatusText(407)), 407)
			return
		}
		rawIDToken := stripBearerPrefixFromTokenString(authHeader)

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			logrus.Warnf("error: Unable to verify idToken. %s", err)
			http.Error(w, fmt.Sprintf("%v:%s", 401, http.StatusText(401)), 401)
			return
		}
		AddBeelineFields(ctx, BeelineField{
			Key:   "verifiedToken",
			Value: fmt.Sprintf("%v", idToken),
		})

		claims := &claims{}
		err = idToken.Claims(claims)
		if err != nil {
			logrus.Errorf("error: Unable to parse email from id token. %s", err)
			http.Error(w, fmt.Sprintf("%v:%s", 400, http.StatusText(400)), 400)
			return
		}
		AddBeelineFields(ctx, BeelineField{
			Key:   "claims",
			Value: fmt.Sprintf("%v", claims),
		})
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
	cachedClientIDtoProfiles *CachedGetClientIDToProfiles,
	oktaClient okta.AppResource,
) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()

		email := getEmailFromCtx(ctx)
		if email == nil {
			logrus.Error("Unable to get email")
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
		AddBeelineFields(ctx, BeelineField{
			Key:   "email",
			Value: *email,
		})

		sub := getSubFromCtx(ctx)
		if sub == nil {
			logrus.Errorf("Unable to get subject ID for %s", *email)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
		AddBeelineFields(ctx, BeelineField{
			Key:   "subjectID",
			Value: *sub,
		})

		clientIDs, err := okta.GetClientIDs(ctx, *sub, oktaClient)
		if err != nil {
			logrus.Errorf("Unable to get list of ClientIDs for %s: %s", *email, err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		AddBeelineFields(ctx, BeelineField{
			Key:   "clientIDs",
			Value: fmt.Sprintf("%v", clientIDs),
		})
		logrus.Debugf("%s's clientIDs: %s", *email, clientIDs)

		clientMapping, err := cachedClientIDtoProfiles.Get(ctx)
		if err != nil {
			logrus.Errorf("error: Unable to create mapping from clientID to roleARNs: %s", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
		AddBeelineFields(ctx, BeelineField{
			Key:   "clientMapping",
			Value: fmt.Sprintf("%v", clientMapping),
		})
		logrus.Debugf("%s's client mapping: %s", *email, clientMapping)

		awsConfig, err := createAWSConfig(ctx, awsGenerationParams, clientMapping, clientIDs)
		if err != nil {
			logrus.Errorf("error: unable to get AWS Config File: %s", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(awsConfig)
		if err != nil {
			logrus.Errorf("error: Unable to write config to http.ResponseWriter: %s", err)
			http.Error(w, fmt.Sprintf("%v:%s", 500, http.StatusText(500)), 500)
			return
		}
	}
}

type RouterConfig struct {
	Verifier              oidcVerifier
	AwsGenerationParams   *AWSConfigGenerationParams
	OktaAppClient         okta.AppResource
	GetClientIDToProfiles *CachedGetClientIDToProfiles
}

func GetRouter(
	ctx context.Context,
	config *RouterConfig,
) http.Handler {
	router := httprouter.New()
	handle := requireAuthentication(
		Index(
			config.AwsGenerationParams,
			config.GetClientIDToProfiles,
			config.OktaAppClient,
		),
		config.Verifier,
	)
	handle = hnyhttprouter.Middleware(handle)
	router.GET("/", handle)
	router.GET("/health", Health)

	handler := handlers.CombinedLoggingHandler(os.Stdout, router)
	handler = handlers.RecoveryHandler()(handler)
	return handler
}
