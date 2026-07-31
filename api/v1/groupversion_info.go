// Package v1 contains the API types for the agent-registry control plane. An Agent
// custom resource is the source of truth for a registered agent and the scoped AWS
// access granted to it. The portal writes Agents and the operator reconciles them into
// per-agent IAM roles.
//
// +kubebuilder:object:generate=true
// +groupName=agents.czi.team
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is the group and version for the Agent API.
var GroupVersion = schema.GroupVersion{Group: "agents.czi.team", Version: "v1"}

// SchemeBuilder registers the Agent types with a runtime scheme. It is built on
// apimachinery alone so the API types can be imported without pulling in the operator's
// controller-runtime dependency (the config server and portal only need the types).
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme adds the Agent types to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion, &Agent{}, &AgentList{})
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
