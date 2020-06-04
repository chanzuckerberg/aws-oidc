package aws_config_server

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestMultipleActions(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mock := client.WithMockIAM(ctrl)

	policyData, err := json.Marshal(revisedPolicyDocument)
	r.NoError(err)
	policyStr := url.PathEscape(string(policyData))

	testRoles3[0].AssumeRolePolicyDocument = &policyStr

	mock.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			ctx context.Context,
			input *iam.ListRolesInput,
			accumulatorFunc func(*iam.ListRolesOutput, bool) bool,
		) error {
			accumulatorFunc(&iam.ListRolesOutput{Roles: testRoles3}, true)
			return nil
		},
	)

	iamOutput, err := listRoles(ctx, mock)
	r.NoError(err)
	r.Len(iamOutput, 1)

	clientRoleMap := make(map[string][]ConfigProfile)
	err = clientRoleMapFromProfile(ctx, "accountName", testRoles3, oidcProvider, clientRoleMap)
	r.NoError(err)                                  // Nothing weird happened
	r.NotEmpty(clientRoleMap)                       // There are valid clientIDs
	r.Contains(clientRoleMap, "clientIDValue3")     // Only the valid ID is present
	r.Len(clientRoleMap, 1)                         // No more got added
	r.NotContains(clientRoleMap, "invalidClientID") // none of the invalid policies (where clientID = invalidClientID) got added
}

func TestSingleStringAction(t *testing.T) {
	r := require.New(t)

	policyData, err := json.Marshal(alternatePolicyDocument)
	r.NoError(err)

	policyDoc := PolicyDocument{}
	err = json.Unmarshal(policyData, &policyDoc)
	r.NoError(err)

	r.NotEmpty(policyDoc)
	r.Len(policyDoc.Statements, 1)
	r.Len(policyDoc.Statements[0].Action, 1)
	r.Equal(policyDoc.Statements[0].Action[0], "sts:AssumeRoleWithWebIdentity")
}
