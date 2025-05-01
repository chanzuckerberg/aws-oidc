package aws_config_server

import (
	"context"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
)

var testClientMapping = []okta.OIDCRoleMapping{
	{
		AWSAccountAlias: "Account1",
		AWSRoleARN:      "OIDCFederatedRole1",
	},
	{
		AWSAccountAlias: "Account2",
		AWSRoleARN:      "arn:aws:iam::984830177581:role/readonly",
	},
	{
		AWSAccountAlias: "Account3",
		AWSRoleARN:      "arn:aws:iam::984830177581:role/readonly",
	},
}

func TestCreateAWSConfig(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	mapping := okta.OIDCRoleMappings(testClientMapping)
	config, err := createAWSConfig(
		ctx,
		"localhost",
		&mapping,
		[]okta.ClientID{"testClientID1", "testClientID2"},
	)
	r.NoError(err)
	r.NotEmpty(config)

	for _, accountName := range []string{"Account1", "Account2", "Account3"} {
		r.True(config.HasAccount(accountName))
	}
	r.Equal(config.GetRoleNames(), []string{"ValidRole"})
	r.Len(config.GetProfilesForAccount(AWSAccount{
		Name:  "Account2",
		Alias: "Account2",
	}), 2)
}
