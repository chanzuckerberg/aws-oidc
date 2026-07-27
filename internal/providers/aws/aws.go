// Package aws is the AWS provider for the agent operator. It turns an AWS grant into a
// per-agent IAM role in the grant's account. The role trusts the shared agent Okta app via
// web-identity federation, scoped to the owning subject.
//
// This is the first iteration. It creates the role if it does not already exist; several
// pieces it depends on do not exist yet and are marked with TODOs:
//   - the shared agent Okta app (its client id feeds the trust policy audience),
//   - cross-account access (assuming each account's agent-provisioner role),
//   - the permissions boundary,
//   - the permissions policy that scopes the agent to a subset of the owner's role.
package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const providerName = "aws"

// IAMAPI is the subset of the IAM client the provider uses. It is an interface so the
// provider can be tested without real AWS calls.
type IAMAPI interface {
	GetRole(ctx context.Context, in *iam.GetRoleInput, opts ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	CreateRole(ctx context.Context, in *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	DeleteRole(ctx context.Context, in *iam.DeleteRoleInput, opts ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
}

// ClientFactory returns an IAM client scoped to a target account.
type ClientFactory func(ctx context.Context, accountID string) (IAMAPI, error)

// Config is the static AWS provisioning policy, the same for every agent.
type Config struct {
	// IssuerHost is the Okta issuer host without scheme (for example "czi.okta.com"). It
	// names the account's OIDC provider and prefixes the trust condition keys.
	IssuerHost string
	// OktaAppClientID is the shared agent Okta app client id used as the trust policy
	// audience.
	//
	// TODO: this app does not exist yet. It is being created in shared-infra under
	// okta-czi. Until its client id is wired in here, the aud condition will not match any
	// real token and no agent can actually assume the role.
	OktaAppClientID string
	// RolePath is the IAM path every agent role lives under (must start and end with "/").
	RolePath string
	// BoundaryPolicyName is the permissions boundary policy name, resolved per account.
	//
	// TODO: the boundary policy does not exist yet (part of the shared-infra bootstrap).
	// When empty the role is created without a boundary; set this once the policy exists.
	BoundaryPolicyName string
}

// Provider provisions AWS grants.
type Provider struct {
	cfg       Config
	clientFor ClientFactory
}

// NewProvider returns an AWS provider. RolePath defaults to /agents/ and IssuerHost to
// czi.okta.com. A nil clientFor uses DefaultClientFactory.
func NewProvider(cfg Config, clientFor ClientFactory) *Provider {
	if cfg.RolePath == "" {
		cfg.RolePath = "/agents/"
	}
	if cfg.IssuerHost == "" {
		cfg.IssuerHost = "czi.okta.com"
	}
	if clientFor == nil {
		clientFor = DefaultClientFactory
	}
	return &Provider{cfg: cfg, clientFor: clientFor}
}

// Name identifies the provider.
func (p *Provider) Name() string { return providerName }

// Handles reports whether the grant targets AWS.
func (p *Provider) Handles(grant agentsv1.Grant) bool { return grant.AWS != nil }

// Ensure creates the per-agent IAM role for an AWS grant if it does not already exist, and
// reports the role ARN in status.
func (p *Provider) Ensure(ctx context.Context, agent *agentsv1.Agent, grant agentsv1.Grant) (agentsv1.GrantStatus, error) {
	g := grant.AWS
	status := agentsv1.GrantStatus{
		Provider: providerName,
		AWS:      &agentsv1.AWSGrantStatus{AccountID: g.AccountID},
	}

	client, err := p.clientFor(ctx, g.AccountID)
	if err != nil {
		return status, fmt.Errorf("iam client for account %s: %w", g.AccountID, err)
	}

	roleName := p.roleName(agent.Name, g)

	existing, err := client.GetRole(ctx, &iam.GetRoleInput{RoleName: awssdk.String(roleName)})
	if err == nil {
		// TODO: reconcile drift on the existing role (repair the trust policy, re-apply the
		// boundary) and attach the scoped permissions policy. For now, treat presence as done.
		status.AWS.RoleARN = awssdk.ToString(existing.Role.Arn)
		status.State = agentsv1.GrantStateProvisioned
		return status, nil
	}

	var notFound *iamtypes.NoSuchEntityException
	if !errors.As(err, &notFound) {
		return status, fmt.Errorf("getting role %s: %w", roleName, err)
	}

	trust, err := p.trustPolicy(g.AccountID, agent.Spec.Owner)
	if err != nil {
		return status, err
	}

	input := &iam.CreateRoleInput{
		RoleName:                 awssdk.String(roleName),
		Path:                     awssdk.String(p.cfg.RolePath),
		AssumeRolePolicyDocument: awssdk.String(trust),
		Description:              awssdk.String(fmt.Sprintf("Agent %s (owner %s)", agent.Name, agent.Spec.Owner)),
		Tags:                     p.tags(agent),
	}
	// TODO: the permissions boundary is mandatory for real. It is skipped only until the
	// boundary policy exists in the account (shared-infra bootstrap). Without it this role
	// is unbounded, so do not enable this provider in a real account until BoundaryPolicyName
	// is set.
	if p.cfg.BoundaryPolicyName != "" {
		input.PermissionsBoundary = awssdk.String(p.boundaryARN(g.AccountID))
	}

	created, err := client.CreateRole(ctx, input)
	if err != nil {
		return status, fmt.Errorf("creating role %s: %w", roleName, err)
	}

	// TODO: attach a permissions policy that scopes the agent to a subset of g.RoleARN (the
	// owner's role the grant derives from). Until then the role can be assumed but grants no
	// permissions.
	status.AWS.RoleARN = awssdk.ToString(created.Role.Arn)
	status.State = agentsv1.GrantStateProvisioned
	return status, nil
}

// Delete removes the per-agent role. A missing role is treated as success.
func (p *Provider) Delete(ctx context.Context, agent *agentsv1.Agent, grant agentsv1.Grant) error {
	g := grant.AWS
	if g == nil {
		return nil
	}

	client, err := p.clientFor(ctx, g.AccountID)
	if err != nil {
		return fmt.Errorf("iam client for account %s: %w", g.AccountID, err)
	}

	roleName := p.roleName(agent.Name, g)
	// TODO: once the role carries attached policies, detach them before DeleteRole.
	_, err = client.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: awssdk.String(roleName)})
	if err != nil {
		var notFound *iamtypes.NoSuchEntityException
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("deleting role %s: %w", roleName, err)
	}
	return nil
}

// roleName derives the per-agent role name for a grant: <agent>-<source role base name>.
//
// TODO: IAM role names cap at 64 characters and allow only [\w+=,.@-]; long agent or role
// names need truncation or hashing.
func (p *Provider) roleName(agentName string, g *agentsv1.AWSGrant) string {
	base := g.RoleName
	if base == "" {
		base = roleNameFromARN(g.RoleARN)
	}
	base = base[strings.LastIndex(base, "/")+1:]
	return fmt.Sprintf("%s-%s", agentName, base)
}

func (p *Provider) boundaryARN(accountID string) string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, p.cfg.BoundaryPolicyName)
}

func (p *Provider) tags(agent *agentsv1.Agent) []iamtypes.Tag {
	return []iamtypes.Tag{
		{Key: awssdk.String("managed-by"), Value: awssdk.String("aws-oidc-agent-operator")},
		{Key: awssdk.String("agent-name"), Value: awssdk.String(agent.Name)},
		{Key: awssdk.String("agent-owner"), Value: awssdk.String(agent.Spec.Owner)},
	}
}

// trustPolicy builds the web-identity trust document: the account's Okta OIDC provider,
// conditioned on the agent app audience and the owner subject.
func (p *Provider) trustPolicy(accountID, owner string) (string, error) {
	providerARN := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountID, p.cfg.IssuerHost)
	doc := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{{
			"Effect":    "Allow",
			"Principal": map[string]any{"Federated": providerARN},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]any{
				"StringEquals": map[string]string{
					p.cfg.IssuerHost + ":aud": p.cfg.OktaAppClientID,
					p.cfg.IssuerHost + ":sub": owner,
				},
			},
		}},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("marshalling trust policy: %w", err)
	}
	return string(raw), nil
}

func roleNameFromARN(roleARN string) string {
	idx := strings.Index(roleARN, ":role/")
	if idx < 0 {
		return roleARN
	}
	return roleARN[idx+len(":role/"):]
}

// DefaultClientFactory builds an IAM client from the operator's ambient credentials.
//
// TODO: this does not do cross-account access. It should assume each target account's
// agent-provisioner role (arn:aws:iam::<accountID>:role/agent-provisioner) from the
// operator's IRSA identity. Until then it only works against the operator's own account.
func DefaultClientFactory(ctx context.Context, accountID string) (IAMAPI, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return iam.NewFromConfig(cfg), nil
}
