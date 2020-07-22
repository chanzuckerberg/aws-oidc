package aws_config_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestListRoles(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	policyData, _ := json.Marshal(samplePolicyDocument)
	policyStr := url.PathEscape(string(policyData))

	testRoles1[0].AssumeRolePolicyDocument = &policyStr

	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			ctx context.Context,
			input *iam.ListRolesInput,
			accumulatorFunc func(*iam.ListRolesOutput, bool) bool,
		) error {
			accumulatorFunc(&iam.ListRolesOutput{Roles: testRoles1}, true)
			return nil
		},
	)

	mock.EXPECT().
		ListRoleTagsWithContext(
			gomock.Any(),
			&iam.ListRoleTagsInput{RoleName: testRoles1[0].RoleName}).
		Return(&iam.ListRoleTagsOutput{
			Tags: testRoles1[0].Tags,
		}, nil)

	mock.EXPECT().
		ListRoleTagsWithContext(
			gomock.Any(),
			&iam.ListRoleTagsInput{RoleName: testRoles1[1].RoleName}).
		Return(&iam.ListRoleTagsOutput{
			Tags: testRoles1[1].Tags,
		}, nil)

	iamOutput, err := listRoles(ctx, mock, &testAWSConfigGenerationParams)
	r.NoError(err)
	r.Equal(*iamOutput[0].RoleName, *testRoles1[0].RoleName)
}

func TestClientRoleMapFromProfile(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	newPolicyDocument := &PolicyDocument{}
	newPolicyDocument.Statements = append(samplePolicyDocument.Statements, invalidPolicyStatements.Statements...)

	newPolicyData, err := json.Marshal(newPolicyDocument)
	r.NoError(err)

	newPolicyStr := url.PathEscape(string(newPolicyData))

	testRoles1[0].AssumeRolePolicyDocument = &newPolicyStr

	clientRoleMap, err := getRoleMappings(ctx, "accountName", "accountAlias", testRoles1, oidcProvider)
	r.NoError(err)                                                 // Nothing weird happened
	r.NotEmpty(clientRoleMap)                                      // There are valid clientIDs
	r.Contains(clientRoleMap, okta.ClientID("clientIDValue1"))     // Only the valid ID is present
	r.Len(clientRoleMap, 1)                                        // No more got added
	r.NotContains(clientRoleMap, okta.ClientID("invalidClientID")) // none of the invalid policies (where clientID = invalidClientID) got added

	// See if we can handle different policy statements (2 allows)
	newPolicyDocument.Statements = validPolicyStatements

	newPolicyData, err = json.Marshal(newPolicyDocument)
	r.NoError(err)
	newPolicyStr = url.PathEscape(string(newPolicyData))
	testRoles2[0].AssumeRolePolicyDocument = &newPolicyStr
}

func TestNoPolicyDocument(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	clientRoleMap, err := getRoleMappings(ctx, "accountName", "accountAlias", testRoles0, oidcProvider)
	r.NoError(err)
	r.Empty(clientRoleMap)
}

func TestGetActiveAccountList(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}

	_, mock := client.WithMockOrganizations(ctrl)

	mock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			ctx context.Context,
			input *organizations.ListAccountsInput,
			accumulatorFunc func(*organizations.ListAccountsOutput, bool) bool,
		) error {
			accumulatorFunc(&organizations.ListAccountsOutput{
				Accounts: []*organizations.Account{
					{
						Name:   aws.String("Account1"),
						Status: aws.String("ACTIVE"),
					},
					{
						Name:   aws.String("Account2"),
						Status: aws.String("INACTIVE"),
					},
				},
			}, true)
			return nil
		},
	)

	acctList, err := GetActiveAccountList(ctx, mock)
	r.NoError(err)
	r.NotEmpty(acctList)
	r.Len(acctList, 1)
	r.Equal(*acctList[0].Name, "Account1") // the active account
}

func TestGetAcctAlias(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	testAlias := "account_alias_1"

	mock.EXPECT().
		ListAccountAliases(gomock.Any()).Return(
		&iam.ListAccountAliasesOutput{AccountAliases: []*string{&testAlias}}, nil,
	)

	outputString, err := getAcctAlias(ctx, mock)
	r.NoError(err)
	r.Equal(testAlias, outputString)
}

func TestGetAcctAliasNoAlias(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	mock.EXPECT().
		ListAccountAliases(gomock.Any()).Return(
		&iam.ListAccountAliasesOutput{AccountAliases: []*string{}}, nil,
	)

	outputString, err := getAcctAlias(ctx, mock)
	r.NoError(err)
	r.Equal("", outputString)
}

func TestGetWorkerRoles(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	// _, iamMock := client.WithMockIAM(ctrl)
	_, orgMock := client.WithMockOrganizations(ctrl)

	policyData, _ := json.Marshal(samplePolicyDocument)
	policyStr := url.PathEscape(string(policyData))

	testRoles1[0].AssumeRolePolicyDocument = &policyStr

	orgMock.EXPECT().
		ListAccountsPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			ctx context.Context,
			input *organizations.ListAccountsInput,
			accumulatorFunc func(*organizations.ListAccountsOutput, bool) bool,
		) error {
			accumulatorFunc(&organizations.ListAccountsOutput{
				Accounts: []*organizations.Account{
					{
						Name:   aws.String("Account1"),
						Status: aws.String("ACTIVE"),
					},
					{
						Name:   aws.String("Account2"),
						Status: aws.String("INACTIVE"),
					},
				},
			}, true)
			return nil
		},
	).AnyTimes()

	mockSess, mockServer := cziAWS.NewMockSession()
	defer mockServer.Close()

	a := ClientIDToAWSRoles{
		awsSession:        mockSess,
		roleARNs:          map[string]arn.ARN{},
		clientRoleMapping: map[okta.ClientID][]ConfigProfile{},
		awsClient:         cziAWS.New(mockSess),
	}
	err := a.getWorkerRoles(ctx, testAWSConfigGenerationParams.AWSOrgRoles, testAWSConfigGenerationParams.AWSWorkerRole)
	r.NoError(err)
	r.NotEmpty(a.roleARNs)
}

func TestPopulateMapping(t *testing.T) {
	// Includes parallelization
	ctx := context.Background()
	r := require.New(t)
	// ctrl := gomock.NewController(t)

	// client := &cziAWS.Client{}
	// _, iamMock := client.WithMockIAM(ctrl)
	// _, orgMock := client.WithMockOrganizations(ctrl)

	policyData, _ := json.Marshal(samplePolicyDocument)
	policyStr := url.PathEscape(string(policyData))

	testRoles1[0].AssumeRolePolicyDocument = &policyStr

	mockSess, mockServer := cziAWS.NewMockSession()
	defer mockServer.Close()

	a := ClientIDToAWSRoles{
		awsSession:        mockSess,
		roleARNs:          map[string]arn.ARN{},
		clientRoleMapping: map[okta.ClientID][]ConfigProfile{},
		awsClient:         cziAWS.New(mockSess),
	}
	a.roleARNs["acct1"] = arn.ARN{
		Partition: "aws",
		Service:   "iam",
		Resource:  fmt.Sprintf("role/%s", *testRoles1[0].RoleName),
	}
	a.roleARNs["acct2"] = arn.ARN{
		Partition: "aws",
		Service:   "iam",
		Resource:  fmt.Sprintf("role/%s", *testRoles1[1].RoleName),
	}

	testAWSConfigGenerationParams.MappingConcurrency = 0
	testAWSConfigGenerationParams.RolesConcurrency = 0
	err := a.populateMapping(ctx, &testAWSConfigGenerationParams)
	r.Error(err)

	testAWSConfigGenerationParams.MappingConcurrency = 1
	testAWSConfigGenerationParams.RolesConcurrency = 1

	testAWSConfigGenerationParams.MappingConcurrency = 3
	testAWSConfigGenerationParams.RolesConcurrency = 3
}
