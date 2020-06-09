package aws_config_server

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type ClientIDToAWSRoles struct {
	clientRoleMapping map[string][]ConfigProfile
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
		logrus.Debugf("function: aws_config_server/assemble_config.go/getRoles(), accountList: %v", accountList)
		for _, acct := range accountList {
			// create a new IAM session for each account
			new_role_arn := arn.ARN{
				Partition: "aws",
				Service:   "iam",
				AccountID: *acct.Id,
				Resource:  fmt.Sprintf("role/%s", workerRole),
			}
			logrus.Debugf("function: aws_config_server/assemble_config.go/getRoles(), new_role_arn: %s", new_role_arn)
			a.roleARNs[*acct.Name] = new_role_arn
		}
	}
	return nil
}

func (a *ClientIDToAWSRoles) mapRoles(ctx context.Context, oidcProvider string) error {
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

		logrus.Debugf("function: aws_config_server/assemble_config.go/mapRoles(), workerRoles: %v", workerRoles)
		err = clientRoleMapFromProfile(ctx, accountName, workerRoles, oidcProvider, a.clientRoleMapping)
		if err != nil {
			return errors.Wrap(err, "Unable to complete mapping between ClientIDs and ConfigProfiles")
		}
	}
	return nil
}

func createAWSConfig(
	ctx context.Context,
	configParams *AWSConfigGenerationParams,
	clientMapping map[string][]ConfigProfile,
	userClientIDs []string) (*ini.File, error) {
	configFile := ini.Empty()

	for _, clientID := range userClientIDs {
		configList := clientMapping[clientID]
		for _, config := range configList {

			profileNoSpace := strings.ReplaceAll(config.AcctName, " ", "-")
			profileNoSpaceLowercase := strings.ToLower(profileNoSpace)

			profileSection := fmt.Sprintf("profile %s", profileNoSpaceLowercase)
			credsProcessValue := fmt.Sprintf(
				"sh -c 'aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s 2> /dev/tty'",
				configParams.OIDCProvider,
				clientID,
				config.RoleARN,
			)

			section, err := configFile.NewSection(profileSection)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to create %s section in AWS Config", profileSection)
			}
			section.Key("output").SetValue("json")
			section.Key("credential_process").SetValue(credsProcessValue)
		}
	}
	return configFile, nil
}
