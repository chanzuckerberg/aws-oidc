package aws_config_server

import (
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
)

var testClientMapping = []okta.OIDCRoleMapping{
	{
		AWSAccountID:    "984830177581",
		AWSAccountAlias: "Account1",
		AWSRoleARN:      "arn:aws:iam::984830177581:role/OIDCFederatedRole1",
		OktaClientID:    "ClientID1",
	},
	{
		AWSAccountID:    "984830177582",
		AWSAccountAlias: "Account2",
		AWSRoleARN:      "arn:aws:iam::984830177582:role/readonly",
		OktaClientID:    "ClientID2",
	},
	{
		AWSAccountID:    "984830177582",
		AWSAccountAlias: "Account2",
		AWSRoleARN:      "arn:aws:iam::984830177582:role/poweruser",
		OktaClientID:    "ClientID4",
	},
	{
		AWSAccountID:    "984830177583",
		AWSAccountAlias: "Account3",
		AWSRoleARN:      "arn:aws:iam::984830177583:role/poweruser",
		OktaClientID:    "ClientID3",
	},
}

func TestCreateAWSConfig(t *testing.T) {
	r := require.New(t)
	clientMappings := okta.OIDCRoleMappings(testClientMapping)
	clientMappingsByKey := make(okta.OIDCRoleMappingsByKey)
	for _, mapping := range clientMappings {
		_, ok := clientMappingsByKey[mapping.OktaClientID]
		if ok {
			clientMappingsByKey[mapping.OktaClientID] = append(clientMappingsByKey[mapping.OktaClientID], mapping)
		} else {
			clientMappingsByKey[mapping.OktaClientID] = []okta.OIDCRoleMapping{mapping}
		}
	}
	config, err := createAWSConfig("localhost", clientMappingsByKey, []okta.ClientID{"ClientID1", "ClientID2"})
	r.NoError(err)
	r.NotEmpty(config)

	for _, accountName := range []string{"Account1", "Account2"} {
		r.True(config.HasAccount(accountName))
	}
	r.ElementsMatch(config.GetRoleNames(), []string{"readonly", "OIDCFederatedRole1"})
	r.Len(config.GetProfilesForAccount(AWSAccount{
		Name:  "Account2",
		Alias: "Account2",
	}), 1)
}
