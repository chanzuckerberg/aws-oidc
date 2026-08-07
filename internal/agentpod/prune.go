package agentpod

import (
	"context"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// pruneThreads deletes the objects of threads the agent no longer declares. Owner references
// do not cover this: the agent still exists, so nothing garbage-collects a thread that was
// merely removed from the spec. The workspace is deleted explicitly, since that is what
// releases the EBS volume.
//
// keep is the set of thread names to preserve; a nil or empty set removes every thread.
func (r *Reconciler) pruneThreads(ctx context.Context, agent *agentsv1.Agent, keep map[string]bool) error {
	log := logf.FromContext(ctx)

	inAgent := client.MatchingLabels{LabelAgent: agent.Name}
	inNamespace := client.InNamespace(r.Namespace)

	sets := &appsv1.StatefulSetList{}
	err := r.List(ctx, sets, inNamespace, inAgent)
	if err != nil {
		return fmt.Errorf("listing statefulsets of agent %s: %w", agent.Name, err)
	}

	accounts := &corev1.ServiceAccountList{}
	err = r.List(ctx, accounts, inNamespace, inAgent)
	if err != nil {
		return fmt.Errorf("listing service accounts of agent %s: %w", agent.Name, err)
	}

	claims := &corev1.PersistentVolumeClaimList{}
	err = r.List(ctx, claims, inNamespace, inAgent)
	if err != nil {
		return fmt.Errorf("listing workspaces of agent %s: %w", agent.Name, err)
	}

	var (
		stale []client.Object
		errs  []error
	)
	for i := range sets.Items {
		if !keep[sets.Items[i].Labels[LabelThread]] {
			stale = append(stale, &sets.Items[i])
		}
	}
	for i := range accounts.Items {
		if !keep[accounts.Items[i].Labels[LabelThread]] {
			stale = append(stale, &accounts.Items[i])
		}
	}
	for i := range claims.Items {
		if !keep[claims.Items[i].Labels[LabelThread]] {
			stale = append(stale, &claims.Items[i])
		}
	}

	for _, obj := range stale {
		log.Info("pruning object of removed agent thread",
			"agent", agent.Name,
			"thread", obj.GetLabels()[LabelThread],
			"kind", fmt.Sprintf("%T", obj),
			"name", obj.GetName())

		err = ignoreNotFound(r.Delete(ctx, obj))
		if err != nil {
			errs = append(errs, fmt.Errorf("pruning %s: %w", obj.GetName(), err))
		}
	}
	return errors.Join(errs...)
}

// pruneShared removes the objects an agent's threads share, for an agent that no longer runs
// in the cluster at all. They are keyed on kind rather than labels, so a thread's objects are
// never caught by this.
func (r *Reconciler) pruneShared(ctx context.Context, agent *agentsv1.Agent) error {
	shared := []client.Object{
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: agent.ServiceName(), Namespace: r.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: agent.AWSConfigMapName(), Namespace: r.Namespace}},
	}

	var errs []error
	for _, obj := range shared {
		err := ignoreNotFound(r.Delete(ctx, obj))
		if err != nil {
			errs = append(errs, fmt.Errorf("pruning %s: %w", obj.GetName(), err))
		}
	}
	return errors.Join(errs...)
}
