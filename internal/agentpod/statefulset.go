package agentpod

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	awsconfigclient "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
)

// ensureStatefulSet creates or updates the workload running one thread and returns its current
// state. A thread is one replica: the workspace is a ReadWriteOnce EBS volume, and two pods
// sharing a working directory is not what a thread is.
//
// A StatefulSet's volume claim template is immutable, so when the requested workspace size
// changes the set is recreated with its pods orphaned, which leaves the running pod and the
// resized claim in place.
func (r *Reconciler) ensureStatefulSet(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) (*appsv1.StatefulSet, error) {
	desired, err := r.statefulSet(agent, thread)
	if err != nil {
		return nil, err
	}

	existing := &appsv1.StatefulSet{}
	err = r.Get(ctx, client.ObjectKeyFromObject(desired), existing)
	switch {
	case apierrors.IsNotFound(err):
		err = r.Create(ctx, desired)
		if err != nil {
			return nil, fmt.Errorf("creating statefulset %s: %w", desired.Name, err)
		}
		return desired, nil
	case err != nil:
		return nil, fmt.Errorf("getting statefulset %s: %w", desired.Name, err)
	}

	if claimTemplatesDiffer(existing, desired) {
		err = r.recreateOrphaned(ctx, existing, desired)
		if err != nil {
			return nil, err
		}
		return desired, nil
	}

	// Carry over the fields the API server defaults or the StatefulSet controller owns, so the
	// update is limited to what this reconciler actually manages.
	updated := existing.DeepCopy()
	updated.Labels = desired.Labels
	updated.Spec.Replicas = desired.Spec.Replicas
	updated.Spec.Template = desired.Spec.Template

	if equalStatefulSets(existing, updated) {
		return existing, nil
	}

	err = r.Update(ctx, updated)
	if err != nil {
		return nil, fmt.Errorf("updating statefulset %s: %w", desired.Name, err)
	}
	return updated, nil
}

// recreateOrphaned replaces a StatefulSet whose immutable volume claim template changed. The
// delete orphans its dependents, so the pod keeps running and the workspace claim survives to
// be adopted by the replacement.
func (r *Reconciler) recreateOrphaned(ctx context.Context, existing, desired *appsv1.StatefulSet) error {
	orphan := metav1.DeletePropagationOrphan
	err := r.Delete(ctx, existing, &client.DeleteOptions{PropagationPolicy: &orphan})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting statefulset %s for workspace resize: %w", existing.Name, err)
	}

	err = r.Create(ctx, desired)
	if err != nil {
		return fmt.Errorf("recreating statefulset %s for workspace resize: %w", desired.Name, err)
	}
	return nil
}

func (r *Reconciler) statefulSet(agent *agentsv1.Agent, thread agentsv1.AgentThread) (*appsv1.StatefulSet, error) {
	size, err := r.storageSize(agent)
	if err != nil {
		return nil, err
	}

	labels := threadLabels(agent, thread.Name)
	replicas := int32(1)
	if thread.Suspended {
		replicas = 0
	}

	set := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.ThreadStatefulSetName(thread.Name),
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: agent.ServiceName(),
			Selector:    &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       r.podSpec(agent, thread),
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name:   agentsv1.WorkspaceVolumeName,
					Labels: labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					StorageClassName: ptr(r.storageClass(agent)),
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: size},
					},
				},
			}},
		},
	}

	err = controllerutil.SetControllerReference(agent, set, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("setting owner on statefulset %s: %w", set.Name, err)
	}
	return set, nil
}

// podSpec is the thread's pod: the agent container, its workspace, the shared AWS config, and
// the projected token it exchanges for the agent's roles.
func (r *Reconciler) podSpec(agent *agentsv1.Agent, thread agentsv1.AgentThread) corev1.PodSpec {
	runtime := agent.Spec.Runtime

	return corev1.PodSpec{
		ServiceAccountName: agent.ThreadServiceAccountName(thread.Name),
		NodeSelector:       runtime.NodeSelector,
		// The pod's only credential is the token projected below, minted for STS. Leaving the
		// default Kubernetes API token out means agent code cannot talk to the cluster at all.
		AutomountServiceAccountToken: ptr(false),
		SecurityContext: &corev1.PodSecurityContext{
			SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
		},
		Containers: []corev1.Container{{
			Name:       agentContainerName,
			Image:      r.image(agent),
			Command:    r.command(agent),
			Args:       runtime.Args,
			Resources:  runtime.Resources,
			WorkingDir: workspaceMountPath,
			Env:        r.env(agent, thread),
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: ptr(false),
			},
			VolumeMounts: []corev1.VolumeMount{
				{Name: agentsv1.WorkspaceVolumeName, MountPath: workspaceMountPath},
				{Name: awsConfigVolume, MountPath: awsConfigMountPath, ReadOnly: true},
				{Name: tokenVolume, MountPath: tokenMountPath, ReadOnly: true},
			},
		}},
		Volumes: []corev1.Volume{
			{
				Name: awsConfigVolume,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: agent.AWSConfigMapName()},
					},
				},
			},
			{
				Name: tokenVolume,
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Audience:          tokenAudience,
								ExpirationSeconds: ptr(tokenExpirationSeconds),
								Path:              "token",
							},
						}},
					},
				},
			},
		},
	}
}

// env points the AWS SDK at the rendered config and tells the agent which agent and thread it
// is. The agent's own variables come last so they cannot overwrite these.
func (r *Reconciler) env(agent *agentsv1.Agent, thread agentsv1.AgentThread) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{Name: "AWS_CONFIG_FILE", Value: awsConfigFilePath},
		{Name: "AWS_PROFILE", Value: awsconfigclient.AgentScopedProfile},
		{Name: "AWS_REGION", Value: r.Region},
		{Name: "AGENT_NAME", Value: agent.Name},
		{Name: "AGENT_THREAD", Value: thread.Name},
		{Name: "AGENT_OWNER_EMAIL", Value: agent.Spec.OwnerEmail},
	}

	reserved := make(map[string]bool, len(env))
	for _, e := range env {
		reserved[e.Name] = true
	}
	for _, e := range agent.Spec.Runtime.Env {
		if reserved[e.Name] {
			continue
		}
		env = append(env, corev1.EnvVar{Name: e.Name, Value: e.Value})
	}
	return env
}

// ensureWorkspaceSize grows a thread's existing workspace to the requested size. A provisioned
// volume cannot shrink, so a smaller request is ignored rather than failing the thread. The
// storage class allows expansion, but EBS limits a volume to one modification every six hours,
// so a rejected patch is reported rather than retried into a wedge.
func (r *Reconciler) ensureWorkspaceSize(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) error {
	desired, err := r.storageSize(agent)
	if err != nil {
		return err
	}

	claim := &corev1.PersistentVolumeClaim{}
	key := types.NamespacedName{Namespace: r.Namespace, Name: agent.ThreadWorkspaceClaimName(thread.Name)}
	err = r.Get(ctx, key, claim)
	if apierrors.IsNotFound(err) {
		// The StatefulSet has not created it yet; the claim template carries the size.
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting workspace %s: %w", key.Name, err)
	}

	current := claim.Spec.Resources.Requests[corev1.ResourceStorage]
	if desired.Cmp(current) <= 0 {
		return nil
	}

	patched := claim.DeepCopy()
	patched.Spec.Resources.Requests[corev1.ResourceStorage] = desired
	err = r.Patch(ctx, patched, client.MergeFrom(claim))
	if err != nil {
		return fmt.Errorf("expanding workspace %s to %s: %w", key.Name, desired.String(), err)
	}
	return nil
}

// adoptWorkspace points the claim's ownership at the Agent. A StatefulSet does not delete the
// claims it created, and its own reference would disappear when the set is recreated for a
// resize, so the Agent owns the claim directly. Deleting the agent then releases the EBS
// volume through the storage class's Delete reclaim policy.
func (r *Reconciler) adoptWorkspace(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) error {
	claim := &corev1.PersistentVolumeClaim{}
	key := types.NamespacedName{Namespace: r.Namespace, Name: agent.ThreadWorkspaceClaimName(thread.Name)}
	err := r.Get(ctx, key, claim)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting workspace %s: %w", key.Name, err)
	}

	// A non-controller reference, because the StatefulSet controller may claim the controller
	// slot under a retention policy.
	patched := claim.DeepCopy()
	err = controllerutil.SetOwnerReference(agent, patched, r.Scheme)
	if err != nil {
		return fmt.Errorf("setting owner on workspace %s: %w", key.Name, err)
	}
	if len(patched.OwnerReferences) == len(claim.OwnerReferences) {
		return nil
	}

	if patched.Labels == nil {
		patched.Labels = map[string]string{}
	}
	for k, v := range threadLabels(agent, thread.Name) {
		patched.Labels[k] = v
	}

	err = r.Patch(ctx, patched, client.MergeFrom(claim))
	if err != nil {
		return fmt.Errorf("adopting workspace %s: %w", key.Name, err)
	}
	return nil
}

// claimTemplatesDiffer reports whether the immutable part of the spec changed, which can only
// be applied by recreating the set.
func claimTemplatesDiffer(existing, desired *appsv1.StatefulSet) bool {
	if len(existing.Spec.VolumeClaimTemplates) != len(desired.Spec.VolumeClaimTemplates) {
		return true
	}
	for i := range desired.Spec.VolumeClaimTemplates {
		want := desired.Spec.VolumeClaimTemplates[i].Spec
		got := existing.Spec.VolumeClaimTemplates[i].Spec
		if !storageRequestsEqual(got, want) {
			return true
		}
		if ptrValue(got.StorageClassName) != ptrValue(want.StorageClassName) {
			return true
		}
	}
	return false
}

func storageRequestsEqual(a, b corev1.PersistentVolumeClaimSpec) bool {
	left := a.Resources.Requests[corev1.ResourceStorage]
	right := b.Resources.Requests[corev1.ResourceStorage]
	return left.Cmp(right) == 0
}

// equalStatefulSets compares the parts of a set this reconciler owns, so a steady-state
// resync does not write.
func equalStatefulSets(existing, desired *appsv1.StatefulSet) bool {
	return equality.Semantic.DeepEqual(existing.Labels, desired.Labels) &&
		equality.Semantic.DeepEqual(existing.Spec.Replicas, desired.Spec.Replicas) &&
		equality.Semantic.DeepEqual(existing.Spec.Template, desired.Spec.Template)
}

func ptr[T any](v T) *T { return &v }

func ptrValue[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
