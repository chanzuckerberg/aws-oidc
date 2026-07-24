// Package provisioner turns a grant into a per-agent IAM role in a target account.
//
// STUB: the real IAM work is deferred. The intended behavior is that the operator's
// control-plane identity (IRSA) assumes the per-account agent-provisioner role, then
// creates or repairs a role under /agents/ that is bounded by the account's permissions
// boundary and whose trust policy only the shared agent Okta client, acting for the owning
// subject, can use. Every role must carry the permissions boundary; without it the
// provisioner is an escalation machine. None of that is implemented yet.
package provisioner

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by the stub until the IAM provisioning is built.
var ErrNotImplemented = errors.New("provisioner: not implemented")

// Config is the static provisioning policy, the same for every agent.
type Config struct {
	// IssuerHost is the Okta issuer host without scheme (for example "czi.okta.com"). It
	// names the account's OIDC provider and prefixes the trust condition keys.
	IssuerHost string
	// AgentClientID is the shared agent Okta client id. The trust policy will condition the
	// audience on it so an agent token cannot assume a human's poweruser role.
	AgentClientID string
	// RolePath is the IAM path every agent role lives under (must start and end with "/").
	RolePath string
	// BoundaryPolicyName is the permissions boundary policy name, resolved per account.
	BoundaryPolicyName string
}

// RoleSpec is the desired per-agent role for one grant.
type RoleSpec struct {
	AccountID string
	AgentName string
	RoleName  string
	// Owner is the Okta subject the trust policy binds the role to.
	Owner string
	// ManagedPolicyARN is the single catalog policy attached to the role.
	ManagedPolicyARN string
}

// Interface is the provisioning surface the controller depends on. Keeping it an interface
// lets the reconciler be wired against the stub now and a real IAM implementation later.
type Interface interface {
	// EnsureRole creates or repairs the per-agent role for a grant and returns its ARN. It
	// must be idempotent so resync can correct drift.
	EnsureRole(ctx context.Context, spec RoleSpec) (roleARN string, err error)
	// DeleteRole removes the per-agent role. It must be a no-op when the role is already
	// gone so a finalizer can call it more than once.
	DeleteRole(ctx context.Context, accountID, roleName string) error
}

// Provisioner is the stub implementation. TODO: replace with the IAM implementation that
// assumes the agent-provisioner role and reconciles boundary-bounded, trust-scoped roles.
type Provisioner struct {
	cfg Config
}

var _ Interface = (*Provisioner)(nil)

// New returns a stub Provisioner. RolePath defaults to /agents/ when empty.
func New(cfg Config) *Provisioner {
	if cfg.RolePath == "" {
		cfg.RolePath = "/agents/"
	}
	return &Provisioner{cfg: cfg}
}

// EnsureRole is not implemented yet.
//
// TODO: assume the per-account agent-provisioner role, then create the role under
// cfg.RolePath with the permissions boundary attached and a trust policy that conditions
// czi.okta.com:aud on cfg.AgentClientID and czi.okta.com:sub on spec.Owner, and attach the
// single catalog managed policy. On resync, repair a drifted trust policy, re-apply the
// boundary, and detach any policy that is not spec.ManagedPolicyARN.
func (p *Provisioner) EnsureRole(ctx context.Context, spec RoleSpec) (string, error) {
	return "", ErrNotImplemented
}

// DeleteRole is not implemented yet.
//
// TODO: detach every managed policy and delete the role, treating a missing role as success.
func (p *Provisioner) DeleteRole(ctx context.Context, accountID, roleName string) error {
	return ErrNotImplemented
}
