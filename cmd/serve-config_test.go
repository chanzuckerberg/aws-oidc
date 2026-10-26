package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOktaEnvEntitlementSettings(t *testing.T) {
	setRequiredOktaEnvironment(t)
	t.Setenv("OKTA_ENTITLEMENT_TOKEN_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("OKTA_ENTITLEMENT_TOKEN_AUDIENCE", "aws-oidc-entitlements")
	t.Setenv(
		"OKTA_ENTITLEMENT_TOKEN_TRUSTED_SUBJECT",
		"system:serviceaccount:argus-agent-registry-rdev:agent-registry-api",
	)

	environment, err := loadOktaEnv()

	require.NoError(t, err)
	require.Equal(t, "https://issuer.example.com", environment.ENTITLEMENT_TOKEN_ISSUER_URL)
	require.Equal(t, "aws-oidc-entitlements", environment.ENTITLEMENT_TOKEN_AUDIENCE)
	require.Equal(
		t,
		"system:serviceaccount:argus-agent-registry-rdev:agent-registry-api",
		environment.ENTITLEMENT_TOKEN_TRUSTED_SUBJECT,
	)
}

func TestLoadOktaEnvRejectsPartialEntitlementSettings(t *testing.T) {
	setRequiredOktaEnvironment(t)
	t.Setenv("OKTA_ENTITLEMENT_TOKEN_ISSUER_URL", "https://issuer.example.com")
	t.Setenv("OKTA_ENTITLEMENT_TOKEN_AUDIENCE", "")
	t.Setenv("OKTA_ENTITLEMENT_TOKEN_TRUSTED_SUBJECT", "")

	_, err := loadOktaEnv()

	require.ErrorContains(t, err, "all entitlement token settings are required")
}

func setRequiredOktaEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("OKTA_PRIVATE_KEY", "private-key")
	t.Setenv("OKTA_SERVICE_CLIENT_ID", "service-client")
	t.Setenv("OKTA_CLIENT_ID", "human-client")
	t.Setenv("OKTA_ISSUER_URL", "https://czi.okta.com")
}
