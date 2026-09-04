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
	"strings"

	"golang.org/x/sync/errgroup"
	appsv1 "k8s.io/api/apps/v1"
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
const (
	agentFinalizer             = "agents.czi.team/finalizer"
	argoCDTrackingIDAnnotation = "argocd.argoproj.io/tracking-id"
)

// defaultGrantConcurrency bounds how many grants of a single agent are provisioned at once.
// An agent can carry many grants across many accounts, so they are reconciled in parallel.
const defaultGrantConcurrency = 8

// WorkspaceReconciler runs an agent's workspaces in the cluster. It is separate from Provider
// because a workspace is per-agent rather than per-grant, and because it needs the provisioned
// role ARNs the providers return.
type WorkspaceReconciler interface {
	Reconcile(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.WorkspaceStatus, error)
}

// AgentReconciler reconciles Agent objects by dispatching their grants to providers and
// running their workspaces in the cluster.
type AgentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Providers []Provider
	// Workspaces runs the agent's workspaces. Nil leaves agents without any pods, which is how the
	// operator behaves in an environment where the runtime is not enabled.
	Workspaces WorkspaceReconciler
	// MaxConcurrentGrants bounds parallel per-grant provisioning within one agent. Zero
	// uses defaultGrantConcurrency.
	MaxConcurrentGrants int
	ArgoCDTrackingID    string
}

// +kubebuilder:rbac:groups=agents.czi.team,resources=agents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agents.czi.team,resources=agents/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts;services;configmaps;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

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

	if r.setManagedMetadata(&agent) {
		err = r.Update(ctx, &agent)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("updating agent metadata: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	grantStatuses, reconcileErr := r.reconcileGrants(ctx, &agent)

	// Workspaces run after the grants, because the AWS config their pods mount is built from the
	// role ARNs the grants provisioned.
	workspaceStatuses, workspaceErr := r.reconcileWorkspaces(ctx, &agent)

	err = r.writeStatus(ctx, &agent, grantStatuses, workspaceStatuses)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Returning the error requeues with the workqueue's exponential backoff.
	return ctrl.Result{}, errors.Join(reconcileErr, workspaceErr)
}

func (r *AgentReconciler) setManagedMetadata(agent *agentsv1.Agent) bool {
	changed := false
	if !controllerutil.ContainsFinalizer(agent, agentFinalizer) {
		controllerutil.AddFinalizer(agent, agentFinalizer)
		changed = true
	}
	if r.ArgoCDTrackingID != "" && agent.Annotations[argoCDTrackingIDAnnotation] != r.ArgoCDTrackingID {
		if agent.Annotations == nil {
			agent.Annotations = make(map[string]string, 1)
		}
		agent.Annotations[argoCDTrackingIDAnnotation] = r.ArgoCDTrackingID
		changed = true
	}
	return changed
}

// reconcileWorkspaces runs the agent's workspaces, if the operator has a workspace reconciler.
func (r *AgentReconciler) reconcileWorkspaces(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.WorkspaceStatus, error) {
	if r.Workspaces == nil {
		return nil, nil
	}
	return r.Workspaces.Reconcile(ctx, agent)
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

// writeStatus records per-grant and per-workspace results, the observed generation, and the
// aggregate Ready and RuntimeReady conditions on the status subresource.
func (r *AgentReconciler) writeStatus(ctx context.Context, agent *agentsv1.Agent, grants []agentsv1.GrantStatus, workspaces []agentsv1.WorkspaceStatus) error {
	// Compute the desired status on a copy so we can compare it to the current one and only
	// write when it actually changed. Writing unconditionally every reconcile creates a
	// status -> watch -> reconcile storm, which (with a lagging cache) shows up as a stream
	// of optimistic-lock "object has been modified" conflicts.
	desired := agent.Status.DeepCopy()
	desired.Grants = grants
	desired.Workspaces = workspaces
	desired.ObservedGeneration = agent.Generation

	ready := true
	for _, s := range grants {
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
	meta.SetStatusCondition(&desired.Conditions, runtimeCondition(agent, workspaces))

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

// runtimeCondition summarizes the agent's workspaces. A suspended workspace is deliberately idle, so
// it does not hold the condition false. An agent that does not run in the cluster reports true
// with nothing to run, which keeps the condition meaningful rather than permanently false.
func runtimeCondition(agent *agentsv1.Agent, workspaces []agentsv1.WorkspaceStatus) metav1.Condition {
	condition := metav1.Condition{
		Type:               agentsv1.ConditionRuntimeReady,
		ObservedGeneration: agent.Generation,
		Status:             metav1.ConditionTrue,
	}

	if agent.Spec.Runtime == nil {
		condition.Reason = "NoRuntime"
		condition.Message = "agent does not run in the cluster"
		return condition
	}

	for _, workspace := range workspaces {
		if workspace.State == agentsv1.WorkspaceStateRunning || workspace.State == agentsv1.WorkspaceStateSuspended {
			continue
		}
		condition.Status = metav1.ConditionFalse
		condition.Reason = "WorkspacesPending"
		condition.Message = fmt.Sprintf("workspace %s is %s", workspace.Name, strings.ToLower(string(workspace.State)))
		if workspace.Message != "" {
			condition.Message += ": " + workspace.Message
		}
		return condition
	}

	condition.Reason = "AllWorkspacesRunning"
	condition.Message = fmt.Sprintf("all %d workspaces running", len(workspaces))
	return condition
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

// SetupWithManager registers the reconciler with the manager. It also watches the StatefulSets
// it owns, so a workspace's pod becoming ready is reflected in the agent's status without waiting
// for a resync.
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentsv1.Agent{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
