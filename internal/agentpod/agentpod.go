// Package agentpod runs an agent's workspaces in the cluster. An agent is not a single session:
// a person runs several workspaces of the same agent at once, so each workspace gets its own pod
// while sharing the agent's access and its workspace.
//
// Per agent it maintains a headless Service (so each workspace pod has a stable DNS name), a
// ConfigMap holding the rendered AWS config, and a single ReadWriteMany PVC backed by an EFS
// access point. Per workspace it maintains a ServiceAccount and a StatefulSet. Each workspace pod
// mounts the shared PVC at a workspace-specific subPath for its working tree and at a common
// subPath for files shared between workspaces. The workspace's pod exchanges its projected service
// account token for the agent's IAM roles, which trust the cluster's OIDC issuer for exactly
// these service accounts.
package agentpod

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/agentdefaults"
)

// defaultWorkspaceSize is the storage request placed on the EFS workspace PVC when the agent
// does not specify one. The EFS CSI driver ignores the value — EFS is elastic — but the
// Kubernetes API requires a positive storage request.
const defaultWorkspaceSize = "50Gi"

const (
	// LabelAgent and LabelWorkspace identify which agent and workspace an object belongs to. The
	// reconciler selects on them to find objects whose workspace is gone from the spec.
	LabelAgent     = "agents.czi.team/agent"
	LabelWorkspace = "agents.czi.team/workspace"

	// labelManagedBy marks the objects this package owns.
	labelManagedBy = "app.kubernetes.io/managed-by"
	managedByValue = "aws-oidc-agent-operator"

	// tokenMountPath is where the projected service account token is mounted, and
	// tokenFilePath is the file the AWS SDK reads it from. IRSA's webhook does not inject a
	// token without a role annotation on the service account, and an agent's roles are
	// per-grant rather than one per service account, so the volume is declared explicitly.
	tokenMountPath = "/var/run/secrets/agents.czi.team/serviceaccount"
	tokenFilePath  = tokenMountPath + "/token"

	// awsConfigMountPath is where the rendered AWS config is mounted, and awsConfigFilePath is
	// what AWS_CONFIG_FILE points at.
	awsConfigMountPath = "/etc/aws"
	awsConfigFilePath  = awsConfigMountPath + "/config"

	// workspaceMountPath is the workspace's persistent working directory.
	workspaceMountPath = "/workspace"

	// tokenAudience is the audience STS requires on a projected token used for web identity.
	tokenAudience = "sts.amazonaws.com"

	// tokenExpirationSeconds is the projected token's lifetime. The kubelet rotates it at 80%
	// of this, and the AWS SDK re-reads the file on each refresh.
	tokenExpirationSeconds int64 = 3600

	agentContainerName = "agent"
	awsConfigVolume    = "aws-config"
	tokenVolume        = "aws-token"

	// anthropicTokenVolume and its paths are the separate projected token the Anthropic SDK
	// exchanges for a short-lived Claude access token via WIF. It uses a different audience
	// than the AWS token and a shorter lifetime so the kubelet rotates it before JTI replay
	// protection can reject a re-used assertion.
	anthropicTokenVolume    = "anthropic-token"
	anthropicTokenMountPath = "/var/run/secrets/anthropic.com"
	anthropicTokenFilePath  = anthropicTokenMountPath + "/token"

	// anthropicTokenExpirationSecs is the Anthropic projected token's lifetime. At 80% of
	// this (480 s) the kubelet rotates the file. The Anthropic SDK advisory refresh also
	// fires at token_lifetime - 120 s = 480 s, so the file always holds a fresh jti by the
	// time the SDK re-reads it.
	anthropicTokenExpirationSecs int64 = 600

	// githubAppKeyVolume mounts the GitHub App's RSA private key. The key itself is never an
	// env var: the image's credential helper reads the file and exchanges it for an
	// installation access token that expires in an hour.
	githubAppKeyVolume    = "github-app-key"
	githubAppKeyMountPath = "/var/run/secrets/github.com"
	githubAppKeyFileName  = "private-key.pem"
	githubAppKeyFilePath  = githubAppKeyMountPath + "/" + githubAppKeyFileName

	// githubAppKeyMode is 0440. Secret volume files are owned by root and grouped to the pod's
	// fsGroup, so group-read is what makes the key readable to uid 1000 and nothing wider.
	githubAppKeyMode int32 = 0o440

	// tailscaleTokenVolume and tailscaleTokenMountPath are the projected SA token the
	// tailscale-up init container exchanges for a Tailscale machine key.
	tailscaleTokenVolume    = "tailscale-token"
	tailscaleTokenMountPath = "/var/run/secrets/tailscale.com"
	tailscaleTokenFilePath  = tailscaleTokenMountPath + "/token"
	tailscaleTunResource    = corev1.ResourceName("agents.czi.team/tun")

	// tailscaleTokenExpirationSecs is the tailscale token's lifetime. Shorter than the AWS
	// token so the kubelet rotates it frequently. The entrypoint re-reads the file each time
	// tailscale up is called, so the pod always enrolls with a fresh token.
	tailscaleTokenExpirationSecs int64 = 600

	// managedSettingsVolume mounts the agent-managed-settings ConfigMap at the path Claude
	// reads as its enterprise-managed settings layer.
	managedSettingsVolume    = "managed-settings"
	managedSettingsMountPath = "/etc/claude-code"
	// managedSettingsMode 0755 makes shell scripts in the ConfigMap executable.
	managedSettingsMode int32 = 0o755
)

// Config is the operator-level policy for running agent workspaces, the same for every agent.
type Config struct {
	// Namespace is where the workspaces run. It is the operator's own namespace, so owner
	// references garbage-collect an agent's objects when the agent is deleted.
	Namespace string
	// DefaultsLoader reads live defaults from the agent-defaults ConfigMap. When set its
	// values take precedence over the static fields below. CRD spec values always win.
	DefaultsLoader *agentdefaults.Loader
	// DefaultImage is the agent image used when spec.runtime.image is unset and the
	// ConfigMap loader does not provide one.
	DefaultImage string
	// DefaultCommand is the command an agent workspace runs when neither the agent nor the image
	// provides a long-running entrypoint. Without it a base image whose entrypoint exits
	// leaves the pod crash-looping.
	DefaultCommand []string
	// StorageClass is the storage class the per-agent workspace PVC is provisioned from.
	// It must be a ReadWriteMany class backed by the EFS CSI driver.
	StorageClass string
	// Region is the AWS region written into the rendered AWS config.
	Region string
	// MaxWorkspaces bounds how many workspaces one agent may run, so a single Agent write cannot
	// ask for an unbounded number of pods.
	MaxWorkspaces int

	// AnthropicFederationRuleID, AnthropicOrganizationID, AnthropicServiceAccountID, and
	// AnthropicTokenAudience configure Workload Identity Federation with Anthropic. When all
	// four are non-empty the operator adds a second projected token (audience
	// AnthropicTokenAudience) to every workspace pod and sets the four ANTHROPIC_* env vars the
	// Claude SDK and CLI need to exchange it for a Claude access token. When any field is
	// empty the Anthropic token and env vars are omitted, so the operator degrades gracefully
	// in clusters that have not yet configured Claude WIF.
	AnthropicFederationRuleID string
	AnthropicOrganizationID   string
	AnthropicServiceAccountID string
	AnthropicTokenAudience    string

	// GitHubAppID, GitHubAppInstallationID and GitHubAppPrivateKeySecret configure the shared
	// GitHub App every workspace clones and opens pull requests as. When all three are non-empty
	// the operator mounts the app's private key from the named Secret and sets the GITHUB_APP_*
	// env vars the image's git credential helper and gh wrapper read. When any is empty the
	// mount and env vars are omitted, so a cluster without a GitHub App still runs agents.
	//
	// EnsureGitHubAppSecret writes that Secret at startup and returns its name.
	GitHubAppID             string
	GitHubAppInstallationID string
	// GitHubAppInstallationMap routes specific repository owners to other installations of the
	// same GitHub App, so one agent can reach repositories in more than one organization. It is
	// a comma or space separated list of owner=installation-id pairs, for example
	// "evolutionaryscale=158867890". An owner that is not listed uses GitHubAppInstallationID.
	// Empty means every owner uses the default installation.
	GitHubAppInstallationMap  string
	GitHubAppPrivateKeySecret string
	// GitHubAPIURL overrides the API endpoint for GitHub Enterprise. Empty means github.com.
	GitHubAPIURL string

	// TailscaleTokenAudience is the audience placed on the projected service-account token
	// the entrypoint passes to "tailscale up --id-token". The form is
	// "api.tailscale.com/<oidc-client-id>". When empty tailscale enrollment is skipped.
	TailscaleTokenAudience string
	// TailscaleTag is the tailscale tag the pod advertises (e.g. "tag:mantis-shrimp").
	TailscaleTag string
	// ManagedSettingsConfigMap is the name of the ConfigMap holding managed-settings.json
	// and ssh-guard.sh. When non-empty the ConfigMap is mounted at /etc/claude-code so
	// Claude picks it up as its enterprise-managed settings layer.
	ManagedSettingsConfigMap string
}

// Reconciler drives an agent's workspaces toward the spec.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config
}

// New returns a Reconciler with defaults applied.
func New(c client.Client, scheme *runtime.Scheme, cfg Config) *Reconciler {
	if cfg.StorageClass == "" {
		cfg.StorageClass = "efs-agent-workspaces"
	}
	if cfg.Region == "" {
		cfg.Region = "us-west-2"
	}
	if cfg.MaxWorkspaces <= 0 {
		cfg.MaxWorkspaces = defaultMaxWorkspaces
	}
	return &Reconciler{Client: c, Scheme: scheme, Config: cfg}
}

// defaultMaxWorkspaces is the per-agent workspace ceiling when none is configured.
const defaultMaxWorkspaces = 5

// Reconcile brings the agent's workspaces in line with its spec and returns their statuses,
// aligned with spec.workspaces. The grant statuses come from the AWS provider and carry the role
// ARNs the pods' AWS config needs, so this runs after the grants are reconciled.
//
// An agent with no runtime has no workspaces, so everything the agent owns is pruned.
func (r *Reconciler) Reconcile(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.WorkspaceStatus, error) {
	if agent.Spec.Runtime == nil {
		err := r.pruneWorkspaces(ctx, agent, nil)
		if err != nil {
			return nil, err
		}
		return nil, r.pruneShared(ctx, agent)
	}

	err := r.ensureService(ctx, agent)
	if err != nil {
		return nil, err
	}
	err = r.ensureAWSConfig(ctx, agent)
	if err != nil {
		return nil, err
	}
	err = r.ensureWorkspace(ctx, agent)
	if err != nil {
		return nil, err
	}

	workspaces := agent.Spec.Workspaces
	statuses := make([]agentsv1.WorkspaceStatus, 0, len(workspaces))
	keep := make(map[string]bool, len(workspaces))
	var errs []error

	for i := range workspaces {
		workspace := workspaces[i]

		if i >= r.MaxWorkspaces {
			statuses = append(statuses, agentsv1.WorkspaceStatus{
				Name:    workspace.Name,
				State:   agentsv1.WorkspaceStateFailed,
				Message: fmt.Sprintf("agent is limited to %d workspaces", r.MaxWorkspaces),
			})
			continue
		}
		keep[workspace.Name] = true

		status, err := r.reconcileWorkspace(ctx, agent, workspace)
		if err != nil {
			status.State = agentsv1.WorkspaceStateFailed
			status.Message = err.Error()
			errs = append(errs, err)
		}
		statuses = append(statuses, status)
	}

	// Prune after reconciling, so a workspace that failed to come up is not also torn down.
	err = r.pruneWorkspaces(ctx, agent, keep)
	if err != nil {
		errs = append(errs, err)
	}

	return statuses, errors.Join(errs...)
}

// reconcileWorkspace ensures one workspace's service account and workload, then reports what it
// found. The shared workspace is created once per agent before any workspace is reconciled.
func (r *Reconciler) reconcileWorkspace(ctx context.Context, agent *agentsv1.Agent, workspace agentsv1.AgentWorkspace) (agentsv1.WorkspaceStatus, error) {
	status := agentsv1.WorkspaceStatus{
		Name:               workspace.Name,
		ServiceAccountName: agent.WorkspaceServiceAccountName(workspace.Name),
		StatefulSetName:    agent.WorkspaceStatefulSetName(workspace.Name),
		State:              agentsv1.WorkspaceStatePending,
	}

	err := r.ensureServiceAccount(ctx, agent, workspace)
	if err != nil {
		return status, err
	}

	set, err := r.ensureStatefulSet(ctx, agent, workspace)
	if err != nil {
		return status, err
	}

	status.ReadyReplicas = set.Status.ReadyReplicas
	switch {
	case workspace.Suspended:
		status.State = agentsv1.WorkspaceStateSuspended
	case set.Status.ReadyReplicas > 0:
		status.State = agentsv1.WorkspaceStateRunning
	}
	return status, nil
}

func (r *Reconciler) ensureService(ctx context.Context, agent *agentsv1.Agent) error {
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.ServiceName(),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Labels = agentLabels(agent)
		service.Spec.ClusterIP = corev1.ClusterIPNone
		service.Spec.Selector = agentLabels(agent)
		// The workspaces serve no traffic yet. A port is declared anyway because a headless
		// service with no ports publishes no DNS records for its pods.
		service.Spec.Ports = []corev1.ServicePort{{Name: "agent", Port: 8080}}
		return controllerutil.SetControllerReference(agent, service, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service %s: %w", service.Name, err)
	}
	return nil
}

// ensureAWSConfig writes the agent's rendered AWS config. Every workspace mounts the same one,
// since workspaces differ in workspace, not in access.
func (r *Reconciler) ensureAWSConfig(ctx context.Context, agent *agentsv1.Agent) error {
	rendered, err := r.renderAWSConfig(agent)
	if err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.AWSConfigMapName(),
		Namespace: r.Namespace,
	}}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Labels = agentLabels(agent)
		configMap.Data = map[string]string{"config": rendered}
		return controllerutil.SetControllerReference(agent, configMap, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring config map %s: %w", configMap.Name, err)
	}
	return nil
}

// ensureServiceAccount creates the identity the workspace's pod runs as. Its name is what the
// agent's IAM roles trust, so the pod can assume them with its projected token.
func (r *Reconciler) ensureServiceAccount(ctx context.Context, agent *agentsv1.Agent, workspace agentsv1.AgentWorkspace) error {
	account := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.WorkspaceServiceAccountName(workspace.Name),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, account, func() error {
		account.Labels = workspaceLabels(agent, workspace.Name)
		return controllerutil.SetControllerReference(agent, account, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service account %s: %w", account.Name, err)
	}
	return nil
}

// ensureWorkspace creates the shared ReadWriteMany PVC for the agent. All workspaces mount it:
// each workspace at its own subPath for an isolated working tree, and every workspace at the shared
// subPath for files passed between workspaces. The EFS CSI driver provisions a fresh access
// point for this PVC, so no other agent's data is reachable inside it.
//
// The PVC is owned by the Agent, so it is garbage-collected when the agent is deleted.
func (r *Reconciler) ensureWorkspace(ctx context.Context, agent *agentsv1.Agent) error {
	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.WorkspaceClaimName(),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.Labels = agentLabels(agent)
		if pvc.Spec.AccessModes == nil {
			pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
			pvc.Spec.StorageClassName = ptr(r.workspaceStorageClass(agent))
			pvc.Spec.Resources = corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: r.workspaceSize(agent)},
			}
		}
		return controllerutil.SetControllerReference(agent, pvc, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring workspace %s: %w", pvc.Name, err)
	}
	return nil
}

// agentLabels identifies every object belonging to one agent.
func agentLabels(agent *agentsv1.Agent) map[string]string {
	return map[string]string{
		LabelAgent:     agent.Name,
		labelManagedBy: managedByValue,
	}
}

// workspaceLabels identifies the objects belonging to one workspace of one agent.
func workspaceLabels(agent *agentsv1.Agent, workspace string) map[string]string {
	labels := agentLabels(agent)
	labels[LabelWorkspace] = workspace
	return labels
}

func (r *Reconciler) loadDefaults() *agentdefaults.Defaults {
	if r.DefaultsLoader == nil {
		return &agentdefaults.Defaults{}
	}
	d, err := r.DefaultsLoader.Load()
	if err != nil {
		slog.Warn("loading agent defaults", "error", err)
	}
	if d == nil {
		return &agentdefaults.Defaults{}
	}
	return d
}

func (r *Reconciler) image(agent *agentsv1.Agent) string {
	if agent.Spec.Runtime.Image != "" {
		return agent.Spec.Runtime.Image
	}
	if d := r.loadDefaults(); d.Image != "" {
		return d.Image
	}
	return r.DefaultImage
}

func (r *Reconciler) command(agent *agentsv1.Agent) []string {
	if len(agent.Spec.Runtime.Command) > 0 {
		return agent.Spec.Runtime.Command
	}
	if d := r.loadDefaults(); len(d.Command) > 0 {
		return d.Command
	}
	return r.DefaultCommand
}

func (r *Reconciler) workspaceStorageClass(agent *agentsv1.Agent) string {
	if agent.Spec.Runtime != nil && agent.Spec.Runtime.StorageClass != "" {
		return agent.Spec.Runtime.StorageClass
	}
	if d := r.loadDefaults(); d.StorageClass != "" {
		return d.StorageClass
	}
	return r.StorageClass
}

func (r *Reconciler) workspaceSize(agent *agentsv1.Agent) resource.Quantity {
	if agent.Spec.Runtime != nil && agent.Spec.Runtime.WorkspaceSize != nil {
		return *agent.Spec.Runtime.WorkspaceSize
	}
	if d := r.loadDefaults(); d.WorkspaceSize != "" {
		if q, err := resource.ParseQuantity(d.WorkspaceSize); err == nil {
			return q
		}
	}
	return resource.MustParse(defaultWorkspaceSize)
}

// ignoreNotFound treats an already-deleted object as success, so pruning is idempotent.
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
