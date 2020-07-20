package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PolicyDocument struct {
	Version    string           `json:"Version"`
	Statements []StatementEntry `json:"Statement"`
}

type StatementEntry struct {
	Effect    string    `json:"Effect"`
	Action    Action    `json:"Action"`
	Sid       string    `json:"Sid"`
	Principal Principal `json:"Principal"`
	Condition Condition `json:"Condition"`
}

// We only care about the "StringEquals" field in Condition
type Condition struct {
	StringEquals map[string]string `json:"StringEquals"`
}

// We only care about the "Federated" field in Principal
type Principal struct {
	Federated string `json:"Federated"`
}

type ConfigProfile struct {
	AcctName  string
	AcctAlias string
	RoleARN   arn.ARN
	RoleName  string
}

type roleARNMatch struct {
	accountName string
	accountARN  arn.ARN
}

func (a *ClientIDToAWSRoles) getWorkerRoles(ctx context.Context, orgRoles []string, workerRole string) error {
	ctx, span := beeline.StartSpan(ctx, "server_get_worker_roles")
	defer span.Send()
	for _, role_arn := range orgRoles {

		beeline.AddField(ctx, "Organization IAM Role ARN", role_arn)

		orgAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(a.awsSession, role_arn),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Retryer:                       a.awsSession.Config.Retryer,
		}

		orgClient := a.awsClient.WithOrganizations(orgAWSConfig).Organizations.Svc
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

func getRoleNames(roles []*iam.Role) string {
	out := []string{}
	for _, role := range roles {
		if role == nil || role.Arn == nil {
			continue
		}
		out = append(out, *role.Arn)
	}

	return strings.Join(out, ",")
}

func listRoleTags(ctx context.Context, svc iamiface.IAMAPI, roleName *string) ([]*iam.Tag, error) {
	ctx, span := beeline.StartSpan(ctx, "list AWS Role Tags")
	defer span.Send()

	if roleName == nil {
		return nil, nil
	}

	input := &iam.ListRoleTagsInput{
		RoleName: roleName,
	}

	output, err := svc.ListRoleTagsWithContext(ctx, input)
	if processAWSErr(err) != nil {
		return nil, errors.Wrapf(err, "could not list tags for role %s", *roleName)
	}
	return output.Tags, nil
}

func listRoles(ctx context.Context, svc iamiface.IAMAPI, configParams *AWSConfigGenerationParams) ([]*iam.Role, error) {
	ctx, span := beeline.StartSpan(ctx, "list AWS Roles")
	defer span.Send()

	// Run the AWS list-roles command and save the output
	input := &iam.ListRolesInput{}
	output := []*iam.Role{}
	err := svc.ListRolesPagesWithContext(ctx,
		input,
		func(page *iam.ListRolesOutput, lastPage bool) bool {
			output = append(output, page.Roles...)
			return !lastPage
		},
	)
	if processAWSErr(err) != nil {
		return output, errors.Wrap(err, "Error listing IAM roles")
	}

	return filterRoles(ctx, svc, output, configParams)
}

type Action []string

func (s *Action) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		*s = []string{str}
		return nil
	}
	// If the error is not an unmarshal type error, then we return the error
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return errors.Wrap(err, "Unexpected error type from unmarshaling")
	}

	var strSlice []string
	err = json.Unmarshal(data, &strSlice)
	if err == nil {
		*s = strSlice
		return nil
	}
	return errors.Wrap(err, "Unable to unmarshal Action")
}

func getRoleMappings(ctx context.Context,
	acctName string,
	acctAlias string,
	roles []*iam.Role,
	oidcProvider string) (map[okta.ClientID][]ConfigProfile, error) {

	clientRoleMapping := make(map[okta.ClientID][]ConfigProfile)

	identityProviderURL, err := url.Parse(oidcProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse OIDC Provider input as an URL")
	}
	oidcProviderHostname := identityProviderURL.Hostname()

	for _, role := range roles {
		if role == nil {
			logrus.Debug("nil role")
			continue
		}

		if role.AssumeRolePolicyDocument == nil {
			continue // role doesn't have an assume role policy document
		}
		// the IAM Role outputs a url-encoded policy document, so we need to escape characters
		policyStr, err := url.PathUnescape(*role.AssumeRolePolicyDocument)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to escape URL encoding")
		}
		policyDoc := PolicyDocument{}
		err = json.Unmarshal([]byte(policyStr), &policyDoc)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to unmarshal policy document to struct policy: %s", policyStr)
		}

		for _, statement := range policyDoc.Statements {
			federatedARN := statement.Principal.Federated
			if !strings.Contains(federatedARN, oidcProviderHostname) {
				continue
			}

			clientKey := fmt.Sprintf("%s:aud", oidcProviderHostname)
			clientIDStr, ok := statement.Condition.StringEquals[clientKey]
			if !ok || (clientIDStr == "") {
				continue
			}

			clientID := okta.ClientID(clientIDStr)

			// Searching through the Actions list
			isWebIdentityAction := false
			for _, action := range statement.Action {
				if action == "sts:AssumeRoleWithWebIdentity" {
					isWebIdentityAction = true
					break
				}
			}
			if !isWebIdentityAction {
				continue
			}

			roleARN, err := arn.Parse(*role.Arn)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse arn %s", *role.Arn)
			}

			currentConfig := ConfigProfile{
				AcctName:  acctName,
				AcctAlias: acctAlias,
				RoleARN:   roleARN,
				RoleName:  *role.RoleName,
			}

			if _, ok := clientRoleMapping[clientID]; !ok {
				clientRoleMapping[clientID] = []ConfigProfile{currentConfig}
				continue
			}
			clientRoleMapping[clientID] = append(clientRoleMapping[clientID], currentConfig)
		}
	}
	return clientRoleMapping, nil
}

func GetActiveAccountList(
	ctx context.Context,
	svc organizationsiface.OrganizationsAPI,
) ([]*organizations.Account, error) {
	orgInput := &organizations.ListAccountsInput{}

	orgAccounts := []*organizations.Account{}

	err := svc.ListAccountsPagesWithContext(
		ctx,
		orgInput,
		func(page *organizations.ListAccountsOutput, lastPage bool) bool {
			orgAccounts = append(orgAccounts, page.Accounts...)
			return !lastPage
		},
	)
	if processAWSErr(err) != nil {
		return nil, errors.Wrap(err, "Unable to List Accounts from organizations session")
	}

	var activeAccounts []*organizations.Account
	for _, acct := range orgAccounts {
		if *acct.Status == "ACTIVE" {
			activeAccounts = append(activeAccounts, acct)
		}
	}

	return activeAccounts, nil
}

func getAcctAlias(ctx context.Context, svc iamiface.IAMAPI) (string, error) {
	input := &iam.ListAccountAliasesInput{}
	output, err := svc.ListAccountAliases(input)
	if processAWSErr(err) != nil {
		return "", errors.Wrap(err, "Error getting account alias")
	}

	// no alias
	if output == nil || len(output.AccountAliases) == 0 {
		return "", nil
	}
	return *output.AccountAliases[0], nil
}

// process an aws err to see if we should skip or not
func processAWSErr(err error) error {
	if err == nil {
		return nil
	}

	awsErr, ok := err.(awserr.Error)
	if !ok {
		return err
	}
	if awsErr.Code() == errAWSAccessDenied {
		logrus.WithError(err).Errorf("AWS error %s", awsErr)
		return nil // we skip the access denied errors, but notify on them
	}

	return err
}
