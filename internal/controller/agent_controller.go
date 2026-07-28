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

	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/equality"
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

// defaultGrantConcurrency bounds how many grants of a single agent are provisioned at once.
// An agent can carry many grants across many accounts, so they are reconciled in parallel.
const defaultGrantConcurrency = 8

// AgentReconciler reconciles Agent objects by dispatching their grants to providers.
type AgentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Providers []Provider
	// MaxConcurrentGrants bounds parallel per-grant provisioning within one agent. Zero
	// uses defaultGrantConcurrency.
	MaxConcurrentGrants int
}

// +kubebuilder:rbac:groups=agents.czi.team,resources=agents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/finalizers,verbs=update

// Reconcile drives one Agent toward its desired state.
func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	statuses, reconcileErr := r.reconcileGrants(ctx, &agent)

	err = r.writeStatus(ctx, &agent, statuses)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Returning the error requeues with the workqueue's exponential backoff.
	return ctrl.Result{}, reconcileErr
}

// reconcileGrants provisions every grant, in parallel up to the concurrency limit, and
// returns the per-grant statuses aligned with spec.grants plus a joined error of any
// transient failures. Each goroutine writes only its own status slot, so no locking is
// needed. A failure of one grant does not cancel the others; they all get attempted each
// pass, and failures requeue with backoff.
func (r *AgentReconciler) reconcileGrants(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.GrantStatus, error) {
	log := logf.FromContext(ctx)
	grants := agent.Spec.Grants

	statuses := make([]agentsv1.GrantStatus, len(grants))
	errs := make([]error, len(grants))

	limit := r.MaxConcurrentGrants
	if limit <= 0 {
		limit = defaultGrantConcurrency
	}

	group := new(errgroup.Group)
	group.SetLimit(limit)

	for i := range grants {
		i := i
		grant := grants[i]

		provider := r.providerFor(grant)
		if provider == nil {
			statuses[i] = agentsv1.GrantStatus{
				State:   agentsv1.GrantStateFailed,
				Message: "no provider registered for grant",
			}
			continue
		}

		group.Go(func() error {
			status, err := provider.Ensure(ctx, agent, grant)
			if err != nil {
				log.Error(err, "provisioning grant", "provider", provider.Name())
				status.State = agentsv1.GrantStateFailed
				status.Message = err.Error()
				errs[i] = err
			}
			statuses[i] = status
			// Return nil so one grant's failure does not cancel its siblings.
			return nil
		})
	}
	_ = group.Wait()

	return statuses, errors.Join(errs...)
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
	// Compute the desired status on a copy so we can compare it to the current one and only
	// write when it actually changed. Writing unconditionally every reconcile creates a
	// status -> watch -> reconcile storm, which (with a lagging cache) shows up as a stream
	// of optimistic-lock "object has been modified" conflicts.
	desired := agent.Status.DeepCopy()
	desired.Grants = statuses
	desired.ObservedGeneration = agent.Generation

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
	meta.SetStatusCondition(&desired.Conditions, condition)

	if equality.Semantic.DeepEqual(&agent.Status, desired) {
		return nil
	}

	desired.DeepCopyInto(&agent.Status)
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
