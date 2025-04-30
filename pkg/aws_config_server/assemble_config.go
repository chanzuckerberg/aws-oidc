package aws_config_server

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
)

type ClientIDToAWSRoles struct {
	clientRoleMapping *oidcFederatedRoles
	roleARNs          map[string]arn.ARN

	awsSession *session.Session
	awsClient  *cziAWS.Client
}

// we send back a json representation of our config that can be consumed by the client
// using the configure command
func createAWSConfig(
	ctx context.Context,
	oidcProvider string,
	clientMapping *okta.OIDCRoleMappingByClientID,
	userClientIDs []okta.ClientID) (*AWSConfig, error) {

	awsConfig := &AWSConfig{
		Profiles: []AWSProfile{},
	}

	for _, clientID := range userClientIDs {
		configList := clientMapping.roles[clientID]

		for _, config := range *clientMapping {
			profile := AWSProfile{
				ClientID: clientID,
				RoleARN:  config.RoleARN.String(),
				AWSAccount: AWSAccount{
					Name:  config.AccountName,
					Alias: *config.AccountAlias,
					ID:    config.RoleARN.AccountID,
				},
				IssuerURL: oidcProvider,
				RoleName:  *config.Role.RoleName,
			}

			awsConfig.Profiles = append(awsConfig.Profiles, profile)
		}
	}

	return awsConfig, nil
}
