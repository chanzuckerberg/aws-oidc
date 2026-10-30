package aws_config_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	oktasdk "github.com/okta/okta-sdk-golang/v3/okta"
	"github.com/stretchr/testify/require"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
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

const testOIDCProvider = "https://issuer.example.com"

type emptyOktaApplications struct{}

func (m *emptyOktaApplications) ListApplications(_ context.Context, _, _ string) ([]oktasdk.ListApplications200ResponseInner, string, error) {
	return nil, "", nil
}

func emptyAccessResolver(t *testing.T) *AccessResolver {
	t.Helper()
	resolver, err := NewAccessResolver(
		testOIDCProvider,
		&emptyOktaApplications{},
		func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
			return nil, nil
		},
	)
	require.NoError(t, err)
	return resolver
}

func TestIndexAllowsMissingEmail(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	idTokenVerifier := &idTokenVerifier{
		expectedIDToken: "this is an id token I want",
		principal: Principal{
			Issuer:  testOIDCProvider,
			Subject: testEntitlementUserID,
		},
	}

	routerConfig := &RouterConfig{
		Verifier:       idTokenVerifier,
		AccessResolver: emptyAccessResolver(t),
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
	r.Equal(http.StatusOK, resp.StatusCode)
}

func TestIndexReturnsResolvedProfiles(t *testing.T) {
	apps := &entitlementAppLister{
		pages: [][]oktasdk.ListApplications200ResponseInner{
			{oktaApplication("client-a")},
		},
	}
	resolver, err := NewAccessResolver(
		testEntitlementIssuer,
		apps,
		func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
			return okta.OIDCRoleMappingsByKey{
				"client-a": {
					{
						AWSAccountID:    "111111111111",
						AWSAccountAlias: "alpha",
						AWSRoleARN:      "arn:aws:iam::111111111111:role/readonly",
						OktaClientID:    "client-a",
					},
				},
			}, nil
		},
	)
	require.NoError(t, err)
	verifier := &idTokenVerifier{
		expectedIDToken: "token",
		principal: Principal{
			Issuer:  testEntitlementIssuer,
			Subject: testEntitlementUserID,
			Email:   "user@example.com",
		},
	}
	server := httptest.NewServer(GetRouter(context.Background(), &RouterConfig{
		Verifier:       verifier,
		AccessResolver: resolver,
	}))
	defer server.Close()
	request, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer token")

	response, err := server.Client().Do(request)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer response.Body.Close()
	var config AWSConfig
	err = json.NewDecoder(response.Body).Decode(&config)
	require.NoError(t, err)
	require.Equal(t, []AWSProfile{
		{
			ClientID: "client-a",
			AWSAccount: AWSAccount{
				ID:    "111111111111",
				Name:  "alpha",
				Alias: "alpha",
			},
			RoleARN:   "arn:aws:iam::111111111111:role/readonly",
			IssuerURL: testEntitlementIssuer,
			RoleName:  "readonly",
		},
	}, config.Profiles)
}

func TestMalformedBearerPrefix(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	idTokenVerifier := &idTokenVerifier{
		expectedIDToken: "this is an id token I want",
	}

	routerConfig := &RouterConfig{
		Verifier:       idTokenVerifier,
		AccessResolver: emptyAccessResolver(t),
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
		Verifier:       &failingVerifier{},
		AccessResolver: emptyAccessResolver(t),
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
		Verifier:       &failingVerifier{},
		AccessResolver: emptyAccessResolver(t),
	}

	router := GetRouter(ctx, routerConfig)
	server := httptest.NewServer(router)
	defer server.Close()

	healthURL := fmt.Sprintf("%s/health", server.URL)
	resp, err := http.Get(healthURL)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
}
