package aws_config_server

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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

	iamOutput, err := listRoles(ctx, mock)
	r.NoError(err)
	r.Len(testRoles1, 2) // we skipped over a role
	r.Len(iamOutput, 1)
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

	clientRoleMap := map[okta.ClientID][]ConfigProfile{}
	err = clientRoleMapFromProfile(ctx, "accountName", "accountAlias", testRoles1, oidcProvider, clientRoleMap)
	r.NoError(err)                                                 // Nothing weird happened
	r.NotEmpty(clientRoleMap)                                      // There are valid clientIDs
	r.Contains(clientRoleMap, okta.ClientID("clientIDValue1"))     // Only the valid ID is present
	r.Len(clientRoleMap, 1)                                        // No more got added
	r.NotContains(clientRoleMap, okta.ClientID("invalidClientID")) // none of the invalid policies (where clientID = invalidClientID) got added

	// See if we can:
	// * append another ARN to the clientRoleMap variable
	// * include a new clientID to the existing clientRoleMap
	newPolicyDocument.Statements = validPolicyStatements

	newPolicyData, err = json.Marshal(newPolicyDocument)
	r.NoError(err)
	newPolicyStr = url.PathEscape(string(newPolicyData))
	testRoles2[0].AssumeRolePolicyDocument = &newPolicyStr

	err = clientRoleMapFromProfile(ctx, "accountName2", "accountAlias2", testRoles2, oidcProvider, clientRoleMap)

	r.NoError(err)
	r.Len(clientRoleMap[okta.ClientID("clientIDValue1")], 2) // original ClientID has another entry
	r.Len(clientRoleMap, 2)                                  // new changes reflected in the updated map
	r.Len(clientRoleMap[okta.ClientID("clientIDValue2")], 1) // new clientID from the new roles got added
}

func TestNoPolicyDocument(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	clientRoleMap := map[okta.ClientID][]ConfigProfile{}
	err := clientRoleMapFromProfile(ctx, "accountName", "accountAlias", testRoles0, oidcProvider, clientRoleMap)
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
