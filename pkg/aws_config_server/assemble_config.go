package aws_config_server

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

// we send back a json representation of our config that can be consumed by the client
// using the configure command
func createAWSConfig(oidcProvider string, clientMapping *okta.OIDCRoleMappings) (*AWSConfig, error) {
	awsConfig := &AWSConfig{
		Profiles: []AWSProfile{},
	}

	for _, mapping := range *clientMapping {
		roleARN, err := arn.Parse(mapping.AWSRoleARN)
		if err != nil {
			return nil, fmt.Errorf("parsing role ARN: %w", err)
		}
		profile := AWSProfile{
			ClientID: mapping.OktaClientID,
			RoleARN:  mapping.AWSRoleARN,
			AWSAccount: AWSAccount{
				Name:  mapping.AWSAccountAlias,
				Alias: mapping.AWSAccountAlias,
				ID:    mapping.AWSAccountID,
			},
			IssuerURL: oidcProvider,
			RoleName:  strings.ReplaceAll(roleARN.Resource, "role/", ""),
		}
		awsConfig.Profiles = append(awsConfig.Profiles, profile)
	}

	return awsConfig, nil
}
