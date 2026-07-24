package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Grant is one unit of desired access: a catalog policy the agent may use in a target
// account. It is set by the owner through the portal and must always be a subset of the
// owner's own access. The operator turns each grant into a per-agent IAM role.
type Grant struct {
	// AccountID is the 12-digit AWS account the grant targets.
	// +kubebuilder:validation:Pattern=`^[0-9]{12}$`
	AccountID string `json:"accountId"`

	// CatalogPolicyID selects one of the curated grantable policies. Free-form policy is
	// never allowed; the value must exist in the catalog and be within the owner's access.
	// +kubebuilder:validation:MinLength=1
	CatalogPolicyID string `json:"catalogPolicyId"`

	// Region is the default AWS region for the agent's profile in this account.
	// +optional
	Region string `json:"region,omitempty"`
}

// AgentSpec is the desired state a human sets through the portal.
type AgentSpec struct {
	// DisplayName is the human-friendly name shown in the portal. Defaults to the
	// resource name when unset.
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Owner is the Okta subject of the person who registered the agent. The portal and
	// the admission webhook stamp and verify this. An agent's access can never exceed the
	// owner's, and one person cannot assume another person's agent roles.
	// +kubebuilder:validation:MinLength=1
	Owner string `json:"owner"`

	// OwnerEmail is the owner's email, kept for display and audit.
	// +optional
	OwnerEmail string `json:"ownerEmail,omitempty"`

	// Grants is the desired set of per-account access.
	// +optional
	// +listType=atomic
	Grants []Grant `json:"grants,omitempty"`
}

// GrantState is the provisioning state of a single grant.
// +kubebuilder:validation:Enum=Pending;Provisioned;Failed
type GrantState string

const (
	// GrantStatePending means the operator has not yet provisioned the grant's IAM role.
	GrantStatePending GrantState = "Pending"
	// GrantStateProvisioned means the per-agent IAM role exists and matches the grant.
	GrantStateProvisioned GrantState = "Provisioned"
	// GrantStateFailed means the operator could not provision the grant; see Message.
	GrantStateFailed GrantState = "Failed"
)

// GrantStatus is what the operator provisioned for one grant. The config server reads
// RoleARN to build the agent's aws-oidc profile.
type GrantStatus struct {
	// AccountID matches the desired grant's account.
	AccountID string `json:"accountId"`

	// CatalogPolicyID matches the desired grant's catalog policy.
	// +optional
	CatalogPolicyID string `json:"catalogPolicyId,omitempty"`

	// RoleARN is the per-agent IAM role the operator created under /agents/.
	// +optional
	RoleARN string `json:"roleArn,omitempty"`

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

// Agent is a registered agent and the scoped AWS access granted to it. The CR is the
// source of truth; there is no separate database.
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
