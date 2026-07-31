package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
	czokta "github.com/chanzuckerberg/aws-oidc/pkg/okta"
	oidc "github.com/coreos/go-oidc"
	"github.com/okta/okta-sdk-golang/v3/okta"
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

func (m *emptyOktaApplications) ListApplications(_ context.Context, _, _ string) ([]okta.ListApplications200ResponseInner, string, error) {
	return nil, "", nil
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

	ctxWithEmail := identity.NewContext(context.Background(), &identity.User{Email: "foobar"})
	email := getEmailFromCtx(ctxWithEmail)
	r.Equal(*email, "foobar")

	emptyCtx := context.Background()
	email = getEmailFromCtx(emptyCtx)
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

// oneOIDCApp is an AppLister that reports a single assigned Okta app with the given id, so a
// human profile can be resolved from the rolemap in tests.
type oneOIDCApp struct{ id string }

func (o *oneOIDCApp) ListApplications(_ context.Context, _, _ string) ([]okta.ListApplications200ResponseInner, string, error) {
	app := okta.NewOpenIdConnectApplication()
	app.Id = &o.id
	return []okta.ListApplications200ResponseInner{
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(app),
	}, "", nil
}

// The endpoint returns the human profiles and, when an AgentsProvider is wired, the caller's
// agent profiles too, in the same response.
func TestIndexReturnsHumanAndAgentProfiles(t *testing.T) {
	r := require.New(t)

	mappings := czokta.OIDCRoleMappingsByKey{
		"ClientHuman": {{
			AWSAccountID:    "111111111111",
			AWSAccountAlias: "prod",
			AWSRoleARN:      "arn:aws:iam::111111111111:role/poweruser",
			OktaClientID:    "ClientHuman",
		}},
	}
	mappingsProvider := func(context.Context) (czokta.OIDCRoleMappingsByKey, error) { return mappings, nil }

	agentsProvider := func(_ context.Context, sub string) ([]AgentConfig, error) {
		r.Equal("00uOwner", sub, "the agents provider is keyed on the caller's subject")
		return []AgentConfig{{
			Name: "data-bot",
			Profiles: []AWSProfile{{
				ClientID:   "0oaAGENT",
				RoleARN:    "arn:aws:iam::111111111111:role/agents/data-bot-ro",
				RoleName:   "agents/data-bot-ro",
				IssuerURL:  "validProvider",
				AWSAccount: AWSAccount{ID: "111111111111", Name: "prod", Alias: "prod"},
			}},
		}}, nil
	}

	handler := Index(&testAWSConfigGenerationParams, &oneOIDCApp{id: "ClientHuman"}, mappingsProvider, agentsProvider)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(identity.NewContext(req.Context(), &identity.User{Sub: "00uOwner", Email: "owner@example.com"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	r.Equal(200, rec.Code)

	var cfg AWSConfig
	err := json.NewDecoder(rec.Body).Decode(&cfg)
	r.NoError(err)

	r.Len(cfg.Profiles, 1)
	r.Equal("poweruser", cfg.Profiles[0].RoleName)

	r.Len(cfg.Agents, 1)
	r.Equal("data-bot", cfg.Agents[0].Name)
	r.Len(cfg.Agents[0].Profiles, 1)
	r.Equal("agents/data-bot-ro", cfg.Agents[0].Profiles[0].RoleName)
	r.Equal("0oaAGENT", cfg.Agents[0].Profiles[0].ClientID)
}

// With no AgentsProvider (the feature disabled, as in prod), the response carries only human
// profiles and no agents.
func TestIndexWithoutAgentsProvider(t *testing.T) {
	r := require.New(t)

	mappingsProvider := func(context.Context) (czokta.OIDCRoleMappingsByKey, error) {
		return czokta.OIDCRoleMappingsByKey{}, nil
	}

	handler := Index(&testAWSConfigGenerationParams, &emptyOktaApplications{}, mappingsProvider, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(identity.NewContext(req.Context(), &identity.User{Sub: "00uOwner", Email: "owner@example.com"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	r.Equal(200, rec.Code)

	var cfg AWSConfig
	err := json.NewDecoder(rec.Body).Decode(&cfg)
	r.NoError(err)
	r.Empty(cfg.Agents)
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
