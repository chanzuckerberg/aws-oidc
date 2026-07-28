package aws

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

type fakeIAM struct {
	deleteErr error
}

func (f *fakeIAM) GetRole(context.Context, *iam.GetRoleInput, ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return nil, &iamtypes.NoSuchEntityException{}
}

func (f *fakeIAM) CreateRole(context.Context, *iam.CreateRoleInput, ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return &iam.CreateRoleOutput{}, nil
}

func (f *fakeIAM) DeleteRole(context.Context, *iam.DeleteRoleInput, ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, f.deleteErr
}

func testProvider(deleteErr error) *Provider {
	factory := func(context.Context, string) (IAMAPI, error) {
		return &fakeIAM{deleteErr: deleteErr}, nil
	}
	return NewProvider(Config{OktaAppClientID: "aud"}, factory)
}

func awsGrant() agentsv1.Grant {
	return agentsv1.Grant{AWS: &agentsv1.AWSGrant{
		AccountID: "111111111111",
		RoleARN:   "arn:aws:iam::111111111111:role/readonly",
		RoleName:  "readonly",
	}}
}

// The SDK surfaces a failed cross-account assume as an STS operation error nested under the
// IAM operation error, mirroring what DeleteRole returns when the provisioner role cannot be
// assumed.
func assumeRoleFailure() error {
	sts := &smithy.OperationError{ServiceID: "STS", OperationName: "AssumeRole", Err: errors.New("AccessDenied")}
	return &smithy.OperationError{
		ServiceID:     "IAM",
		OperationName: "DeleteRole",
		Err:           fmt.Errorf("get identity: get credentials: failed to refresh cached credentials: %w", sts),
	}
}

func TestDeleteBestEffort(t *testing.T) {
	ctx := context.Background()
	agent := &agentsv1.Agent{}
	agent.Name = "bot"
	grant := awsGrant()

	// Role already gone: success.
	err := testProvider(&iamtypes.NoSuchEntityException{}).Delete(ctx, agent, grant)
	require.NoError(t, err)

	// Cannot assume into the account: treated as nothing to clean up, so the finalizer is
	// not wedged.
	err = testProvider(assumeRoleFailure()).Delete(ctx, agent, grant)
	require.NoError(t, err)

	// A genuine IAM error still surfaces so it can be retried.
	err = testProvider(errors.New("throttled")).Delete(ctx, agent, grant)
	require.Error(t, err)
}

func TestUnreachableAccount(t *testing.T) {
	require.True(t, unreachableAccount(assumeRoleFailure()))
	require.False(t, unreachableAccount(errors.New("throttled")))
	require.False(t, unreachableAccount(&smithy.OperationError{ServiceID: "IAM", OperationName: "DeleteRole", Err: errors.New("boom")}))
}
