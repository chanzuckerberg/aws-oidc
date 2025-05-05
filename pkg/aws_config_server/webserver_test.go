package aws_config_server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	oidc "github.com/coreos/go-oidc"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/stretchr/testify/require"
)

type idTokenVerifier struct {
	expectedIDToken string
}

func (idtv *idTokenVerifier) Verify(ctx context.Context, idToken string) (*oidc.IDToken, error) {
	if idtv.expectedIDToken != idToken {
		return nil, fmt.Errorf("id tokens do not match!")
	}
	return &oidc.IDToken{}, nil
}

var testAWSConfigGenerationParams = AWSConfigGenerationParams{
	OIDCProvider: "validProvider",
	Concurrency:  1,
}

type emptyOktaApplications struct{}

func (oktaApp *emptyOktaApplications) ListApplications(ctx context.Context, qp *query.Params) ([]okta.App, *okta.Response, error) {
	return nil, nil, nil
}

func TestNoEmail(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	idTokenVerifier := &idTokenVerifier{
		expectedIDToken: "this is an id token I want",
	}

	routerConfig := &RouterConfig{
		Verifier:            idTokenVerifier,
		AwsGenerationParams: &testAWSConfigGenerationParams,
		OktaAppClient:       &emptyOktaApplications{},
	}

	router := GetRouter(ctx, routerConfig)
	server := httptest.NewServer(router)
	defer server.Close()
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	r.NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("BEARER %s", idTokenVerifier.expectedIDToken))
	client := &http.Client{}

	resp, err := client.Do(req)
	r.NoError(err)
	r.Nil(getEmailFromCtx(req.Context()))
	r.Equal(400, resp.StatusCode)
}

func TestGetEmailFromCtx(t *testing.T) {
	r := require.New(t)

	ctxWithEmail := context.WithValue(context.Background(), contextKeyEmail, "foobar")
	email := getEmailFromCtx(ctxWithEmail)
	r.Equal(*email, "foobar")

	emptyCtx := context.Background()
	email = getEmailFromCtx(emptyCtx)
	r.Nil(email)

	newKeyValue := contextKeyEmail + 1
	ctxWithOtherKey := context.WithValue(context.Background(), newKeyValue, "barfoo")
	email = getEmailFromCtx(ctxWithOtherKey)
	r.Nil(email)
}

func TestMalformedBearerPrefix(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	idTokenVerifier := &idTokenVerifier{
		expectedIDToken: "this is an id token I want",
	}

	routerConfig := &RouterConfig{
		Verifier:            idTokenVerifier,
		AwsGenerationParams: &testAWSConfigGenerationParams,
		OktaAppClient:       &emptyOktaApplications{},
	}

	router := GetRouter(ctx, routerConfig)
	server := httptest.NewServer(router)
	defer server.Close()
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	r.NoError(err)
	// Given that we have an auth header and malformed prefix, we should get an error
	req.Header.Set("Authorization", fmt.Sprintf("BEARE %s", idTokenVerifier.expectedIDToken))
	client := &http.Client{}

	resp, err := client.Do(req)
	r.NoError(err)
	r.Nil(getEmailFromCtx(req.Context()))
	r.Equal(401, resp.StatusCode)
}

func TestMissingAuthHeader(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	routerConfig := &RouterConfig{
		Verifier:            &failingVerifier{},
		AwsGenerationParams: &testAWSConfigGenerationParams,
		OktaAppClient:       &emptyOktaApplications{},
	}

	router := GetRouter(ctx, routerConfig)
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL)
	r.NoError(err)
	r.Nil(getEmailFromCtx(resp.Request.Context()))
	r.Equal(407, resp.StatusCode)
}

type failingVerifier struct{}

func (fv *failingVerifier) Verify(ctx context.Context, idToken string) (*oidc.IDToken, error) {
	return nil, fmt.Errorf("Failing verifier")
}
func TestHealthEndpoint(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	routerConfig := &RouterConfig{
		Verifier:            &failingVerifier{},
		AwsGenerationParams: &testAWSConfigGenerationParams,
		OktaAppClient:       &emptyOktaApplications{},
	}

	router := GetRouter(ctx, routerConfig)
	server := httptest.NewServer(router)
	defer server.Close()

	healthURL := fmt.Sprintf("%s/health", server.URL)
	resp, err := http.Get(healthURL)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
}
