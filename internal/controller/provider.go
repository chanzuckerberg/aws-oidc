package controller

import (
	"context"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// Provider reconciles the grants it owns for an agent. Each third-party system (AWS today;
// Jira, GitHub, Slack, Google Workspace, ... later) implements this interface and is
// registered with the reconciler. The reconciler dispatches every grant to the provider
// that Handles it, so onboarding a new system is additive: implement Provider and register
// it in cmd/operator.go. No change to the reconcile loop is needed.
type Provider interface {
	// Name identifies the provider, used for logging and status.provider.
	Name() string

	// Handles reports whether this provider owns the given grant (by inspecting which
	// section of the grant union is set).
	Handles(grant agentsv1.Grant) bool

	// Ensure provisions one grant and returns its status. A returned error signals a
	// transient failure the reconciler should retry with backoff; a permanent problem
	// should instead be reported as a Failed GrantStatus with a nil error.
	Ensure(ctx context.Context, agent *agentsv1.Agent, grant agentsv1.Grant) (agentsv1.GrantStatus, error)

	// Delete tears down whatever the grant provisioned. It must be idempotent so the
	// finalizer can call it more than once and treat an already-gone resource as success.
	Delete(ctx context.Context, agent *agentsv1.Agent, grant agentsv1.Grant) error
}
