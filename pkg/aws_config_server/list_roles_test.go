package aws_config_server

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
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

// func TestListRoles(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)
// 	ctrl := gomock.NewController(t)

// 	client := &cziAWS.Client{}
// 	_, mock := client.WithMockIAM(ctrl)

// 	policyData, _ := json.Marshal(samplePolicyDocument)
// 	policyStr := url.PathEscape(string(policyData))

// 	testRoles1[0].AssumeRolePolicyDocument = &policyStr

// 	mock.EXPECT().
// 		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
// 		func(
// 			ctx context.Context,
// 			input *iam.ListRolesInput,
// 			accumulatorFunc func(*iam.ListRolesOutput, bool) bool,
// 		) error {
// 			accumulatorFunc(&iam.ListRolesOutput{Roles: testRoles1}, true)
// 			return nil
// 		},
// 	)

// 	mock.EXPECT().
// 		ListRoleTagsWithContext(
// 			gomock.Any(),
// 			&iam.ListRoleTagsInput{RoleName: testRoles1[0].RoleName}).
// 		Return(&iam.ListRoleTagsOutput{
// 			Tags: testRoles1[0].Tags,
// 		}, nil)

// 	mock.EXPECT().
// 		ListRoleTagsWithContext(
// 			gomock.Any(),
// 			&iam.ListRoleTagsInput{RoleName: testRoles1[1].RoleName}).
// 		Return(&iam.ListRoleTagsOutput{
// 			Tags: testRoles1[1].Tags,
// 		}, nil)

// 	iamOutput, err := listRoles(ctx, mock, &testAWSConfigGenerationParams)
// 	r.NoError(err)
// 	r.Len(testRoles1, 2) // we skipped over a role
// 	r.Len(iamOutput, 1)
// 	r.Equal(*iamOutput[0].RoleName, *testRoles1[0].RoleName)
// }

// func TestClientRoleMapFromProfile(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)

// 	newPolicyDocument := &PolicyDocument{}
// 	newPolicyDocument.Statements = append(samplePolicyDocument.Statements, invalidPolicyStatements.Statements...)

// 	newPolicyData, err := json.Marshal(newPolicyDocument)
// 	r.NoError(err)

// 	newPolicyStr := url.PathEscape(string(newPolicyData))

// 	testRoles1[0].AssumeRolePolicyDocument = &newPolicyStr

// 	clientRoleMap, err := getRoleMappings(ctx, "accountName", "accountAlias", testRoles1, oidcProvider)
// 	r.NoError(err)                                                 // Nothing weird happened
// 	r.NotEmpty(clientRoleMap)                                      // There are valid clientIDs
// 	r.Contains(clientRoleMap, okta.ClientID("clientIDValue1"))     // Only the valid ID is present
// 	r.Len(clientRoleMap, 1)                                        // No more got added
// 	r.NotContains(clientRoleMap, okta.ClientID("invalidClientID")) // none of the invalid policies (where clientID = invalidClientID) got added

// 	// See if we can handle different policy statements (2 allows)
// 	newPolicyDocument.Statements = validPolicyStatements

// 	newPolicyData, err = json.Marshal(newPolicyDocument)
// 	r.NoError(err)
// 	newPolicyStr = url.PathEscape(string(newPolicyData))
// 	testRoles2[0].AssumeRolePolicyDocument = &newPolicyStr
// }

// func TestNoPolicyDocument(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)

// 	clientRoleMap, err := getRoleMappings(ctx, "accountName", "accountAlias", testRoles0, oidcProvider)
// 	r.NoError(err)
// 	r.Empty(clientRoleMap)
// }

// func TestGetActiveAccountList(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)
// 	ctrl := gomock.NewController(t)

// 	client := &cziAWS.Client{}

// 	_, mock := client.WithMockOrganizations(ctrl)

// 	mock.EXPECT().
// 		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
// 		func(
// 			ctx context.Context,
// 			input *organizations.ListAccountsInput,
// 			accumulatorFunc func(*organizations.ListAccountsOutput, bool) bool,
// 		) error {
// 			accumulatorFunc(&organizations.ListAccountsOutput{
// 				Accounts: []*organizations.Account{
// 					{
// 						Name:   aws.String("Account1"),
// 						Status: aws.String("ACTIVE"),
// 					},
// 					{
// 						Name:   aws.String("Account2"),
// 						Status: aws.String("INACTIVE"),
// 					},
// 				},
// 			}, true)
// 			return nil
// 		},
// 	)

// 	acctList, err := GetActiveAccountList(ctx, mock)
// 	r.NoError(err)
// 	r.NotEmpty(acctList)
// 	r.Len(acctList, 1)
// 	r.Equal(*acctList[0].Name, "Account1") // the active account
// }

// func TestGetAcctAlias(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)
// 	ctrl := gomock.NewController(t)

// 	client := &cziAWS.Client{}
// 	_, mock := client.WithMockIAM(ctrl)

// 	testAlias := "account_alias_1"

// 	mock.EXPECT().
// 		ListAccountAliases(gomock.Any()).Return(
// 		&iam.ListAccountAliasesOutput{AccountAliases: []*string{&testAlias}}, nil,
// 	)

// 	outputString, err := getAcctAlias(ctx, mock)
// 	r.NoError(err)
// 	r.Equal(testAlias, outputString)
// }

// func TestGetAcctAliasNoAlias(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)
// 	ctrl := gomock.NewController(t)

// 	client := &cziAWS.Client{}
// 	_, mock := client.WithMockIAM(ctrl)

// 	mock.EXPECT().
// 		ListAccountAliases(gomock.Any()).Return(
// 		&iam.ListAccountAliasesOutput{AccountAliases: []*string{}}, nil,
// 	)

// 	outputString, err := getAcctAlias(ctx, mock)
// 	r.NoError(err)
// 	r.Equal("", outputString)
// }

// func TestParallelization(t *testing.T) {
// 	ctx := context.Background()
// 	r := require.New(t)
// 	ctrl := gomock.NewController(t)

// 	client := &cziAWS.Client{}
// 	_, mock := client.WithMockIAM(ctrl)

// 	policyData, _ := json.Marshal(samplePolicyDocument)
// 	policyStr := url.PathEscape(string(policyData))

// 	testRoles1[0].AssumeRolePolicyDocument = &policyStr

// 	mock.EXPECT().
// 		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
// 		func(
// 			ctx context.Context,
// 			input *iam.ListRolesInput,
// 			accumulatorFunc func(*iam.ListRolesOutput, bool) bool,
// 		) error {
// 			accumulatorFunc(&iam.ListRolesOutput{Roles: testRoles1}, true)
// 			return nil
// 		},
// 	).AnyTimes()

// 	mock.EXPECT().
// 		ListRoleTagsWithContext(
// 			gomock.Any(),
// 			&iam.ListRoleTagsInput{RoleName: testRoles1[0].RoleName}).
// 		Return(&iam.ListRoleTagsOutput{
// 			Tags: testRoles1[0].Tags,
// 		}, nil).AnyTimes()

// 	mock.EXPECT().
// 		ListRoleTagsWithContext(
// 			gomock.Any(),
// 			&iam.ListRoleTagsInput{RoleName: testRoles1[1].RoleName}).
// 		Return(&iam.ListRoleTagsOutput{
// 			Tags: testRoles1[1].Tags,
// 		}, nil).AnyTimes()

// 	cfgGeneration0Concurrency := AWSConfigGenerationParams{
// 		OIDCProvider:       "validProvider",
// 		AWSWorkerRole:      "validWorker",
// 		AWSOrgRoles:        []string{"arn:aws:iam::AccountNumber1:role/OrgRole1"},
// 		MappingConcurrency: 0,
// 		RolesConcurrency:   0,
// 	}
// 	iamOutput, err := listRoles(ctx, mock, &cfgGeneration0Concurrency)
// 	r.Error(err)
// 	r.Empty(iamOutput)

// 	cfgGeneration1Concurrency := AWSConfigGenerationParams{
// 		OIDCProvider:       "validProvider",
// 		AWSWorkerRole:      "validWorker",
// 		AWSOrgRoles:        []string{"arn:aws:iam::AccountNumber1:role/OrgRole1"},
// 		MappingConcurrency: 1,
// 		RolesConcurrency:   1,
// 	}
// 	iamOutput, err = listRoles(ctx, mock, &cfgGeneration1Concurrency)
// 	r.NoError(err)
// 	r.NotEmpty(iamOutput)

// 	cfgGeneration3Concurrency := AWSConfigGenerationParams{
// 		OIDCProvider:       "validProvider",
// 		AWSWorkerRole:      "validWorker",
// 		AWSOrgRoles:        []string{"arn:aws:iam::AccountNumber1:role/OrgRole1"},
// 		MappingConcurrency: 3,
// 		RolesConcurrency:   3,
// 	}
// 	iamOutput, err = listRoles(ctx, mock, &cfgGeneration3Concurrency)
// 	r.NoError(err)
// 	r.NotEmpty(iamOutput)
// }
