package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOktaEnvAdditionalIdentitySettings(t *testing.T) {
	setRequiredOktaEnvironment(t)
	t.Setenv(
		"OKTA_ADDITIONAL_TOKEN_VERIFIERS",
		`[{"issuerUrl":"https://first.example.com","clientId":"first"},{"issuerUrl":"https://second.example.com","clientId":"second"}]`,
	)
	t.Setenv(
		"OKTA_PRIVILEGED_IDENTITIES",
		`[{"issuerUrl":"https://first.example.com","subject":"system:serviceaccount:namespace:api"}]`,
	)

	environment, err := loadOktaEnv()

	require.NoError(t, err)
	verifiers, privileged, err := parseAdditionalIdentityConfiguration(environment)
	require.NoError(t, err)
	require.Equal(t, []TokenVerifierEnvironment{
		{IssuerURL: "https://first.example.com", ClientID: "first"},
		{IssuerURL: "https://second.example.com", ClientID: "second"},
	}, verifiers)
	require.Equal(t, []PrivilegedIdentityEnvironment{
		{
			IssuerURL: "https://first.example.com",
			Subject:   "system:serviceaccount:namespace:api",
		},
	}, privileged)
}

func TestLoadOktaEnvRejectsInvalidIdentitySettings(t *testing.T) {
	setRequiredOktaEnvironment(t)
	t.Setenv("OKTA_ADDITIONAL_TOKEN_VERIFIERS", "not-json")

	_, err := loadOktaEnv()

	require.ErrorContains(t, err, "parse additional token verifiers")
}

func setRequiredOktaEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("OKTA_PRIVATE_KEY", "private-key")
	t.Setenv("OKTA_SERVICE_CLIENT_ID", "service-client")
	t.Setenv("OKTA_CLIENT_ID", "human-client")
	t.Setenv("OKTA_ISSUER_URL", "https://czi.okta.com")
}
