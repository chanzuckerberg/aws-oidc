package aws_config_server

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/hashicorp/go-multierror"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountAndRole struct {
	AccountName  string
	AccountAlias *string

	RoleARN *arn.ARN
	Role    *iam.Role
}

type workerRole struct {
	role        *arn.ARN
	accountName string
}

type awsOrgRoleAssumer func(*aws.Config) organizationsiface.OrganizationsAPI
type awsIAMRoleAssumer func(*aws.Config) iamiface.IAMAPI

// getWorkerRoles gets the roles for each active account in the organizations provided
func getWorkerRoles(
	ctx context.Context,
	session *session.Session,
	awsOrgRoleAssumer awsOrgRoleAssumer,
	orgRoles []string,
	workerRoleName string) ([]workerRole, error) {
	ctx, span := beeline.StartSpan(ctx, "server_get_worker_roles")
	defer span.Send()

	workerRoles := []workerRole{}
	for _, role_arn := range orgRoles {
		orgAWSConfig := &aws.Config{
			Credentials:                   stscreds.NewCredentials(session, role_arn),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Retryer:                       session.Config.Retryer,
		}

		orgClient := awsOrgRoleAssumer(orgAWSConfig)
		accountList, err := getActiveAccountList(ctx, orgClient)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get list of AWS Profiles")
		}
		for _, acct := range accountList {
			// create a new IAM session for each account
			roleARN := &arn.ARN{
				Partition: "aws",
				Service:   "iam",
				AccountID: *acct.Id,
				Resource:  fmt.Sprintf("role/%s", workerRoleName),
			}
			workerRoles = append(workerRoles, workerRole{
				accountName: *acct.Name,
				role:        roleARN,
			})
		}
	}
	return workerRoles, nil
}

func processAccountRoles(
	ctx context.Context,
	iamClient iamiface.IAMAPI,
	oidcProvider string,
	accountName string,
) (*oidcFederatedRoles, error) {
	alias, err := getAcctAlias(ctx, iamClient)
	if err != nil {
		return nil, err
	}
	roles, err := listRoles(ctx, iamClient)
	if err != nil {
		return nil, err
	}

	accountAndRoles := []accountAndRole{}
	for _, role := range roles {
		roleARN, err := arn.Parse(*role.Arn)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse arn %s", *role.Arn)
		}
		accountAndRoles = append(accountAndRoles, accountAndRole{
			AccountName:  accountName,
			AccountAlias: alias,

			Role:    role,
			RoleARN: &roleARN,
		})
	}

	federatedRoles, err := getOIDCFederatedRoles(ctx, oidcProvider, accountAndRoles)
	if err != nil {
		return nil, err
	}

	return filterOIDCFederatedRoles(ctx, iamClient, federatedRoles)
}

// listRolesForAccounts will get all roles for all accounts
func listRolesForAccounts(
	ctx context.Context,
	session *session.Session,
	awsIAMRoleAssumer awsIAMRoleAssumer,
	workerRoles []workerRole,
	oidcProvider string,
	concurrency int,
) (*oidcFederatedRoles, error) {
	wg := sync.WaitGroup{}
	errs := make(chan error, len(workerRoles))
	queue := make(chan workerRole, len(workerRoles))
	output := make(chan *oidcFederatedRoles, len(workerRoles))

	processor := func() {
		defer wg.Done()
		for element := range queue {
			awsConfig := &aws.Config{
				Credentials:                   stscreds.NewCredentials(session, element.role.String()),
				CredentialsChainVerboseErrors: aws.Bool(true),
				Retryer:                       session.Config.Retryer,
			}
			iamClient := awsIAMRoleAssumer(awsConfig)

			federatedRoles, err := processAccountRoles(ctx, iamClient, oidcProvider, element.accountName)
			if err != nil {
				errs <- err
				continue
			}
			output <- federatedRoles
		}
	}

	for _, workerRole := range workerRoles {
		queue <- workerRole
	}
	close(queue)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go processor()
	}
	wg.Wait()
	close(errs)
	close(output)

	allErrs := &multierror.Error{}
	for err := range errs {
		allErrs = multierror.Append(allErrs, err)
	}
	logrus.Debug("added errors to error channel")

	// small lesson: don't set a length for the outputList!
	// Or else we'll get nil pointers in the output,
	// 	which will cause segmentation violations
	allOIDCRoles := &oidcFederatedRoles{}
	for federatedRoles := range output {
		allOIDCRoles.Merge(*federatedRoles)
	}

	return allOIDCRoles, allErrs.ErrorOrNil()
}

// listRoleTags lists the tags for a role
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

// listRoles will list all roles in an account
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
	return output, errors.Wrap(processAWSErr(err), "could not list IAM roles")
}

func getActiveAccountList(
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

func getAcctAlias(ctx context.Context, svc iamiface.IAMAPI) (*string, error) {
	input := &iam.ListAccountAliasesInput{}
	output, err := svc.ListAccountAliasesWithContext(ctx, input)
	if processAWSErr(err) != nil {
		return nil, errors.Wrap(err, "could not get account alias")
	}

	// no alias
	if output == nil || len(output.AccountAliases) == 0 {
		return nil, nil
	}
	// NOTE: according to AWS docs can only be one alias on an account
	return output.AccountAliases[0], nil
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
		logrus.Errorf("AWS error %s", awsErr.Message())
		return nil // we skip the access denied errors, but notify on them
	}

	return err
}
