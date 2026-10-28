package aws_config_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okta/okta-sdk-golang/v3/okta"
	"github.com/stretchr/testify/require"
)

type idTokenVerifier struct {
	expectedIDToken string
	principal       Principal
}

func (idtv *idTokenVerifier) Verify(_ context.Context, idToken string) (Principal, error) {
	if idtv.expectedIDToken != idToken {
		return Principal{}, fmt.Errorf("id tokens do not match")
	}
	return idtv.principal, nil
}

var testAWSConfigGenerationParams = AWSConfigGenerationParams{
	OIDCProvider: "validProvider",
	Concurrency:  1,
}

type emptyOktaApplications struct{}

func (m *emptyOktaApplications) ListApplications(_ context.Context, _, _ string) ([]okta.ListApplications200ResponseInner, string, error) {
	return nil, "", nil
}

func TestNoEmail(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	idTokenVerifier := &idTokenVerifier{
		expectedIDToken: "this is an id token I want",
		principal: Principal{
			Issuer:  testAWSConfigGenerationParams.OIDCProvider,
			Subject: testEntitlementUserID,
		},
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
	r.Equal(403, resp.StatusCode)
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
	r.Equal(401, resp.StatusCode)
}

type failingVerifier struct{}

func (fv *failingVerifier) Verify(context.Context, string) (Principal, error) {
	return Principal{}, errors.New("failing verifier")
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
