// Package controller reconciles Agent custom resources into provisioned access.
//
// STUB: this is the skeleton of the reconcile loop, not the running operator. It walks each
// grant and records the intended per-provider status, but does not yet talk to Kubernetes or
// any third party. The remaining work is called out in TODOs:
//
//   - Wire this as a controller-runtime reconciler (SetupWithManager, a rate-limited
//     workqueue, leader election) driven by cmd/operator.go.
//   - Add a finalizer so deleting an Agent tears down whatever each grant provisioned.
//   - Provision each grant through its provider (AWS today via internal/provisioner; other
//     providers as they are added) and correct drift on resync.
//   - Write status (per-grant result, the Ready condition, observedGeneration) back through
//     the Kubernetes API status subresource.
package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/provisioner"
)

// AgentReconciler reconciles an Agent into provisioned per-provider access.
type AgentReconciler struct {
	// AWS provisions the AWS side of a grant. Other providers get their own fields here as
	// the registry grows.
	AWS provisioner.Interface
}

// NewAgentReconciler returns a reconciler wired to the AWS provisioner.
func NewAgentReconciler(aws provisioner.Interface) *AgentReconciler {
	return &AgentReconciler{AWS: aws}
}

// Reconcile computes the desired status for an Agent, one entry per grant, marking each
// grant Pending.
//
// TODO: once wired to a manager, provision each grant through its provider, set the grant
// state to Provisioned (or Failed with a message), run the deletion finalizer, and persist
// this status through the Kubernetes API.
func (r *AgentReconciler) Reconcile(ctx context.Context, agent *agentsv1.Agent) (agentsv1.AgentStatus, error) {
	status := agentsv1.AgentStatus{
		ObservedGeneration: agent.Generation,
		Grants:             make([]agentsv1.GrantStatus, 0, len(agent.Spec.Grants)),
	}

	for _, grant := range agent.Spec.Grants {
		status.Grants = append(status.Grants, grantStatusFor(grant))
	}

	// No grant is provisioned yet, so the agent is never Ready in the stub.
	status.Conditions = []metav1.Condition{{
		Type:               agentsv1.ConditionReady,
		Status:             metav1.ConditionFalse,
		Reason:             "ProvisioningNotImplemented",
		ObservedGeneration: agent.Generation,
		LastTransitionTime: metav1.Now(),
	}}

	return status, nil
}

// grantStatusFor maps a desired grant to its initial status, dispatching on the provider
// union. New providers add a case here.
func grantStatusFor(grant agentsv1.Grant) agentsv1.GrantStatus {
	switch {
	case grant.AWS != nil:
		// TODO: roleARN, err := r.AWS.EnsureRole(ctx, spec); set Provisioned/Failed.
		return agentsv1.GrantStatus{
			Provider: agentsv1.ProviderAWS,
			AWS:      &agentsv1.AWSGrantStatus{AccountID: grant.AWS.AccountID},
			State:    agentsv1.GrantStatePending,
			Message:  "provisioning not implemented",
		}
	default:
		return agentsv1.GrantStatus{
			State:   agentsv1.GrantStateFailed,
			Message: "grant has no known provider",
		}
	}
}
