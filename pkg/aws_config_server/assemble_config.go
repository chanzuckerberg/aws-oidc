package aws_config_server

import (
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
)

// createAWSConfig projects a person's resolved AWS access into the config the client
// consumes. The access is already deduplicated by account and role, so this is a straight
// projection with no dedup of its own.
func createAWSConfig(oidcProvider string, access *awsaccess.Access) *AWSConfig {
	profiles := []AWSProfile{}
	for _, acct := range access.Accounts {
		for _, role := range acct.Roles {
			profiles = append(profiles, AWSProfile{
				ClientID: role.OktaClientID,
				RoleARN:  role.RoleARN,
				AWSAccount: AWSAccount{
					Name:  acct.Alias,
					Alias: acct.Alias,
					ID:    acct.ID,
				},
				IssuerURL: oidcProvider,
				RoleName:  role.RoleName,
			})
		}
	}
	return &AWSConfig{Profiles: profiles}
}
