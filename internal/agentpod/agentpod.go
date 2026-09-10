// Package agentpod runs one pod for each registered agent.
package agentpod

import (
	"context"
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

// defaultStorageSize is the storage request placed on the EFS PVC when the agent
// does not specify one. The EFS CSI driver ignores the value — EFS is elastic — but the
// Kubernetes API requires a positive storage request.
const defaultStorageSize = "50Gi"

const (
	LabelAgent = "agents.czi.team/agent"

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

	agentDataMountPath = "/workspace"

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

	userClaudeConfigVolume    = "user-claude-config"
	userClaudeConfigMountPath = "/etc/agent-user-config"
)

// Config is the operator-level policy for running agent pods.
type Config struct {
	// Namespace is where agents run. It is the operator's own namespace, so owner
	// references garbage-collect an agent's objects when the agent is deleted.
	Namespace string
	// DefaultsLoader reads live defaults from the agent-defaults ConfigMap. When set its
	// values take precedence over the static fields below. CRD spec values always win.
	DefaultsLoader *agentdefaults.Loader
	// DefaultImage is the agent image used when spec.runtime.image is unset and the
	// ConfigMap loader does not provide one.
	DefaultImage string
	// DefaultCommand is the command an agent runs when neither the agent nor the image
	// provides a long-running entrypoint. Without it a base image whose entrypoint exits
	// leaves the pod crash-looping.
	DefaultCommand []string
	// StorageClass is the storage class the per-agent PVC is provisioned from.
	// It must be a ReadWriteMany class backed by the EFS CSI driver.
	StorageClass string
	// Region is the AWS region written into the rendered AWS config.
	Region string
	// AnthropicFederationRuleID, AnthropicOrganizationID, AnthropicServiceAccountID, and
	// AnthropicTokenAudience configure Workload Identity Federation with Anthropic. When all
	// four are non-empty the operator adds a second projected token (audience
	// AnthropicTokenAudience) to every agent pod and sets the four ANTHROPIC_* env vars the
	// Claude SDK and CLI need to exchange it for a Claude access token. When any field is
	// empty the Anthropic token and env vars are omitted, so the operator degrades gracefully
	// in clusters that have not yet configured Claude WIF.
	AnthropicFederationRuleID string
	AnthropicOrganizationID   string
	AnthropicServiceAccountID string
	AnthropicTokenAudience    string

	// GitHubAppID, GitHubAppInstallationID and GitHubAppPrivateKeySecret configure the shared
	// GitHub App every agent clones and opens pull requests as. When all three are non-empty
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

// Reconciler drives an agent pod toward the spec.
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
	return &Reconciler{Client: c, Scheme: scheme, Config: cfg}
}

// Reconcile brings the agent pod in line with its spec.
func (r *Reconciler) Reconcile(ctx context.Context, agent *agentsv1.Agent) (*agentsv1.RuntimeStatus, error) {
	if agent.Spec.Runtime == nil {
		return nil, r.pruneRuntime(ctx, agent)
	}

	err := r.pruneObsoleteRuntimeObjects(ctx, agent)
	if err != nil {
		return nil, err
	}

	err = r.ensureService(ctx, agent)
	if err != nil {
		return nil, err
	}
	err = r.ensureAWSConfig(ctx, agent)
	if err != nil {
		return nil, err
	}
	err = r.ensureClaudeConfig(ctx, agent)
	if err != nil {
		return nil, err
	}
	err = r.ensureStorage(ctx, agent)
	if err != nil {
		return nil, err
	}

	status := &agentsv1.RuntimeStatus{
		ServiceAccountName: agent.ServiceAccountName(),
		StatefulSetName:    agent.StatefulSetName(),
		State:              agentsv1.RuntimeStatePending,
	}
	err = r.ensureServiceAccount(ctx, agent)
	if err != nil {
		status.State = agentsv1.RuntimeStateFailed
		status.Message = err.Error()
		return status, err
	}

	status.MemoryImports, err = r.reconcileMemoryImports(ctx, agent)
	if err != nil {
		status.State = agentsv1.RuntimeStateFailed
		status.Message = err.Error()
		return status, err
	}

	set, err := r.ensureStatefulSet(ctx, agent)
	if err != nil {
		status.State = agentsv1.RuntimeStateFailed
		status.Message = err.Error()
		return status, err
	}

	status.ReadyReplicas = set.Status.ReadyReplicas
	switch {
	case agent.Spec.Runtime.Suspended:
		status.State = agentsv1.RuntimeStateSuspended
	case set.Status.ReadyReplicas > 0:
		status.State = agentsv1.RuntimeStateRunning
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
		// The agent serves no traffic yet. A port is declared anyway because a headless
		// service with no ports publishes no DNS records for its pods.
		service.Spec.Ports = []corev1.ServicePort{{Name: "agent", Port: 8080}}
		return controllerutil.SetControllerReference(agent, service, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service %s: %w", service.Name, err)
	}
	return nil
}

// ensureAWSConfig writes the AWS config mounted by the agent pod.
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

func (r *Reconciler) ensureClaudeConfig(ctx context.Context, agent *agentsv1.Agent) error {
	claudeMD := r.loadDefaults().ClaudeMD
	settingsJSON := "{}\n"
	if agent.Spec.Claude != nil {
		claudeMD = agent.Spec.Claude.ClaudeMD
		if agent.Spec.Claude.SettingsJSON != "" {
			settingsJSON = agent.Spec.Claude.SettingsJSON
		}
	}

	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.ClaudeConfigMapName(),
		Namespace: r.Namespace,
	}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Labels = agentLabels(agent)
		configMap.Data = map[string]string{
			"CLAUDE.md":     claudeMD,
			"settings.json": settingsJSON,
		}
		return controllerutil.SetControllerReference(agent, configMap, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring config map %s: %w", configMap.Name, err)
	}
	return nil
}

// ensureServiceAccount creates the identity the agent pod runs as. Its name is what the
// agent's IAM roles trust, so the pod can assume them with its projected token.
func (r *Reconciler) ensureServiceAccount(ctx context.Context, agent *agentsv1.Agent) error {
	account := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.ServiceAccountName(),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, account, func() error {
		account.Labels = agentLabels(agent)
		return controllerutil.SetControllerReference(agent, account, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service account %s: %w", account.Name, err)
	}
	return nil
}

// ensureStorage creates the ReadWriteMany PVC that stores the agent's working directory.
//
// The PVC is owned by the Agent, so it is garbage-collected when the agent is deleted.
func (r *Reconciler) ensureStorage(ctx context.Context, agent *agentsv1.Agent) error {
	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.PersistentVolumeClaimName(),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.Labels = agentLabels(agent)
		if pvc.Spec.AccessModes == nil {
			pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
			pvc.Spec.StorageClassName = ptr(r.storageClass(agent))
			pvc.Spec.Resources = corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: r.storageSize(agent)},
			}
		}
		return controllerutil.SetControllerReference(agent, pvc, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring storage %s: %w", pvc.Name, err)
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

func (r *Reconciler) storageClass(agent *agentsv1.Agent) string {
	if agent.Spec.Runtime != nil && agent.Spec.Runtime.StorageClass != "" {
		return agent.Spec.Runtime.StorageClass
	}
	if d := r.loadDefaults(); d.StorageClass != "" {
		return d.StorageClass
	}
	return r.StorageClass
}

func (r *Reconciler) storageSize(agent *agentsv1.Agent) resource.Quantity {
	if agent.Spec.Runtime != nil && agent.Spec.Runtime.StorageSize != nil {
		return *agent.Spec.Runtime.StorageSize
	}
	if d := r.loadDefaults(); d.StorageSize != "" {
		if q, err := resource.ParseQuantity(d.StorageSize); err == nil {
			return q
		}
	}
	return resource.MustParse(defaultStorageSize)
}

// ignoreNotFound treats an already-deleted object as success, so pruning is idempotent.
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
