package agentpod

import (
	"context"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

func (r *Reconciler) pruneObsoleteRuntimeObjects(ctx context.Context, agent *agentsv1.Agent) error {
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

	var obsolete []client.Object
	for i := range sets.Items {
		if sets.Items[i].Name != agent.StatefulSetName() {
			obsolete = append(obsolete, &sets.Items[i])
		}
	}
	for i := range accounts.Items {
		if accounts.Items[i].Name != agent.ServiceAccountName() {
			obsolete = append(obsolete, &accounts.Items[i])
		}
	}

	var errs []error
	for _, object := range obsolete {
		err = ignoreNotFound(r.Delete(ctx, object))
		if err != nil {
			errs = append(errs, fmt.Errorf("pruning %s: %w", object.GetName(), err))
		}
	}
	return errors.Join(errs...)
}

// pruneRuntime removes runtime objects when the agent no longer runs in the cluster. The PVC
// remains until the Agent is deleted so disabling the runtime does not destroy the owner's data.
func (r *Reconciler) pruneRuntime(ctx context.Context, agent *agentsv1.Agent) error {
	var errs []error
	err := r.pruneMemoryImportArtifacts(ctx, agent, map[string]bool{})
	if err != nil {
		errs = append(errs, err)
	}

	objects := []client.Object{
		&appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: agent.StatefulSetName(), Namespace: r.Namespace}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: agent.ServiceAccountName(), Namespace: r.Namespace}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: agent.ServiceName(), Namespace: r.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: agent.AWSConfigMapName(), Namespace: r.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: agent.ClaudeConfigMapName(), Namespace: r.Namespace}},
	}

	for _, obj := range objects {
		err = ignoreNotFound(r.Delete(ctx, obj))
		if err != nil {
			errs = append(errs, fmt.Errorf("pruning %s: %w", obj.GetName(), err))
		}
	}
	return errors.Join(errs...)
}
