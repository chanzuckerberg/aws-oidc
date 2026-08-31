package agentpod

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	awsconfigclient "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
)

// ensureStatefulSet creates or updates the workload running one thread and returns its current
// state. A thread is one replica: each pod mounts the agent's shared EFS workspace at its own
// subPath, so the thread has an isolated working tree even though the volume is shared.
func (r *Reconciler) ensureStatefulSet(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) (*appsv1.StatefulSet, error) {
	desired := r.statefulSet(agent, thread)

	existing := &appsv1.StatefulSet{}
	err := r.Get(ctx, client.ObjectKeyFromObject(desired), existing)
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

func (r *Reconciler) statefulSet(agent *agentsv1.Agent, thread agentsv1.AgentThread) *appsv1.StatefulSet {
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
		},
	}

	_ = controllerutil.SetControllerReference(agent, set, r.Scheme)
	return set
}

// sharedMountPath is where the shared workspace directory is mounted in every thread pod.
const sharedMountPath = "/shared"

// podSpec is the thread's pod: the agent container, its thread-private working tree and the
// shared directory (both subPaths of the agent's EFS workspace PVC), the AWS config, and the
// projected token it exchanges for the agent's roles.
func (r *Reconciler) podSpec(agent *agentsv1.Agent, thread agentsv1.AgentThread) corev1.PodSpec {
	agentRuntime := agent.Spec.Runtime
	uid := int64(1000)

	return corev1.PodSpec{
		ServiceAccountName: agent.ThreadServiceAccountName(thread.Name),
		NodeSelector:       agentRuntime.NodeSelector,
		// The pod's only credential is the token projected below, minted for STS. Leaving the
		// default Kubernetes API token out means agent code cannot talk to the cluster at all.
		AutomountServiceAccountToken: ptr(false),
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser:      &uid,
			RunAsGroup:     &uid,
			FSGroup:        &uid,
			SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
		},
		Containers: []corev1.Container{{
			Name:       agentContainerName,
			Image:      r.image(agent),
			Command:    r.command(agent),
			Args:       agentRuntime.Args,
			Resources:  agentRuntime.Resources,
			WorkingDir: workspaceMountPath,
			Env:        r.env(agent, thread),
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: ptr(false),
			},
			VolumeMounts: r.volumeMounts(agent, thread),
		}},
		Volumes: r.volumes(agent),
	}
}

// volumeMounts returns the container's volume mounts, including the Anthropic token when WIF
// is configured.
func (r *Reconciler) volumeMounts(agent *agentsv1.Agent, thread agentsv1.AgentThread) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      agentsv1.WorkspaceVolumeName,
			MountPath: workspaceMountPath,
			SubPath:   agent.ThreadWorkspaceSubPath(thread.Name),
		},
		{
			Name:      agentsv1.WorkspaceVolumeName,
			MountPath: sharedMountPath,
			SubPath:   agent.SharedWorkspaceSubPath(),
		},
		{Name: awsConfigVolume, MountPath: awsConfigMountPath, ReadOnly: true},
		{Name: tokenVolume, MountPath: tokenMountPath, ReadOnly: true},
	}
	if r.anthropicWIFConfigured() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      anthropicTokenVolume,
			MountPath: anthropicTokenMountPath,
			ReadOnly:  true,
		})
	}
	if r.githubAppConfigured() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      githubAppKeyVolume,
			MountPath: githubAppKeyMountPath,
			ReadOnly:  true,
		})
	}
	return mounts
}

// volumes returns the pod volumes, including the Anthropic projected token when WIF is
// configured.
func (r *Reconciler) volumes(agent *agentsv1.Agent) []corev1.Volume {
	vols := []corev1.Volume{
		{
			Name: agentsv1.WorkspaceVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: agent.WorkspaceClaimName(),
				},
			},
		},
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
	}
	if r.anthropicWIFConfigured() {
		vols = append(vols, corev1.Volume{
			Name: anthropicTokenVolume,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Audience:          r.AnthropicTokenAudience,
							ExpirationSeconds: ptr(anthropicTokenExpirationSecs),
							Path:              "token",
						},
					}},
				},
			},
		})
	}
	if r.githubAppConfigured() {
		vols = append(vols, corev1.Volume{
			Name: githubAppKeyVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  r.GitHubAppPrivateKeySecret,
					DefaultMode: ptr(githubAppKeyMode),
					Items: []corev1.KeyToPath{{
						Key:  r.GitHubAppPrivateKeySecretKey,
						Path: githubAppKeyFileName,
					}},
				},
			},
		})
	}
	return vols
}

// env points the AWS SDK at the rendered config and tells the agent which agent and thread it
// is. Anthropic WIF env vars are injected when the operator is configured for Claude WIF. The
// agent's own variables come last so they cannot overwrite these reserved names.
func (r *Reconciler) env(agent *agentsv1.Agent, thread agentsv1.AgentThread) []corev1.EnvVar {
	env := []corev1.EnvVar{
		// HOME must be writable so the AWS CLI can cache STS credentials and SSO tokens.
		// /workspace is the thread's own persistent directory; using it keeps the cache across
		// pod restarts without needing a separate writable volume.
		{Name: "HOME", Value: workspaceMountPath},
		{Name: "AWS_CONFIG_FILE", Value: awsConfigFilePath},
		{Name: "AWS_PROFILE", Value: awsconfigclient.AgentScopedProfile},
		{Name: "AWS_REGION", Value: r.Region},
		{Name: "AGENT_NAME", Value: agent.Name},
		{Name: "AGENT_THREAD", Value: thread.Name},
		{Name: "AGENT_OWNER_EMAIL", Value: agent.Spec.OwnerEmail},
	}
	if r.anthropicWIFConfigured() {
		env = append(env,
			corev1.EnvVar{Name: "ANTHROPIC_IDENTITY_TOKEN_FILE", Value: anthropicTokenFilePath},
			corev1.EnvVar{Name: "ANTHROPIC_FEDERATION_RULE_ID", Value: r.AnthropicFederationRuleID},
			corev1.EnvVar{Name: "ANTHROPIC_ORGANIZATION_ID", Value: r.AnthropicOrganizationID},
			corev1.EnvVar{Name: "ANTHROPIC_SERVICE_ACCOUNT_ID", Value: r.AnthropicServiceAccountID},
		)
	}
	if r.githubAppConfigured() {
		env = append(env,
			corev1.EnvVar{Name: "GITHUB_APP_ID", Value: r.GitHubAppID},
			corev1.EnvVar{Name: "GITHUB_APP_INSTALLATION_ID", Value: r.GitHubAppInstallationID},
			corev1.EnvVar{Name: "GITHUB_APP_PRIVATE_KEY_FILE", Value: githubAppKeyFilePath},
		)
		if r.GitHubAPIURL != "" {
			env = append(env, corev1.EnvVar{Name: "GITHUB_API_URL", Value: r.GitHubAPIURL})
		}
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

// anthropicWIFConfigured reports whether the operator has all four fields needed to project
// an Anthropic token into pods.
func (r *Reconciler) anthropicWIFConfigured() bool {
	return r.AnthropicFederationRuleID != "" &&
		r.AnthropicOrganizationID != "" &&
		r.AnthropicServiceAccountID != "" &&
		r.AnthropicTokenAudience != ""
}

// githubAppConfigured reports whether the operator has everything needed to give a thread the
// GitHub App's identity.
func (r *Reconciler) githubAppConfigured() bool {
	return r.GitHubAppID != "" &&
		r.GitHubAppInstallationID != "" &&
		r.GitHubAppPrivateKeySecret != ""
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
