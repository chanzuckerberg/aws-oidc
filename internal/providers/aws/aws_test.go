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
	// sourceAttached are the managed policy ARNs the source role has.
	sourceAttached []string
	// createdRoleName and attachedToAgent capture what the provider did, for assertions.
	createdRoleName string
	attachedToAgent []string
}

func (f *fakeIAM) GetRole(context.Context, *iam.GetRoleInput, ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return nil, &iamtypes.NoSuchEntityException{}
}

func (f *fakeIAM) CreateRole(_ context.Context, in *iam.CreateRoleInput, _ ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	f.createdRoleName = *in.RoleName
	arn := "arn:aws:iam::111111111111:role/agents/" + *in.RoleName
	return &iam.CreateRoleOutput{Role: &iamtypes.Role{Arn: &arn, RoleName: in.RoleName}}, nil
}

func (f *fakeIAM) DeleteRole(context.Context, *iam.DeleteRoleInput, ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, f.deleteErr
}

func (f *fakeIAM) ListAttachedRolePolicies(context.Context, *iam.ListAttachedRolePoliciesInput, ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	out := &iam.ListAttachedRolePoliciesOutput{}
	for _, arn := range f.sourceAttached {
		out.AttachedPolicies = append(out.AttachedPolicies, iamtypes.AttachedPolicy{PolicyArn: strPtr(arn)})
	}
	return out, nil
}

func (f *fakeIAM) AttachRolePolicy(_ context.Context, in *iam.AttachRolePolicyInput, _ ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	f.attachedToAgent = append(f.attachedToAgent, *in.PolicyArn)
	return &iam.AttachRolePolicyOutput{}, nil
}

func (f *fakeIAM) ListRolePolicies(context.Context, *iam.ListRolePoliciesInput, ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error) {
	return &iam.ListRolePoliciesOutput{}, nil
}

func (f *fakeIAM) GetRolePolicy(context.Context, *iam.GetRolePolicyInput, ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error) {
	return &iam.GetRolePolicyOutput{}, nil
}

func (f *fakeIAM) PutRolePolicy(context.Context, *iam.PutRolePolicyInput, ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error) {
	return &iam.PutRolePolicyOutput{}, nil
}

func strPtr(s string) *string { return &s }

func providerWithFake(f *fakeIAM) *Provider {
	factory := func(context.Context, string) (IAMAPI, error) { return f, nil }
	return NewProvider(Config{OktaAppClientID: "aud"}, factory)
}

func testProvider(deleteErr error) *Provider {
	return providerWithFake(&fakeIAM{deleteErr: deleteErr})
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

func TestEnsureNamingAndMirror(t *testing.T) {
	ctx := context.Background()
	f := &fakeIAM{sourceAttached: []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"}}
	p := providerWithFake(f)

	agent := &agentsv1.Agent{}
	agent.Name = "playground-readonly"
	agent.Spec.Owner = "00uSUB123"
	agent.Spec.OwnerEmail = "jheath@chanzuckerberg.com"
	grant := agentsv1.Grant{AWS: &agentsv1.AWSGrant{
		AccountID: "111111111111",
		RoleARN:   "arn:aws:iam::111111111111:role/readonly",
		RoleName:  "readonly",
	}}

	status, err := p.Ensure(ctx, agent, grant)
	require.NoError(t, err)
	require.Equal(t, agentsv1.GrantStateProvisioned, status.State)

	// Role name carries the owner shortname.
	require.Equal(t, "jheath-agent-playground-readonly-readonly", f.createdRoleName)
	// The source role's managed policy is mirrored onto the agent role.
	require.Equal(t, []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"}, f.attachedToAgent)
}

func TestUnreachableAccount(t *testing.T) {
	require.True(t, unreachableAccount(assumeRoleFailure()))
	require.False(t, unreachableAccount(errors.New("throttled")))
	require.False(t, unreachableAccount(&smithy.OperationError{ServiceID: "IAM", OperationName: "DeleteRole", Err: errors.New("boom")}))
}
