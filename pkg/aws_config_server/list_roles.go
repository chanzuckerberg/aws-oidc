package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
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
	AcctName string
	RoleARN  arn.ARN
}

const ignoreAWSError = "AccessDenied"

func listRoles(ctx context.Context, svc iamiface.IAMAPI) ([]*iam.Role, error) {
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
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == ignoreAWSError) {
			logrus.WithError(err).Error("Unable to list roles for this account")
			return output, nil
		}
		return output, errors.Wrap(err, "Error listing IAM roles")
	}
	return output, nil
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

func clientRoleMapFromProfile(ctx context.Context, acctName string, roles []*iam.Role, oidcProvider string, clientRoleMapping map[string][]ConfigProfile) error {
	identityProviderURL, err := url.Parse(oidcProvider)
	if err != nil {
		return errors.Wrap(err, "Failed to parse OIDC Provider input as an URL")
	}
	oidcProviderHostname := identityProviderURL.Hostname()
	logrus.Debugf("function: aws_config_server/list_roles.go/clientRoleMapFromProfile(), oidcProviderHostname: %s", oidcProviderHostname)

	for _, role := range roles {
		if role.AssumeRolePolicyDocument == nil {
			continue // role doesn't have an assume role policy document
		}
		// the IAM Role outputs a url-encoded policy document, so we need to escape characters
		policyStr, err := url.PathUnescape(*role.AssumeRolePolicyDocument)
		if err != nil {
			return errors.Wrap(err, "Unable to escape URL encoding")
		}
		policyDoc := PolicyDocument{}
		err = json.Unmarshal([]byte(policyStr), &policyDoc)
		if err != nil {
			return errors.Wrapf(err, "Unable to unmarshal policy document to struct \npolicy: %s", policyStr)
		}

		for _, statement := range policyDoc.Statements {
			federatedARN := statement.Principal.Federated
			if !strings.Contains(federatedARN, oidcProviderHostname) {
				continue
			}

			clientKey := fmt.Sprintf("%s:aud", oidcProviderHostname)
			clientID, ok := statement.Condition.StringEquals[clientKey]
			if !ok || (clientID == "") {
				continue
			}

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
				return errors.Wrapf(err, "could not parse arn %s", *role.Arn)
			}

			currentConfig := ConfigProfile{
				AcctName: acctName,
				RoleARN:  roleARN,
			}

			if _, ok := clientRoleMapping[clientID]; !ok {
				clientRoleMapping[clientID] = []ConfigProfile{currentConfig}
				continue
			}
			logrus.Debug("function: aws_config_server/list_roles.go/clientRoleMapFromProfile()\n About to append currentConfig to clientRoleMapping")
			clientRoleMapping[clientID] = append(clientRoleMapping[clientID], currentConfig)
		}
	}
	return nil
}

func GetActiveAccountList(ctx context.Context, svc organizationsiface.OrganizationsAPI) ([]*organizations.Account, error) {
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
	if err != nil {
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
