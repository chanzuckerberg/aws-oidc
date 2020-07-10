package aws_config_server

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
)

type ClientIDToAWSRoles struct {
	clientRoleMapping map[okta.ClientID][]ConfigProfile
	roleARNs          map[string]arn.ARN

	awsSession *session.Session
	awsClient  *cziAWS.Client
}

// we send back a json representation of our config that can be consumed by the client
// using the configure command
func createAWSConfig(
	ctx context.Context,
	configParams *AWSConfigGenerationParams,
	clientMapping map[okta.ClientID][]ConfigProfile,
	userClientIDs []okta.ClientID) (*AWSConfig, error) {

	awsConfig := &AWSConfig{
		Profiles: []AWSProfile{},
	}

	for _, clientID := range userClientIDs {
		configList := clientMapping[clientID]
		for _, config := range configList {
			profile := AWSProfile{
				ClientID: clientID,
				RoleARN:  config.RoleARN.String(),
				AWSAccount: AWSAccount{
					Name:  config.AcctName,
					Alias: config.AcctAlias,
					ID:    config.RoleARN.AccountID,
				},
				IssuerURL: configParams.OIDCProvider,
				RoleName:  config.RoleName,
			}

			awsConfig.Profiles = append(awsConfig.Profiles, profile)
		}
	}

	return awsConfig, nil
}
