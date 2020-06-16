package aws_config_server

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
)

type ClientIDToAWSRoles struct {
	clientRoleMapping map[okta.ClientID][]ConfigProfile
	roleARNs          map[string]arn.ARN

	awsSession *session.Session
	awsClient  *cziAWS.Client
}

func (a *ClientIDToAWSRoles) getRoles(ctx context.Context, masterRoles []string, workerRole string) error {
	for _, role_arn := range masterRoles {
		masterAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, role_arn),
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
		orgClient := a.awsClient.WithOrganizations(masterAWSConfig).Organizations.Svc
		accountList, err := GetActiveAccountList(ctx, orgClient)
		if err != nil {
			return errors.Wrap(err, "Unable to get list of AWS Profiles")
		}
		for _, acct := range accountList {
			// create a new IAM session for each account
			new_role_arn := arn.ARN{
				Partition: "aws",
				Service:   "iam",
				AccountID: *acct.Id,
				Resource:  fmt.Sprintf("role/%s", workerRole),
			}
			a.roleARNs[*acct.Name] = new_role_arn
		}
	}
	return nil
}

func (a *ClientIDToAWSRoles) mapRoles(
	ctx context.Context,
	oidcProvider string,
) error {
	for accountName, roleARN := range a.roleARNs {
		workerAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, roleARN.String()),
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
		iamClient := a.awsClient.WithIAM(workerAWSConfig).IAM.Svc
		workerRoles, err := listRoles(ctx, iamClient)
		if err != nil {
			return errors.Wrapf(err, "%s error", accountName)
		}

		err = clientRoleMapFromProfile(ctx, accountName, workerRoles, oidcProvider, a.clientRoleMapping)
		if err != nil {
			return errors.Wrap(err, "Unable to complete mapping between ClientIDs and ConfigProfiles")
		}
	}
	return nil
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
					Name: config.AcctName,
					ID:   config.RoleARN.AccountID,
				},
				IssuerURL: configParams.OIDCProvider,
				RoleName:  config.RoleName,
			}
			awsConfig.Profiles = append(awsConfig.Profiles, profile)
		}
	}
	return awsConfig, nil
}
