package aws_config_server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	oktasdk "github.com/okta/okta-sdk-golang/v3/okta"
	"github.com/stretchr/testify/require"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

const (
	testEntitlementIssuer  = "https://czi.okta.com"
	testEntitlementSubject = "system:serviceaccount:argus-agent-registry-rdev:agent-registry-api"
	testEntitlementUserID  = "00u12345678901234567"
)

type staticIdentityVerifier struct {
	principal Principal
	err       error
}

func (v *staticIdentityVerifier) Verify(context.Context, string) (Principal, error) {
	if v.err != nil {
		return Principal{}, v.err
	}
	return v.principal, nil
}

type entitlementAppLister struct {
	pages [][]oktasdk.ListApplications200ResponseInner
	calls int
}

func (l *entitlementAppLister) ListApplications(
	_ context.Context,
	filter string,
	_ string,
) ([]oktasdk.ListApplications200ResponseInner, string, error) {
	if filter != `user.id eq "`+testEntitlementUserID+`"` {
		return nil, "", errors.New("unexpected filter")
	}
	page := l.pages[l.calls]
	l.calls++
	if l.calls < len(l.pages) {
		return page, "next", nil
	}
	return page, "", nil
}

func TestEntitlementEndpoint(t *testing.T) {
	apps := &entitlementAppLister{
		pages: [][]oktasdk.ListApplications200ResponseInner{
			{oktaApplication("client-b")},
			{oktaApplication("client-a")},
		},
	}
	mappings := okta.OIDCRoleMappingsByKey{
		"client-a": {
			{
				AWSAccountID:    "111111111111",
				AWSAccountAlias: "alpha",
				AWSRoleARN:      "arn:aws:iam::111111111111:role/readonly",
				OktaClientID:    "client-a",
			},
		},
		"client-b": {
			{
				AWSAccountID:    "222222222222",
				AWSAccountAlias: "beta",
				AWSRoleARN:      "arn:aws:iam::222222222222:role/team/poweruser",
				OktaClientID:    "client-b",
			},
			{
				AWSAccountID:    "111111111111",
				AWSAccountAlias: "alpha",
				AWSRoleARN:      "arn:aws:iam::111111111111:role/readonly",
				OktaClientID:    "client-b",
			},
		},
	}
	handler := entitlementEndpoint(t, &staticIdentityVerifier{principal: privilegedTestPrincipal()}, apps, func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
		return mappings, nil
	})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/entitlements",
		strings.NewReader(`{"issuer":"https://czi.okta.com","userId":"00u12345678901234567"}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set(requestIDHeader, "request-123")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, "request-123", response.Header().Get(requestIDHeader))
	var result EntitlementList
	err := json.NewDecoder(response.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, EntitlementList{
		Items: []Entitlement{
			{
				AccountID:    "111111111111",
				AccountAlias: "alpha",
				RoleARN:      "arn:aws:iam::111111111111:role/readonly",
				RoleName:     "readonly",
			},
			{
				AccountID:    "222222222222",
				AccountAlias: "beta",
				RoleARN:      "arn:aws:iam::222222222222:role/team/poweruser",
				RoleName:     "team/poweruser",
			},
		},
	}, result)
	require.Equal(t, 2, apps.calls)
}

func TestEntitlementAuthentication(t *testing.T) {
	tests := map[string]struct {
		authorization string
		verifier      *staticIdentityVerifier
		status        int
	}{
		"missing token": {
			verifier: &staticIdentityVerifier{principal: privilegedTestPrincipal()},
			status:   http.StatusUnauthorized,
		},
		"invalid token": {
			authorization: "Bearer invalid",
			verifier:      &staticIdentityVerifier{err: errors.New("wrong issuer, audience or expiry")},
			status:        http.StatusUnauthorized,
		},
		"unprivileged cross-user lookup": {
			authorization: "Bearer token",
			verifier: &staticIdentityVerifier{principal: Principal{
				Issuer:  "https://issuer.example.com",
				Subject: "unprivileged",
			}},
			status: http.StatusForbidden,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler := entitlementEndpoint(
				t,
				test.verifier,
				&entitlementAppLister{pages: [][]oktasdk.ListApplications200ResponseInner{{}}},
				func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
					return nil, nil
				},
			)
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/entitlements",
				strings.NewReader(`{"issuer":"https://czi.okta.com","userId":"00u12345678901234567"}`),
			)
			request.Header.Set("Authorization", test.authorization)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, test.status, response.Code)
		})
	}
}

func TestEntitlementEndpointUsesNormalUserIdentity(t *testing.T) {
	apps := &entitlementAppLister{
		pages: [][]oktasdk.ListApplications200ResponseInner{{}},
	}
	handler := entitlementEndpoint(
		t,
		&staticIdentityVerifier{principal: Principal{
			Issuer:  testEntitlementIssuer,
			Subject: testEntitlementUserID,
			Email:   "user@example.com",
		}},
		apps,
		func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
			return nil, nil
		},
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/entitlements",
		strings.NewReader(`{}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, 1, apps.calls)
}

func TestEntitlementRequestValidation(t *testing.T) {
	handler := entitlementEndpoint(
		t,
		&staticIdentityVerifier{principal: privilegedTestPrincipal()},
		&entitlementAppLister{pages: [][]oktasdk.ListApplications200ResponseInner{{}}},
		func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
			return nil, nil
		},
	)
	tests := map[string]string{
		"wrong issuer":  `{"issuer":"https://example.com","userId":"00u12345678901234567"}`,
		"invalid user":  `{"issuer":"https://czi.okta.com","userId":"00u\" or true"}`,
		"unknown field": `{"issuer":"https://czi.okta.com","userId":"00u12345678901234567","other":true}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/entitlements", strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer token")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestEntitlementDependencyFailure(t *testing.T) {
	handler := entitlementEndpoint(
		t,
		&staticIdentityVerifier{principal: privilegedTestPrincipal()},
		&entitlementAppLister{pages: [][]oktasdk.ListApplications200ResponseInner{{}}},
		func(context.Context) (okta.OIDCRoleMappingsByKey, error) {
			return nil, errors.New("rolemap unavailable")
		},
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/entitlements",
		strings.NewReader(`{"issuer":"https://czi.okta.com","userId":"00u12345678901234567"}`),
	)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestResolveEntitlementsRejectsAccountMismatch(t *testing.T) {
	apps := &entitlementAppLister{
		pages: [][]oktasdk.ListApplications200ResponseInner{{oktaApplication("client-a")}},
	}
	_, err := resolveEntitlements(
		context.Background(),
		testEntitlementUserID,
		apps,
		okta.OIDCRoleMappingsByKey{
			"client-a": {
				{
					AWSAccountID: "111111111111",
					AWSRoleARN:   "arn:aws:iam::222222222222:role/readonly",
				},
			},
		},
	)
	require.ErrorContains(t, err, "does not match account")
}

func entitlementEndpoint(
	t *testing.T,
	verifier IdentityVerifier,
	apps okta.AppLister,
	mappings MappingsProvider,
) http.Handler {
	t.Helper()
	handler, err := NewEntitlementHandler(testEntitlementIssuer, apps, mappings)
	require.NoError(t, err)
	return NewAuthMiddleware(handler, verifier)
}

func privilegedTestPrincipal() Principal {
	return Principal{
		Issuer:     "https://oidc.eks.example.com",
		Subject:    testEntitlementSubject,
		Privileged: true,
	}
}

func oktaApplication(id string) oktasdk.ListApplications200ResponseInner {
	app := oktasdk.NewOpenIdConnectApplication()
	app.Id = &id
	return oktasdk.OpenIdConnectApplicationAsListApplications200ResponseInner(app)
}
