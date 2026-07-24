package v1

import (
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

// Condition types set on an Agent.
const (
	// ConditionReady is true when every grant is Provisioned.
	ConditionReady = "Ready"
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
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=agt,categories=czi
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.ownerEmail`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
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
