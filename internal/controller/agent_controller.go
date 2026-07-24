// Package controller reconciles Agent custom resources into per-agent IAM roles.
//
// STUB: this is the skeleton of the reconcile loop, not the running operator. It computes
// the desired per-agent role for each grant from the catalog, but does not yet talk to
// Kubernetes or AWS. The remaining work is called out in TODOs:
//
//   - Wire this as a controller-runtime reconciler (SetupWithManager, a rate-limited
//     workqueue, leader election) driven by cmd/operator.go.
//   - Add a finalizer so deleting an Agent deletes its target-account roles.
//   - Call the provisioner to create/repair/delete roles and correct drift on resync.
//   - Write status (per-grant roleArn and state, the Ready condition, observedGeneration)
//     back through the Kubernetes API status subresource.
package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/catalog"
	"github.com/chanzuckerberg/aws-oidc/internal/provisioner"
)

// AgentReconciler reconciles an Agent into per-agent IAM roles.
type AgentReconciler struct {
	Provisioner provisioner.Interface
	Catalog     *catalog.Catalog
}

// NewAgentReconciler returns a reconciler wired to a provisioner and catalog.
func NewAgentReconciler(p provisioner.Interface, c *catalog.Catalog) *AgentReconciler {
	return &AgentReconciler{Provisioner: p, Catalog: c}
}

// roleSpecFor resolves a single grant to the per-agent role that should exist for it. The
// role name is <agent>-<catalogPolicyId>, matching the plan (for example data-bot-s3-readonly).
func (r *AgentReconciler) roleSpecFor(agent *agentsv1.Agent, grant agentsv1.Grant) (provisioner.RoleSpec, error) {
	policy, ok := r.Catalog.Get(grant.CatalogPolicyID)
	if !ok {
		return provisioner.RoleSpec{}, fmt.Errorf("catalog policy %q not found", grant.CatalogPolicyID)
	}
	return provisioner.RoleSpec{
		AccountID:        grant.AccountID,
		AgentName:        agent.Name,
		RoleName:         fmt.Sprintf("%s-%s", agent.Name, grant.CatalogPolicyID),
		Owner:            agent.Spec.Owner,
		ManagedPolicyARN: policy.ARN(grant.AccountID),
	}, nil
}

// Reconcile computes the desired status for an Agent. It resolves each grant against the
// catalog and records the intended role, marking every grant Pending.
//
// TODO: once wired to a manager, call r.Provisioner.EnsureRole for each grant, set the
// grant state to Provisioned (or Failed with a message), run the deletion finalizer via
// r.Provisioner.DeleteRole, and persist this status through the Kubernetes API.
func (r *AgentReconciler) Reconcile(ctx context.Context, agent *agentsv1.Agent) (agentsv1.AgentStatus, error) {
	status := agentsv1.AgentStatus{
		ObservedGeneration: agent.Generation,
		Grants:             make([]agentsv1.GrantStatus, 0, len(agent.Spec.Grants)),
	}

	allReady := true
	for _, grant := range agent.Spec.Grants {
		spec, err := r.roleSpecFor(agent, grant)
		if err != nil {
			allReady = false
			status.Grants = append(status.Grants, agentsv1.GrantStatus{
				AccountID:       grant.AccountID,
				CatalogPolicyID: grant.CatalogPolicyID,
				State:           agentsv1.GrantStateFailed,
				Message:         err.Error(),
			})
			continue
		}

		// TODO: roleARN, err := r.Provisioner.EnsureRole(ctx, spec); set Provisioned/Failed.
		allReady = false
		status.Grants = append(status.Grants, agentsv1.GrantStatus{
			AccountID:       spec.AccountID,
			CatalogPolicyID: grant.CatalogPolicyID,
			State:           agentsv1.GrantStatePending,
			Message:         "provisioning not implemented",
		})
	}

	readyStatus := metav1.ConditionFalse
	reason := "Provisioning"
	if allReady {
		readyStatus = metav1.ConditionTrue
		reason = "AllGrantsProvisioned"
	}
	status.Conditions = []metav1.Condition{{
		Type:               agentsv1.ConditionReady,
		Status:             readyStatus,
		Reason:             reason,
		ObservedGeneration: agent.Generation,
		LastTransitionTime: metav1.Now(),
	}}

	return status, nil
}
