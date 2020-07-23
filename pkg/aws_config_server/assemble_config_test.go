package aws_config_server

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
)

var testClientMapping = &oidcFederatedRoles{
	roles: map[okta.ClientID][]accountAndRole{
		"testClientID1": {
			{
				AccountName:  "Account1",
				AccountAlias: aws.String("Account1"),
				RoleARN:      BareRoleARN("OIDCFederatedRole1"),
				Role: &iam.Role{
					RoleName: aws.String("ValidRole"),
				},
			},
			{
				AccountName:  "Account2",
				AccountAlias: aws.String("Account2"),
				RoleARN:      &arn.ARN{},
				Role: &iam.Role{
					RoleName: aws.String("ValidRole"),
				},
			},
			{
				AccountName:  "Account3",
				AccountAlias: aws.String("Account3"),
				RoleARN:      &arn.ARN{},
				Role: &iam.Role{
					RoleName: aws.String("ValidRole"),
				},
			},
		},
		"testClientID2": {
			{
				AccountName:  "Account2",
				AccountAlias: aws.String("Account2"),
				RoleARN:      &arn.ARN{},
				Role: &iam.Role{
					RoleName: aws.String("ValidRole"),
				},
			},
		},
	},
}

func TestCreateAWSConfig(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	config, err := createAWSConfig(
		ctx,
		"localhost",
		testClientMapping,
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
