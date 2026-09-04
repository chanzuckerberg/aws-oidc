package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Provider names a third-party system a grant can target. AWS is supported today; more are
// added as the registry expands to other systems.
const (
	ProviderAWS = "aws"
)

// AWSGrant is access to a single AWS role in an account. It is the same shape the portal
// collects from the owner's existing access: the account and the role the agent may assume.
type AWSGrant struct {
	// AccountID is the 12-digit AWS account the grant targets.
	// +kubebuilder:validation:Pattern=`^[0-9]{12}$`
	AccountID string `json:"accountId"`

	// AccountAlias is the human-friendly account name, kept for display.
	// +optional
	AccountAlias string `json:"accountAlias,omitempty"`

	// RoleARN is the role the agent may assume in the account.
	// +kubebuilder:validation:MinLength=1
	RoleARN string `json:"roleArn"`

	// RoleName is the role name derived from the ARN, kept for display.
	// +optional
	RoleName string `json:"roleName,omitempty"`

	// Region is the default AWS region for the agent's profile in this account.
	// +optional
	Region string `json:"region,omitempty"`
}

// Grant is one unit of access granted to an agent. It is a union: exactly one provider
// section is set. AWS is the only provider today; other third parties (for example Jira,
// GitHub, Slack, Google Workspace) are added as sibling fields here so an agent can hold
// scoped access across many systems in one resource.
//
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type Grant struct {
	// AWS grants the agent a role in an AWS account.
	// +optional
	AWS *AWSGrant `json:"aws,omitempty"`

	// Additional providers go here as pointer fields, for example:
	//   Jira   *JiraGrant   `json:"jira,omitempty"`
	//   GitHub *GitHubGrant `json:"github,omitempty"`
	// Each new provider keeps the union invariant (exactly one section set) via the
	// MinProperties/MaxProperties markers above.
}

// TailscaleAccess enrolls the agent's thread pods in the tailnet and pins the SSH login
// name they may use. Presence of this field means tailnet access is enabled for the agent.
// The portal derives SSHUser from the owner's email local part and never allows root.
type TailscaleAccess struct {
	// SSHUser is the only Linux username the agent may use when connecting to tailnet devices
	// over SSH. The portal fixes this to the owner's email local part.
	// +kubebuilder:validation:Pattern=`^[a-z_][a-z0-9_-]*$`
	// +kubebuilder:validation:MinLength=1
	SSHUser string `json:"sshUser"`
}

// Repository is a GitHub repository in "owner/repo" form that the agent clones into its
// workspace at boot. The portal validates each entry is reachable by the agent's GitHub App
// installations before saving; the pattern here is a second line of defense for writes that
// bypass the portal, and keeps the value safe to hand to a shell as a single argument.
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`
// +kubebuilder:validation:MaxLength=128
type Repository string

// AgentEnvVar is a plain environment variable for an agent's containers. It deliberately has
// no valueFrom: an agent owner must not be able to project arbitrary namespace secrets into
// their own pod.
type AgentEnvVar struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +optional
	Value string `json:"value,omitempty"`
}

// AgentRuntime is the shape every thread of an agent runs as. It is a curated subset of a
// pod spec rather than an embedded PodSpec: the operator owns the service account, the
// projected token volume, and the AWS config mount, so an owner cannot point their pod at
// another identity or mount a host path.
type AgentRuntime struct {
	// Image overrides the operator's default agent image.
	// +optional
	Image string `json:"image,omitempty"`

	// Command overrides the image entrypoint.
	// +optional
	// +listType=atomic
	Command []string `json:"command,omitempty"`

	// Args overrides the image arguments.
	// +optional
	// +listType=atomic
	Args []string `json:"args,omitempty"`

	// Env is extra environment for the agent container.
	// +optional
	// +listType=map
	// +listMapKey=name
	Env []AgentEnvVar `json:"env,omitempty"`

	// Resources is the compute the container requests. The portal sets requests and limits
	// to the same values so the pod is Guaranteed.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector pins the pods to a class of node.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// StorageClass overrides the operator's default storage class for the agent's workspace
	// PVC. Must be a ReadWriteMany class; on EFS the value is a billing-only placeholder and
	// the filesystem grows without bound.
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

	// WorkspaceSize is the storage requested for the agent's shared workspace PVC. On EFS
	// this is a placeholder only; the filesystem is not actually bounded by this value.
	// Defaults to 50Gi when unset.
	// +optional
	WorkspaceSize *resource.Quantity `json:"workspaceSize,omitempty"`
}

// AgentThread is one running thread of an agent: its own pod and its own workspace, sharing
// the agent's access. A person runs several threads of the same agent at once.
type AgentThread struct {
	// Name identifies the thread within the agent and names its Kubernetes objects.
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	// +kubebuilder:validation:MaxLength=24
	Name string `json:"name"`

	// DisplayName is the human-friendly label shown in the portal.
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Suspended scales the thread to zero replicas while keeping its workspace intact.
	// +optional
	Suspended bool `json:"suspended,omitempty"`
}

// AgentSpec is the desired state a human sets through the portal.
type AgentSpec struct {
	// DisplayName is the human-friendly name shown in the portal. Defaults to the
	// resource name when unset.
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Owner is the Okta subject of the person who registered the agent. The portal and
	// the admission webhook stamp and verify this. An agent's access can never exceed the
	// owner's, and one person cannot assume another person's agent access.
	// +kubebuilder:validation:MinLength=1
	Owner string `json:"owner"`

	// OwnerEmail is the owner's email, kept for display and audit.
	// +optional
	OwnerEmail string `json:"ownerEmail,omitempty"`

	// Grants is the desired set of access, one entry per provider target.
	// +optional
	// +listType=atomic
	Grants []Grant `json:"grants,omitempty"`

	// Repositories is the set of GitHub repositories, each "owner/repo", to clone into
	// /workspace when a thread pod boots, so sessions find the source already checked out.
	// Clones are idempotent: a repository already present is left as is. Only meaningful when
	// Runtime is set, since it is the thread pods that clone.
	// +optional
	// +listType=set
	Repositories []Repository `json:"repositories,omitempty"`

	// Tailscale enrolls the agent's pods in the tailnet and fixes the SSH login name to the
	// owner's email local part. When nil the agent has no tailnet identity.
	// +optional
	Tailscale *TailscaleAccess `json:"tailscale,omitempty"`

	// Runtime describes how the agent runs in the cluster. When unset the agent has no pods
	// and exists only as the access granted to it.
	// +optional
	Runtime *AgentRuntime `json:"runtime,omitempty"`

	// Threads is the set of threads to run. Each gets its own pod and workspace. It is a map
	// list so the API server rejects duplicate names and two writers editing different
	// threads do not clobber each other. Ignored when Runtime is unset.
	// +optional
	// +listType=map
	// +listMapKey=name
	Threads []AgentThread `json:"threads,omitempty"`
}

// GrantState is the provisioning state of a single grant.
// +kubebuilder:validation:Enum=Pending;Provisioned;Failed
type GrantState string

const (
	// GrantStatePending means the operator has not yet provisioned the grant.
	GrantStatePending GrantState = "Pending"
	// GrantStateProvisioned means the grant is provisioned and matches the desired state.
	GrantStateProvisioned GrantState = "Provisioned"
	// GrantStateFailed means the operator could not provision the grant; see Message.
	GrantStateFailed GrantState = "Failed"
)

// AWSGrantStatus is the AWS-specific provisioning result for a grant.
type AWSGrantStatus struct {
	// AccountID echoes the grant's account.
	// +optional
	AccountID string `json:"accountId,omitempty"`

	// RoleARN is the per-agent role the operator provisioned for the agent to assume.
	// +optional
	RoleARN string `json:"roleArn,omitempty"`
}

// GrantStatus is what the operator provisioned for one grant, mirroring the spec union.
type GrantStatus struct {
	// Provider names the third party this grant targets (for example "aws").
	Provider string `json:"provider"`

	// AWS holds AWS provisioning detail when Provider is "aws".
	// +optional
	AWS *AWSGrantStatus `json:"aws,omitempty"`

	// State is the provisioning state of this grant.
	// +optional
	State GrantState `json:"state,omitempty"`

	// Message carries the reason when State is Failed.
	// +optional
	Message string `json:"message,omitempty"`
}

// ThreadState is the running state of one thread.
// +kubebuilder:validation:Enum=Pending;Running;Suspended;Failed
type ThreadState string

const (
	// ThreadStatePending means the thread's objects exist but no pod is ready yet.
	ThreadStatePending ThreadState = "Pending"
	// ThreadStateRunning means the thread has a ready pod.
	ThreadStateRunning ThreadState = "Running"
	// ThreadStateSuspended means the thread is intentionally scaled to zero.
	ThreadStateSuspended ThreadState = "Suspended"
	// ThreadStateFailed means the thread could not be provisioned; see Message.
	ThreadStateFailed ThreadState = "Failed"
)

// ThreadStatus is what the operator provisioned for one thread.
type ThreadStatus struct {
	// Name matches the thread's name in spec.threads.
	Name string `json:"name"`

	// ServiceAccountName is the thread's Kubernetes service account, whose projected token
	// the pod exchanges for the agent's AWS roles.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// StatefulSetName is the workload running the thread.
	// +optional
	StatefulSetName string `json:"statefulSetName,omitempty"`

	// ReadyReplicas is how many of the thread's pods are ready (zero or one).
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// State is the thread's running state.
	// +optional
	State ThreadState `json:"state,omitempty"`

	// Message carries the reason when State is Failed.
	// +optional
	Message string `json:"message,omitempty"`
}

// Condition types set on an Agent.
const (
	// ConditionReady is true when every grant is Provisioned.
	ConditionReady = "Ready"
	// ConditionRuntimeReady is true when every non-suspended thread has a ready pod.
	ConditionRuntimeReady = "RuntimeReady"
)

// AgentStatus is what the operator has provisioned. It is a subresource, so the portal's
// writes to spec never race the operator's writes to status.
type AgentStatus struct {
	// ObservedGeneration is the spec generation the operator last reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the standard Ready condition.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Grants is the per-grant provisioning result, aligned with spec.grants.
	// +optional
	// +listType=atomic
	Grants []GrantStatus `json:"grants,omitempty"`

	// WorkspaceClaimName is the PVC shared by all of the agent's threads. Each thread
	// mounts it at its own subPath, so threads can share files through the shared directory
	// while keeping their own working trees separate.
	// +optional
	WorkspaceClaimName string `json:"workspaceClaimName,omitempty"`

	// Threads is the per-thread provisioning result, one entry per spec.threads entry.
	// +optional
	// +listType=map
	// +listMapKey=name
	Threads []ThreadStatus `json:"threads,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=agt,categories=czi
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.ownerEmail`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Threads",type=string,JSONPath=`.status.threads[*].name`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Agent is a registered agent and the scoped access granted to it. The CR is the source of
// truth; there is no separate database.
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentList is a list of Agents.
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Agent `json:"items"`
}
