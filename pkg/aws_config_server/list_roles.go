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

// We can skip over roles with specific tags
func filterRoles(
	ctx context.Context,
	svc iamiface.IAMAPI,
	roles []*iam.Role) ([]*iam.Role, error) {
	ctx, span := beeline.StartSpan(ctx, "filtering AWS roles")
	defer span.Send()

	out := []*iam.Role{}
	shouldSkipTags := func(tags []*iam.Tag) bool {
		for _, tag := range tags {
			if tag != nil && tag.Key != nil && *tag.Key == skipRolesTagKey {
				return true
			}
		}
		return false
	}

	for _, role := range roles {
		if role == nil {
			continue
		}
		tags, err := listRoleTags(ctx, svc, role.RoleName)
		if err != nil {
			return nil, err
		}

		if shouldSkipTags(tags) {
			continue
		}

		out = append(out, role)
	}
	return out, nil
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

func listRoles(ctx context.Context, svc iamiface.IAMAPI) ([]*iam.Role, error) {
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

	return filterRoles(ctx, svc, output)
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

func clientRoleMapFromProfile(
	ctx context.Context,
	acctName string,
	acctAlias string,
	roles []*iam.Role,
	oidcProvider string,
	clientRoleMapping map[okta.ClientID][]ConfigProfile) error {
	identityProviderURL, err := url.Parse(oidcProvider)
	if err != nil {
		return errors.Wrap(err, "Failed to parse OIDC Provider input as an URL")
	}
	oidcProviderHostname := identityProviderURL.Hostname()
	logrus.Debugf("oidcProviderHostname: %s", oidcProviderHostname)

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
			return errors.Wrapf(err, "Unable to unmarshal policy document to struct policy: %s", policyStr)
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
				return errors.Wrapf(err, "could not parse arn %s", *role.Arn)
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
			logrus.Debugf("Found oidc role %s", roleARN)
			clientRoleMapping[clientID] = append(clientRoleMapping[clientID], currentConfig)
		}
	}
	return nil
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
