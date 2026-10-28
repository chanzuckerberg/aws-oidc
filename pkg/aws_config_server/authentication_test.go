package aws_config_server

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	oidc "github.com/coreos/go-oidc"
	"github.com/stretchr/testify/require"
)

type fakeOIDCVerifier struct {
	token   *oidc.IDToken
	err     error
	calls   int
	rawSeen string
}

func (v *fakeOIDCVerifier) Verify(_ context.Context, rawToken string) (*oidc.IDToken, error) {
	v.calls++
	v.rawSeen = rawToken
	if v.err != nil {
		return nil, v.err
	}
	return v.token, nil
}

func TestMultiVerifierSelectsIssuerAndMarksPrivilegedSubject(t *testing.T) {
	oktaVerifier := &fakeOIDCVerifier{
		token: &oidc.IDToken{
			Issuer:  "https://czi.okta.com",
			Subject: testEntitlementUserID,
		},
	}
	kubernetesVerifier := &fakeOIDCVerifier{
		token: &oidc.IDToken{
			Issuer:  "https://oidc.eks.example.com",
			Subject: testEntitlementSubject,
		},
	}
	verifier, err := NewMultiVerifier(
		[]VerifierConfig{
			{Issuer: "https://czi.okta.com", Verifier: oktaVerifier},
			{Issuer: "https://oidc.eks.example.com", Verifier: kubernetesVerifier},
		},
		[]PrivilegedIdentity{
			{
				Issuer:  "https://oidc.eks.example.com",
				Subject: testEntitlementSubject,
			},
		},
	)
	require.NoError(t, err)
	rawToken := unsignedTestToken("https://oidc.eks.example.com", "")

	principal, err := verifier.Verify(context.Background(), rawToken)

	require.NoError(t, err)
	require.Equal(t, Principal{
		Issuer:     "https://oidc.eks.example.com",
		Subject:    testEntitlementSubject,
		Privileged: true,
	}, principal)
	require.Zero(t, oktaVerifier.calls)
	require.Equal(t, 1, kubernetesVerifier.calls)
	require.Equal(t, rawToken, kubernetesVerifier.rawSeen)
}

func TestMultiVerifierSupportsMultipleAudiencesForOneIssuer(t *testing.T) {
	first := &fakeOIDCVerifier{err: errors.New("wrong audience")}
	second := &fakeOIDCVerifier{
		token: &oidc.IDToken{
			Issuer:  "https://czi.okta.com",
			Subject: testEntitlementUserID,
		},
	}
	verifier, err := NewMultiVerifier(
		[]VerifierConfig{
			{Issuer: "https://czi.okta.com", Verifier: first},
			{Issuer: "https://czi.okta.com", Verifier: second},
		},
		nil,
	)
	require.NoError(t, err)

	principal, err := verifier.Verify(
		context.Background(),
		unsignedTestToken("https://czi.okta.com", "user@example.com"),
	)

	require.NoError(t, err)
	require.Equal(t, "user@example.com", principal.Email)
	require.Equal(t, 1, first.calls)
	require.Equal(t, 1, second.calls)
}

func TestMultiVerifierMatchesPrivilegedIssuerAndSubject(t *testing.T) {
	oktaVerifier := &fakeOIDCVerifier{
		token: &oidc.IDToken{
			Issuer:  "https://czi.okta.com",
			Subject: testEntitlementSubject,
		},
	}
	kubernetesVerifier := &fakeOIDCVerifier{
		token: &oidc.IDToken{
			Issuer:  "https://oidc.eks.example.com",
			Subject: testEntitlementSubject,
		},
	}
	verifier, err := NewMultiVerifier(
		[]VerifierConfig{
			{Issuer: "https://czi.okta.com", Verifier: oktaVerifier},
			{Issuer: "https://oidc.eks.example.com", Verifier: kubernetesVerifier},
		},
		[]PrivilegedIdentity{
			{
				Issuer:  "https://oidc.eks.example.com",
				Subject: testEntitlementSubject,
			},
		},
	)
	require.NoError(t, err)

	principal, err := verifier.Verify(
		context.Background(),
		unsignedTestToken("https://czi.okta.com", ""),
	)

	require.NoError(t, err)
	require.False(t, principal.Privileged)
}

func TestMultiVerifierRejectsUnknownIssuer(t *testing.T) {
	configured := &fakeOIDCVerifier{}
	verifier, err := NewMultiVerifier(
		[]VerifierConfig{
			{Issuer: "https://czi.okta.com", Verifier: configured},
		},
		nil,
	)
	require.NoError(t, err)

	_, err = verifier.Verify(
		context.Background(),
		unsignedTestToken("https://untrusted.example.com", ""),
	)

	require.ErrorContains(t, err, "not trusted")
	require.Zero(t, configured.calls)
}

func TestAuthMiddlewareAttachesVerifiedPrincipal(t *testing.T) {
	expected := Principal{
		Issuer:  "https://czi.okta.com",
		Subject: testEntitlementUserID,
		Email:   "user@example.com",
	}
	verifier := &staticIdentityVerifier{principal: expected}
	handler := NewAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromContext(request.Context())
		require.True(t, ok)
		require.Equal(t, expected, principal)
		w.WriteHeader(http.StatusNoContent)
	}), verifier)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set(requestIDHeader, "request-123")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "request-123", response.Header().Get(requestIDHeader))
}

func unsignedTestToken(issuer string, email string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"iss":"` + issuer + `","email":"` + email + `"}`),
	)
	return header + "." + payload + ".signature"
}
