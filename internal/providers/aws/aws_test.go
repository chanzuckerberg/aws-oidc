package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
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
	// existingTrust, when set, makes GetRole report an existing role carrying this trust
	// document (URL-encoded, the way IAM returns it). Empty means the role does not exist.
	existingTrust string
	// createdRoleName, attachedToAgent, and detachedFromAgent capture what the provider did.
	createdRoleName   string
	createdTrust      string
	updatedTrust      string
	attachedToAgent   []string
	detachedFromAgent []string
}

func (f *fakeIAM) GetRole(_ context.Context, in *iam.GetRoleInput, _ ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	if f.existingTrust == "" {
		return nil, &iamtypes.NoSuchEntityException{}
	}
	arn := "arn:aws:iam::111111111111:role/agents/" + *in.RoleName
	return &iam.GetRoleOutput{Role: &iamtypes.Role{
		Arn:                      &arn,
		RoleName:                 in.RoleName,
		AssumeRolePolicyDocument: strPtr(url.QueryEscape(f.existingTrust)),
	}}, nil
}

func (f *fakeIAM) CreateRole(_ context.Context, in *iam.CreateRoleInput, _ ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	f.createdRoleName = *in.RoleName
	f.createdTrust = *in.AssumeRolePolicyDocument
	arn := "arn:aws:iam::111111111111:role/agents/" + *in.RoleName
	return &iam.CreateRoleOutput{Role: &iamtypes.Role{Arn: &arn, RoleName: in.RoleName}}, nil
}

func (f *fakeIAM) UpdateAssumeRolePolicy(_ context.Context, in *iam.UpdateAssumeRolePolicyInput, _ ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error) {
	f.updatedTrust = *in.PolicyDocument
	return &iam.UpdateAssumeRolePolicyOutput{}, nil
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

func (f *fakeIAM) DetachRolePolicy(_ context.Context, in *iam.DetachRolePolicyInput, _ ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	f.detachedFromAgent = append(f.detachedFromAgent, *in.PolicyArn)
	return &iam.DetachRolePolicyOutput{}, nil
}

func (f *fakeIAM) DeleteRolePolicy(context.Context, *iam.DeleteRolePolicyInput, ...func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error) {
	return &iam.DeleteRolePolicyOutput{}, nil
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

func TestTagsMergeStandardAndAgent(t *testing.T) {
	p := NewProvider(Config{
		DefaultTags: map[string]string{"project": "agent-registry", "env": "rdev", "service": "aws-oidc"},
	}, func(context.Context, string) (IAMAPI, error) { return &fakeIAM{}, nil })

	agent := &agentsv1.Agent{}
	agent.Name = "playground-readonly"
	agent.Spec.Owner = "00uSUB123"
	agent.Spec.OwnerEmail = "jheath@chanzuckerberg.com"

	got := map[string]string{}
	for _, tag := range p.tags(agent) {
		got[*tag.Key] = *tag.Value
	}

	require.Equal(t, map[string]string{
		"project":     "agent-registry",
		"env":         "rdev",
		"service":     "aws-oidc",
		"managedBy":   "aws-oidc-agent-operator",
		"owner":       "jheath@chanzuckerberg.com",
		"agent-name":  "playground-readonly",
		"agent-owner": "00uSUB123",
	}, got)
}

func TestDeleteDetachesPoliciesFirst(t *testing.T) {
	ctx := context.Background()
	// The agent role has a managed policy attached (mirrored), so delete must detach it
	// before DeleteRole, which otherwise fails with DeleteConflict.
	f := &fakeIAM{sourceAttached: []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"}}
	p := providerWithFake(f)

	agent := &agentsv1.Agent{}
	agent.Name = "playground-readonly"
	agent.Spec.OwnerEmail = "jheath@chanzuckerberg.com"

	err := p.Delete(ctx, agent, awsGrant())
	require.NoError(t, err)
	require.Equal(t, []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"}, f.detachedFromAgent)
}

// clusterProvider mirrors an EKS cluster's OIDC issuer host and path.
const clusterProvider = "oidc.eks.us-west-2.amazonaws.com/id/EXAMPLE"

func runtimeAgent() *agentsv1.Agent {
	agent := &agentsv1.Agent{}
	agent.Name = "bot"
	agent.UID = "0f8fad5b-d9cb-469f-a165-70867728950e"
	agent.Spec.Owner = "00uSUB123"
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}
	return agent
}

func clusterProviderWithFake(f *fakeIAM) *Provider {
	return NewProvider(Config{
		OktaAppClientID:     "aud",
		ClusterOIDCProvider: clusterProvider,
		PodNamespace:        "argus-aws-oidc-rdev",
	}, func(context.Context, string) (IAMAPI, error) { return f, nil })
}

// trustStatements pulls the statements out of a trust document, keyed by Sid.
func trustStatements(t *testing.T, doc string) map[string]map[string]any {
	t.Helper()
	var parsed struct {
		Statement []map[string]any `json:"Statement"`
	}
	require.NoError(t, json.Unmarshal([]byte(doc), &parsed))

	byStatementID := make(map[string]map[string]any, len(parsed.Statement))
	for _, statement := range parsed.Statement {
		sid, _ := statement["Sid"].(string)
		byStatementID[sid] = statement
	}
	return byStatementID
}

func TestTrustPolicyTrustsThreadServiceAccounts(t *testing.T) {
	f := &fakeIAM{}
	agent := runtimeAgent()

	_, err := clusterProviderWithFake(f).Ensure(context.Background(), agent, awsGrant())
	require.NoError(t, err)

	statements := trustStatements(t, f.createdTrust)
	require.Len(t, statements, 2)

	cluster := statements["AgentThreadServiceAccounts"]
	require.Equal(t,
		"arn:aws:iam::111111111111:oidc-provider/"+clusterProvider,
		cluster["Principal"].(map[string]any)["Federated"],
	)

	condition := cluster["Condition"].(map[string]any)
	require.Equal(t,
		"sts.amazonaws.com",
		condition["StringEquals"].(map[string]any)[clusterProvider+":aud"],
	)
	// One wildcard covers every thread, so adding a thread needs no IAM write. The prefix is
	// the agent's uid, not its name, so it cannot match another agent's threads.
	require.Equal(t,
		"system:serviceaccount:argus-aws-oidc-rdev:agent-0f8fad5bd9cb-*",
		condition["StringLike"].(map[string]any)[clusterProvider+":sub"],
	)
}

func TestTrustPolicyOmitsClusterStatementWithoutRuntime(t *testing.T) {
	f := &fakeIAM{}
	agent := runtimeAgent()
	agent.Spec.Runtime = nil

	_, err := clusterProviderWithFake(f).Ensure(context.Background(), agent, awsGrant())
	require.NoError(t, err)

	statements := trustStatements(t, f.createdTrust)
	require.Len(t, statements, 1)
	require.Contains(t, statements, "OktaAgentApp")
}

func TestEnsureCorrectsTrustDriftOnExistingRole(t *testing.T) {
	ctx := context.Background()
	agent := runtimeAgent()

	// A role created before the agent had a runtime trusts only Okta.
	oktaOnly, err := NewProvider(Config{OktaAppClientID: "aud"}, nil).trustPolicy("111111111111", agent)
	require.NoError(t, err)

	f := &fakeIAM{existingTrust: oktaOnly}
	p := clusterProviderWithFake(f)

	_, err = p.Ensure(ctx, agent, awsGrant())
	require.NoError(t, err)
	require.Empty(t, f.createdRoleName, "an existing role must not be recreated")
	require.Contains(t, f.updatedTrust, "AgentThreadServiceAccounts")

	// Already correct: no write, so a steady-state resync does not churn IAM.
	settled := &fakeIAM{existingTrust: f.updatedTrust}
	_, err = clusterProviderWithFake(settled).Ensure(ctx, agent, awsGrant())
	require.NoError(t, err)
	require.Empty(t, settled.updatedTrust)
}

func TestUnreachableAccount(t *testing.T) {
	require.True(t, unreachableAccount(assumeRoleFailure()))
	require.False(t, unreachableAccount(errors.New("throttled")))
	require.False(t, unreachableAccount(&smithy.OperationError{ServiceID: "IAM", OperationName: "DeleteRole", Err: errors.New("boom")}))
}
