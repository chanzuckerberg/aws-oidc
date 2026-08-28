// Package agentpod runs an agent's threads in the cluster. An agent is not a single session:
// a person runs several threads of the same agent at once, so each thread gets its own pod
// while sharing the agent's access and its workspace.
//
// Per agent it maintains a headless Service (so each thread pod has a stable DNS name), a
// ConfigMap holding the rendered AWS config, and a single ReadWriteMany PVC backed by an EFS
// access point. Per thread it maintains a ServiceAccount and a StatefulSet. Each thread pod
// mounts the shared PVC at a thread-specific subPath for its working tree and at a common
// subPath for files shared between threads. The thread's pod exchanges its projected service
// account token for the agent's IAM roles, which trust the cluster's OIDC issuer for exactly
// these service accounts.
package agentpod

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// workspacePlaceholderSize is the storage request we put on the EFS workspace PVC. The EFS
// CSI driver ignores it — EFS is elastic — but the Kubernetes API requires a positive request.
const workspacePlaceholderSize = "1Gi"

const (
	// LabelAgent and LabelThread identify which agent and thread an object belongs to. The
	// reconciler selects on them to find objects whose thread is gone from the spec.
	LabelAgent  = "agents.czi.team/agent"
	LabelThread = "agents.czi.team/thread"

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

	// workspaceMountPath is the thread's persistent working directory.
	workspaceMountPath = "/workspace"

	// tokenAudience is the audience STS requires on a projected token used for web identity.
	tokenAudience = "sts.amazonaws.com"

	// tokenExpirationSeconds is the projected token's lifetime. The kubelet rotates it at 80%
	// of this, and the AWS SDK re-reads the file on each refresh.
	tokenExpirationSeconds int64 = 3600

	agentContainerName = "agent"
	awsConfigVolume    = "aws-config"
	tokenVolume        = "aws-token"
)

// Config is the operator-level policy for running agent threads, the same for every agent.
type Config struct {
	// Namespace is where the threads run. It is the operator's own namespace, so owner
	// references garbage-collect an agent's objects when the agent is deleted.
	Namespace string
	// DefaultImage is the agent image used when spec.runtime.image is unset.
	DefaultImage string
	// DefaultCommand is the command an agent thread runs when neither the agent nor the image
	// provides a long-running entrypoint. Without it a base image whose entrypoint exits
	// leaves the pod crash-looping.
	DefaultCommand []string
	// StorageClass is the storage class the per-agent workspace PVC is provisioned from.
	// It must be a ReadWriteMany class backed by the EFS CSI driver.
	StorageClass string
	// Region is the AWS region written into the rendered AWS config.
	Region string
	// MaxThreads bounds how many threads one agent may run, so a single Agent write cannot
	// ask for an unbounded number of pods.
	MaxThreads int
}

// Reconciler drives an agent's threads toward the spec.
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
	if cfg.MaxThreads <= 0 {
		cfg.MaxThreads = defaultMaxThreads
	}
	return &Reconciler{Client: c, Scheme: scheme, Config: cfg}
}

// defaultMaxThreads is the per-agent thread ceiling when none is configured.
const defaultMaxThreads = 5

// Reconcile brings the agent's threads in line with its spec and returns their statuses,
// aligned with spec.threads. The grant statuses come from the AWS provider and carry the role
// ARNs the pods' AWS config needs, so this runs after the grants are reconciled.
//
// An agent with no runtime has no threads, so everything the agent owns is pruned.
func (r *Reconciler) Reconcile(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.ThreadStatus, error) {
	if agent.Spec.Runtime == nil {
		err := r.pruneThreads(ctx, agent, nil)
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

	threads := agent.Spec.Threads
	statuses := make([]agentsv1.ThreadStatus, 0, len(threads))
	keep := make(map[string]bool, len(threads))
	var errs []error

	for i := range threads {
		thread := threads[i]

		if i >= r.MaxThreads {
			statuses = append(statuses, agentsv1.ThreadStatus{
				Name:    thread.Name,
				State:   agentsv1.ThreadStateFailed,
				Message: fmt.Sprintf("agent is limited to %d threads", r.MaxThreads),
			})
			continue
		}
		keep[thread.Name] = true

		status, err := r.reconcileThread(ctx, agent, thread)
		if err != nil {
			status.State = agentsv1.ThreadStateFailed
			status.Message = err.Error()
			errs = append(errs, err)
		}
		statuses = append(statuses, status)
	}

	// Prune after reconciling, so a thread that failed to come up is not also torn down.
	err = r.pruneThreads(ctx, agent, keep)
	if err != nil {
		errs = append(errs, err)
	}

	return statuses, errors.Join(errs...)
}

// reconcileThread ensures one thread's service account and workload, then reports what it
// found. The shared workspace is created once per agent before any thread is reconciled.
func (r *Reconciler) reconcileThread(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) (agentsv1.ThreadStatus, error) {
	status := agentsv1.ThreadStatus{
		Name:               thread.Name,
		ServiceAccountName: agent.ThreadServiceAccountName(thread.Name),
		StatefulSetName:    agent.ThreadStatefulSetName(thread.Name),
		State:              agentsv1.ThreadStatePending,
	}

	err := r.ensureServiceAccount(ctx, agent, thread)
	if err != nil {
		return status, err
	}

	set, err := r.ensureStatefulSet(ctx, agent, thread)
	if err != nil {
		return status, err
	}

	status.ReadyReplicas = set.Status.ReadyReplicas
	switch {
	case thread.Suspended:
		status.State = agentsv1.ThreadStateSuspended
	case set.Status.ReadyReplicas > 0:
		status.State = agentsv1.ThreadStateRunning
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
		// The threads serve no traffic yet. A port is declared anyway because a headless
		// service with no ports publishes no DNS records for its pods.
		service.Spec.Ports = []corev1.ServicePort{{Name: "agent", Port: 8080}}
		return controllerutil.SetControllerReference(agent, service, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service %s: %w", service.Name, err)
	}
	return nil
}

// ensureAWSConfig writes the agent's rendered AWS config. Every thread mounts the same one,
// since threads differ in workspace, not in access.
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

// ensureServiceAccount creates the identity the thread's pod runs as. Its name is what the
// agent's IAM roles trust, so the pod can assume them with its projected token.
func (r *Reconciler) ensureServiceAccount(ctx context.Context, agent *agentsv1.Agent, thread agentsv1.AgentThread) error {
	account := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.ThreadServiceAccountName(thread.Name),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, account, func() error {
		account.Labels = threadLabels(agent, thread.Name)
		return controllerutil.SetControllerReference(agent, account, r.Scheme)
	})
	if err != nil {
		return fmt.Errorf("ensuring service account %s: %w", account.Name, err)
	}
	return nil
}

// ensureWorkspace creates the shared ReadWriteMany PVC for the agent. All threads mount it:
// each thread at its own subPath for an isolated working tree, and every thread at the shared
// subPath for files passed between threads. The EFS CSI driver provisions a fresh access
// point for this PVC, so no other agent's data is reachable inside it.
//
// The PVC is owned by the Agent, so it is garbage-collected when the agent is deleted.
func (r *Reconciler) ensureWorkspace(ctx context.Context, agent *agentsv1.Agent) error {
	placeholder := resource.MustParse(workspacePlaceholderSize)
	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.WorkspaceClaimName(),
		Namespace: r.Namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.Labels = agentLabels(agent)
		if pvc.Spec.AccessModes == nil {
			pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
			pvc.Spec.StorageClassName = ptr(r.StorageClass)
			pvc.Spec.Resources = corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: placeholder},
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

// threadLabels identifies the objects belonging to one thread of one agent.
func threadLabels(agent *agentsv1.Agent, thread string) map[string]string {
	labels := agentLabels(agent)
	labels[LabelThread] = thread
	return labels
}

func (r *Reconciler) image(agent *agentsv1.Agent) string {
	if agent.Spec.Runtime.Image != "" {
		return agent.Spec.Runtime.Image
	}
	return r.DefaultImage
}

func (r *Reconciler) command(agent *agentsv1.Agent) []string {
	if len(agent.Spec.Runtime.Command) > 0 {
		return agent.Spec.Runtime.Command
	}
	return r.DefaultCommand
}

// ignoreNotFound treats an already-deleted object as success, so pruning is idempotent.
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
