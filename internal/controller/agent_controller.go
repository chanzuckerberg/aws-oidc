// Package controller reconciles Agent custom resources into provisioned access across one
// or more third-party providers. The reconcile loop itself is provider-agnostic: it walks
// spec.grants and dispatches each grant to the registered Provider that handles it, then
// records per-grant status and an aggregate Ready condition. AWS is the only provider wired
// today; others are added by implementing Provider and registering them.
package controller

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// agentFinalizer gates deletion so the operator can tear down provisioned access before the
// Agent object is removed.
const agentFinalizer = "agents.czi.team/finalizer"

// AgentReconciler reconciles Agent objects by dispatching their grants to providers.
type AgentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Providers []Provider
}

// +kubebuilder:rbac:groups=agents.czi.team,resources=agents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/finalizers,verbs=update

// Reconcile drives one Agent toward its desired state.
func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var agent agentsv1.Agent
	err := r.Get(ctx, req.NamespacedName, &agent)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !agent.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &agent)
	}

	if !controllerutil.ContainsFinalizer(&agent, agentFinalizer) {
		controllerutil.AddFinalizer(&agent, agentFinalizer)
		err = r.Update(ctx, &agent)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("adding finalizer: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	statuses := make([]agentsv1.GrantStatus, 0, len(agent.Spec.Grants))
	var reconcileErr error
	for _, grant := range agent.Spec.Grants {
		provider := r.providerFor(grant)
		if provider == nil {
			statuses = append(statuses, agentsv1.GrantStatus{
				State:   agentsv1.GrantStateFailed,
				Message: "no provider registered for grant",
			})
			continue
		}

		status, err := provider.Ensure(ctx, &agent, grant)
		if err != nil {
			log.Error(err, "provisioning grant", "provider", provider.Name())
			status.State = agentsv1.GrantStateFailed
			status.Message = err.Error()
			reconcileErr = errors.Join(reconcileErr, err)
		}
		statuses = append(statuses, status)
	}

	err = r.writeStatus(ctx, &agent, statuses)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Returning the error requeues with the workqueue's exponential backoff.
	return ctrl.Result{}, reconcileErr
}

// reconcileDelete runs the providers' teardown, then drops the finalizer.
func (r *AgentReconciler) reconcileDelete(ctx context.Context, agent *agentsv1.Agent) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(agent, agentFinalizer) {
		return ctrl.Result{}, nil
	}

	for _, grant := range agent.Spec.Grants {
		provider := r.providerFor(grant)
		if provider == nil {
			continue
		}
		err := provider.Delete(ctx, agent, grant)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting grant via %s: %w", provider.Name(), err)
		}
	}

	controllerutil.RemoveFinalizer(agent, agentFinalizer)
	err := r.Update(ctx, agent)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("removing finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

// writeStatus records per-grant results, the observed generation, and the aggregate Ready
// condition on the status subresource.
func (r *AgentReconciler) writeStatus(ctx context.Context, agent *agentsv1.Agent, statuses []agentsv1.GrantStatus) error {
	agent.Status.Grants = statuses
	agent.Status.ObservedGeneration = agent.Generation

	ready := true
	for _, s := range statuses {
		if s.State != agentsv1.GrantStateProvisioned {
			ready = false
			break
		}
	}

	condition := metav1.Condition{
		Type:               agentsv1.ConditionReady,
		ObservedGeneration: agent.Generation,
		Status:             metav1.ConditionFalse,
		Reason:             "GrantsPending",
		Message:            "one or more grants are not yet provisioned",
	}
	if ready {
		condition.Status = metav1.ConditionTrue
		condition.Reason = "AllGrantsProvisioned"
		condition.Message = "all grants provisioned"
	}
	meta.SetStatusCondition(&agent.Status.Conditions, condition)

	err := r.Status().Update(ctx, agent)
	if err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	return nil
}

// providerFor returns the first registered provider that handles the grant, or nil.
func (r *AgentReconciler) providerFor(grant agentsv1.Grant) Provider {
	for _, p := range r.Providers {
		if p.Handles(grant) {
			return p
		}
	}
	return nil
}

// SetupWithManager registers the reconciler with the manager.
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentsv1.Agent{}).
		Complete(r)
}
