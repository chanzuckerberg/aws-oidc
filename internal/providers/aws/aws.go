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
	"log/slog"
	"net/url"
	"sort"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const providerName = "aws"

// IAMAPI is the subset of the IAM client the provider uses. It is an interface so the
// provider can be tested without real AWS calls.
type IAMAPI interface {
	GetRole(ctx context.Context, in *iam.GetRoleInput, opts ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	CreateRole(ctx context.Context, in *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	DeleteRole(ctx context.Context, in *iam.DeleteRoleInput, opts ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	ListAttachedRolePolicies(ctx context.Context, in *iam.ListAttachedRolePoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	AttachRolePolicy(ctx context.Context, in *iam.AttachRolePolicyInput, opts ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
	ListRolePolicies(ctx context.Context, in *iam.ListRolePoliciesInput, opts ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error)
	GetRolePolicy(ctx context.Context, in *iam.GetRolePolicyInput, opts ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error)
	PutRolePolicy(ctx context.Context, in *iam.PutRolePolicyInput, opts ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error)
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
	// DefaultTags are the deployment-level standard tags (for example project, env,
	// service) applied to every agent role. The operator layers its own managedBy/owner and
	// the per-agent tags on top of these.
	DefaultTags map[string]string
}

// managedByValue is the managedBy tag value for roles this operator creates, matching the
// standard tagging convention (Terraform-managed resources use "terraform").
const managedByValue = "aws-oidc-agent-operator"

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

	roleName := p.roleName(agent, g)
	sourceRole := sourceRoleName(g)

	existing, err := client.GetRole(ctx, &iam.GetRoleInput{RoleName: awssdk.String(roleName)})
	roleARN := ""
	if err == nil {
		roleARN = awssdk.ToString(existing.Role.Arn)
	} else {
		var notFound *iamtypes.NoSuchEntityException
		if !errors.As(err, &notFound) {
			return status, fmt.Errorf("getting role %s: %w", roleName, err)
		}

		trust, trustErr := p.trustPolicy(g.AccountID, agent.Spec.Owner)
		if trustErr != nil {
			return status, trustErr
		}

		input := &iam.CreateRoleInput{
			RoleName:                 awssdk.String(roleName),
			Path:                     awssdk.String(p.cfg.RolePath),
			AssumeRolePolicyDocument: awssdk.String(trust),
			Description:              awssdk.String(fmt.Sprintf("Agent %s for %s, mirrors %s", agent.Name, agent.Spec.OwnerEmail, sourceRole)),
			Tags:                     p.tags(agent),
		}
		// TODO: the permissions boundary is mandatory for real. It is skipped only until the
		// boundary policy exists in the account (shared-infra bootstrap). Without it this role
		// is unbounded, so do not enable this provider in a real account until
		// BoundaryPolicyName is set.
		if p.cfg.BoundaryPolicyName != "" {
			input.PermissionsBoundary = awssdk.String(p.boundaryARN(g.AccountID))
		}

		created, createErr := client.CreateRole(ctx, input)
		if createErr != nil {
			return status, fmt.Errorf("creating role %s: %w", roleName, createErr)
		}
		roleARN = awssdk.ToString(created.Role.Arn)
	}

	// Mirror the permissions of the role the grant was selected from, so the agent gets the
	// same access. Idempotent, so it also repairs an existing agent role on resync.
	err = p.mirrorPolicies(ctx, client, sourceRole, roleName)
	if err != nil {
		return status, fmt.Errorf("mirroring policies from %s onto %s: %w", sourceRole, roleName, err)
	}

	status.AWS.RoleARN = roleARN
	status.State = agentsv1.GrantStateProvisioned
	return status, nil
}

// mirrorPolicies copies the source role's permissions onto the agent role: every attached
// managed policy and every inline policy. It is idempotent (attaching an already-attached
// policy and putting an identical inline policy are both no-ops).
//
// TODO: this only adds; it does not detach policies the source role no longer has. Full
// drift reconciliation (removing policies the source dropped) is future work.
func (p *Provider) mirrorPolicies(ctx context.Context, client IAMAPI, sourceRole, agentRole string) error {
	attached, err := client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: awssdk.String(sourceRole),
	})
	if err != nil {
		return fmt.Errorf("listing attached policies of %s: %w", sourceRole, err)
	}
	for _, ap := range attached.AttachedPolicies {
		_, err = client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  awssdk.String(agentRole),
			PolicyArn: ap.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("attaching %s: %w", awssdk.ToString(ap.PolicyArn), err)
		}
	}

	inline, err := client.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: awssdk.String(sourceRole),
	})
	if err != nil {
		return fmt.Errorf("listing inline policies of %s: %w", sourceRole, err)
	}
	for _, name := range inline.PolicyNames {
		got, getErr := client.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
			RoleName:   awssdk.String(sourceRole),
			PolicyName: awssdk.String(name),
		})
		if getErr != nil {
			return fmt.Errorf("reading inline policy %s: %w", name, getErr)
		}
		// GetRolePolicy returns the document URL-encoded; PutRolePolicy wants plain JSON.
		doc, decodeErr := url.QueryUnescape(awssdk.ToString(got.PolicyDocument))
		if decodeErr != nil {
			return fmt.Errorf("decoding inline policy %s: %w", name, decodeErr)
		}
		_, putErr := client.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       awssdk.String(agentRole),
			PolicyName:     awssdk.String(name),
			PolicyDocument: awssdk.String(doc),
		})
		if putErr != nil {
			return fmt.Errorf("writing inline policy %s: %w", name, putErr)
		}
	}
	return nil
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

	roleName := p.roleName(agent, g)
	// TODO: once the role carries attached policies, detach them before DeleteRole.
	_, err = client.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: awssdk.String(roleName)})
	if err != nil {
		var notFound *iamtypes.NoSuchEntityException
		if errors.As(err, &notFound) {
			return nil
		}
		// If we cannot assume into the account (the provisioner role is absent or does not
		// trust us), there is nothing we could have provisioned there to clean up. Treat it
		// as done rather than wedging the finalizer forever on an account we cannot reach.
		if unreachableAccount(err) {
			slog.Warn("skipping teardown: cannot reach target account, treating as nothing to clean up",
				"provider", providerName, "account", g.AccountID, "role", roleName, "error", err)
			return nil
		}
		return fmt.Errorf("deleting role %s: %w", roleName, err)
	}
	return nil
}

// unreachableAccount reports whether the error was caused by failing to assume into the
// target account (the STS AssumeRole step), rather than by the IAM call itself. It walks the
// wrapped error chain looking for a smithy operation error from STS, which the SDK surfaces
// when credential resolution (the assume-role) fails.
func unreachableAccount(err error) bool {
	for e := err; e != nil; e = errors.Unwrap(e) {
		opErr, ok := e.(*smithy.OperationError)
		if ok && opErr.ServiceID == "STS" {
			return true
		}
	}
	return false
}

// roleName derives the per-agent role name, including the owner's shortname so the role is
// identifiable by human at a glance: <shortname>-agent-<agent>-<source role base name>, for
// example jheath-agent-playground-readonly-readonly. When the owner email is unknown the
// shortname is omitted.
//
// TODO: IAM role names cap at 64 characters and allow only [\w+=,.@-]; long names need
// truncation or hashing.
func (p *Provider) roleName(agent *agentsv1.Agent, g *agentsv1.AWSGrant) string {
	base := sourceRoleName(g)
	short := shortName(agent.Spec.OwnerEmail)
	if short == "" {
		return fmt.Sprintf("agent-%s-%s", agent.Name, base)
	}
	return fmt.Sprintf("%s-agent-%s-%s", short, agent.Name, base)
}

// sourceRoleName is the base name of the role a grant was selected from (the role whose
// permissions the agent role mirrors).
func sourceRoleName(g *agentsv1.AWSGrant) string {
	base := g.RoleName
	if base == "" {
		base = roleNameFromARN(g.RoleARN)
	}
	return base[strings.LastIndex(base, "/")+1:]
}

// shortName is the local part of an email (before the @), used as the human's shortname.
func shortName(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return ""
	}
	return email[:at]
}

func (p *Provider) boundaryARN(accountID string) string {
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, p.cfg.BoundaryPolicyName)
}

// tags builds the role's tag set: the deployment-level standard tags (project/env/service),
// the standard managedBy and owner, and the per-agent tags. owner is the human's email (the
// person responsible for the agent); agent-owner is their Okta subject. Keys are emitted in
// stable order.
func (p *Provider) tags(agent *agentsv1.Agent) []iamtypes.Tag {
	merged := map[string]string{}
	for k, v := range p.cfg.DefaultTags {
		merged[k] = v
	}

	merged["managedBy"] = managedByValue
	merged["agent-name"] = agent.Name
	merged["agent-owner"] = agent.Spec.Owner // Okta subject (the "client id")

	owner := agent.Spec.OwnerEmail
	if owner == "" {
		owner = agent.Spec.Owner
	}
	merged["owner"] = coalesceUnknown(owner)

	tags := make([]iamtypes.Tag, 0, len(merged))
	for k, v := range merged {
		tags = append(tags, iamtypes.Tag{Key: awssdk.String(k), Value: awssdk.String(v)})
	}
	sort.Slice(tags, func(i, j int) bool { return awssdk.ToString(tags[i].Key) < awssdk.ToString(tags[j].Key) })
	return tags
}

// coalesceUnknown mirrors the fogg tagging convention of defaulting an empty value to
// "unknown".
func coalesceUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
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

// DefaultClientFactory builds an IAM client from the operator's ambient credentials, with no
// role assumption, so it only reaches the operator's own account. It is a fallback for local
// runs and tests. For real cross-account provisioning the operator uses
// NewAssumeRoleClientFactory, which assumes each account's provisioner role.
func DefaultClientFactory(ctx context.Context, accountID string) (IAMAPI, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return iam.NewFromConfig(cfg), nil
}
