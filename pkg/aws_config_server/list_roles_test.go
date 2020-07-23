package aws_config_server

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestProcessAWSError(t *testing.T) {
	r := require.New(t)

	type test struct {
		name     string
		in       error
		expected error
	}

	tests := []test{
		{
			name:     "nil error",
			in:       nil,
			expected: nil,
		},
		{
			name:     "rando error",
			in:       fmt.Errorf("rando"),
			expected: fmt.Errorf("rando"),
		},
		{
			name:     "aws error not access denied",
			in:       awserr.New("not access denied", "bla", nil),
			expected: awserr.New("not access denied", "bla", nil),
		},
		{
			name:     "aws error access denied",
			in:       awserr.New(errAWSAccessDenied, "bla", nil),
			expected: nil,
		},
	}
	for _, test := range tests {
		fmt.Println(test.name)
		r.Equal(test.expected, processAWSErr(test.in))
	}
}

func TestGetAcctAlias(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	// no alias
	mock.EXPECT().ListAccountAliasesWithContext(gomock.Any(), gomock.Any()).Return(&iam.ListAccountAliasesOutput{}, nil)
	alias, err := getAcctAlias(ctx, mock)
	r.NoError(err)
	r.Nil(alias)

	// error
	mock.EXPECT().ListAccountAliasesWithContext(gomock.Any(), gomock.Any()).Return(&iam.ListAccountAliasesOutput{}, fmt.Errorf("some error"))
	alias, err = getAcctAlias(ctx, mock)
	r.Equal(fmt.Errorf("some error"), errors.Cause(err))
	r.Nil(alias)

	// get an alias
	mock.EXPECT().ListAccountAliasesWithContext(gomock.Any(), gomock.Any()).Return(&iam.ListAccountAliasesOutput{
		AccountAliases: []*string{aws.String("alias test")},
	}, nil)
	alias, err = getAcctAlias(ctx, mock)
	r.NoError(err)
	r.Equal("alias test", *alias)
}

func TestGetActiveAccountList(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockOrganizations(ctrl)

	// ignores access denied errors
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(awserr.New(errAWSAccessDenied, "ignore me", nil))

	actual, err := getActiveAccountList(ctx, mock)
	r.NoError(err)
	r.Empty(actual)

	// returns other errors
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("do not ignore"))

	actual, err = getActiveAccountList(ctx, mock)
	r.Equal(fmt.Errorf("do not ignore"), errors.Cause(err))
	r.Empty(actual)

	// adds some active, some inactive. returns only active
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				ctx context.Context,
				input *organizations.ListAccountsInput,
				accumulator func(*organizations.ListAccountsOutput, bool) bool,
			) error {
				accumulator(&organizations.ListAccountsOutput{
					Accounts: []*organizations.Account{
						{
							Status: aws.String("ACTIVE"),
						},
						{
							Status: aws.String("INACTIVE"),
						},
					},
				}, true)
				return nil
			},
		)
	actual, err = getActiveAccountList(ctx, mock)
	r.NoError(err)
	r.NotEmpty(actual)
	r.Len(actual, 1)
}

func TestListRoles(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	// ignores access denied errors
	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(awserr.New(errAWSAccessDenied, "ignore me", nil))

	actual, err := listRoles(ctx, mock)
	r.NoError(err)
	r.Empty(actual)

	// returns other errors
	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("do not ignore"))

	actual, err = listRoles(ctx, mock)
	r.Equal(fmt.Errorf("do not ignore"), errors.Cause(err))
	r.Empty(actual)

	// adds some active, some inactive. returns only active
	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				ctx context.Context,
				input *iam.ListRolesInput,
				accumulator func(*iam.ListRolesOutput, bool) bool,
			) error {
				accumulator(&iam.ListRolesOutput{
					Roles: []*iam.Role{
						{}, {}, {}, // 3 roles
					},
				}, true)
				return nil
			},
		)
	actual, err = listRoles(ctx, mock)
	r.NoError(err)
	r.NotEmpty(actual)
	r.Len(actual, 3)
}

func TestListRolesForAccountsNoWorkerRoles(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	roleAssumer := func(config *aws.Config) iamiface.IAMAPI {
		return mock
	}

	sess, server := cziAWS.NewMockSession()
	defer server.Close()

	workerRoles := []workerRole{}
	oidcProvider := "foo-provider"

	// no worker roles
	federatedRoles, err := listRolesForAccounts(ctx, sess, roleAssumer, workerRoles, oidcProvider, 10)
	r.NoError(err)
	r.Empty(federatedRoles)
}

func TestListRolesForAccountsNoRolesFound(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	roleAssumer := func(config *aws.Config) iamiface.IAMAPI {
		return mock
	}

	sess, server := cziAWS.NewMockSession()
	defer server.Close()

	workerRoles := []workerRole{}
	oidcProvider := "foo-provider"

	mock.EXPECT().ListAccountAliasesWithContext(gomock.Any(), gomock.Any()).Return(nil, nil) // no alias

	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				ctx context.Context,
				input *iam.ListRolesInput,
				accumulator func(*iam.ListRolesOutput, bool) bool,
			) error {
				accumulator(&iam.ListRolesOutput{
					Roles: []*iam.Role{},
				}, true)
				return nil
			},
		)
	federatedRoles, err := listRolesForAccounts(ctx, sess, roleAssumer, workerRoles, oidcProvider, 10)
	r.NoError(err)
	r.Empty(federatedRoles)
}

func TestListRolesForAccountsRolesFound(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	roleAssumer := func(config *aws.Config) iamiface.IAMAPI {
		return mock
	}

	sess, server := cziAWS.NewMockSession()
	defer server.Close()

	workerRoles := []workerRole{
		{
			role: &arn.ARN{
				Partition: "aws",
				Service:   "iam",
				AccountID: "1234567891012",
				Resource:  "role/foobar",
			},
		},
	}
	oidcProvider := "https://localhost"

	roles := []*iam.Role{
		// want to show up at the end
		{
			RoleName:                 aws.String("foo"),
			Arn:                      aws.String("arn:aws:iam:::role/foo"),
			AssumeRolePolicyDocument: policyDocumentToString(revisedPolicyDocument),
		},
		// skip tags, skip
		{
			RoleName:                 aws.String("bar"),
			Arn:                      aws.String("arn:aws:iam:::role/bar"),
			AssumeRolePolicyDocument: policyDocumentToString(revisedPolicyDocument),
		},
		// no assume role policy, skip
		{
			RoleName: aws.String("baz"),
			Arn:      aws.String("arn:aws:iam:::role/baz"),
		},
	}

	mock.EXPECT().ListAccountAliasesWithContext(gomock.Any(), gomock.Any()).Return(nil, nil) // no alias

	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				ctx context.Context,
				input *iam.ListRolesInput,
				accumulator func(*iam.ListRolesOutput, bool) bool,
			) error {
				accumulator(&iam.ListRolesOutput{
					Roles: roles,
				}, true)
				return nil
			},
		)

		// skip the bar role
	mock.EXPECT().
		ListRoleTagsWithContext(
			gomock.Any(),
			gomock.Eq(&iam.ListRoleTagsInput{RoleName: aws.String("bar")})).
		Return(&iam.ListRoleTagsOutput{
			Tags: []*iam.Tag{
				{
					Key:   aws.String(skipRolesTagKey),
					Value: aws.String("does not matter"),
				},
			},
		}, nil)

		// keep the foo role (no tags)
	mock.EXPECT().
		ListRoleTagsWithContext(
			gomock.Any(),
			gomock.Eq(&iam.ListRoleTagsInput{RoleName: aws.String("foo")})).
		Return(&iam.ListRoleTagsOutput{
			Tags: []*iam.Tag{},
		}, nil)

	federatedRoles, err := listRolesForAccounts(ctx, sess, roleAssumer, workerRoles, oidcProvider, 10)
	r.NoError(err)

	r.Equal(&oidcFederatedRoles{
		roles: map[okta.ClientID][]accountAndRole{
			"clientIDValue3": {
				{
					AccountName:  "",
					AccountAlias: nil,
					RoleARN: &arn.ARN{
						Partition: "aws",
						Service:   "iam",
						Region:    "",
						AccountID: "",
						Resource:  "role/foo",
					},
					Role: roles[0],
				},
			},
		},
	}, federatedRoles)
}

func TestGetWorkerRoles(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockOrganizations(ctrl)

	sess, server := cziAWS.NewMockSession()
	defer server.Close()

	orgAssumer := func(config *aws.Config) organizationsiface.OrganizationsAPI {
		return mock
	}

	orgRoles := []string{"first"}
	workerRoleName := "foobar"

	// ignores access denied errors
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(awserr.New(errAWSAccessDenied, "ignore me", nil))

	actual, err := getWorkerRoles(ctx, sess, orgAssumer, orgRoles, workerRoleName)
	r.NoError(err)
	r.Empty(actual)

	// returns other errors
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("do not ignore"))

	actual, err = getWorkerRoles(ctx, sess, orgAssumer, orgRoles, workerRoleName)
	r.Equal(fmt.Errorf("do not ignore"), errors.Cause(err))
	r.Empty(actual)

	// adds some active, some inactive. returns only active
	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				ctx context.Context,
				input *organizations.ListAccountsInput,
				accumulator func(*organizations.ListAccountsOutput, bool) bool,
			) error {
				accumulator(&organizations.ListAccountsOutput{
					Accounts: []*organizations.Account{
						{
							Status: aws.String("ACTIVE"),
							Id:     aws.String("0123456789"),
							Name:   aws.String("active"),
						},
						{
							Status: aws.String("INACTIVE"),
							Id:     aws.String("0000000000000000"),
							Name:   aws.String("inactive"),
						},
					},
				}, true)
				return nil
			},
		)
	actual, err = getWorkerRoles(ctx, sess, orgAssumer, orgRoles, workerRoleName)
	r.NoError(err)
	r.Equal([]workerRole{
		{
			role: &arn.ARN{
				Partition: "aws",
				Service:   "iam",
				AccountID: "0123456789",
				Resource:  "role/foobar",
			},
			accountName: "active",
		},
	}, actual)
}
