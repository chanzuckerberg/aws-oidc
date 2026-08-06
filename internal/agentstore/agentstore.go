// Package agentstore reads and writes Agent custom resources (agents.czi.team/v1), one CR
// per agent. It is the single reader shared by the portal (which CRUDs agents a person
// registers) and the config server (which reads a person's agents to build their scoped AWS
// config). Both speak the api/v1 types directly, so there is no lossy domain model in
// between and status is available to both.
//
// It uses the dynamic client so callers need neither a generated clientset nor the
// operator's controller-runtime dependency.
package agentstore

import (
	"context"
	"fmt"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// agentGVR is the Agent custom resource this store reads and writes.
var agentGVR = agentsv1.GroupVersion.WithResource("agents")

// AgentStore reads and writes agents. The portal depends on the full interface; the config
// server uses only ListByOwner.
type AgentStore interface {
	List(ctx context.Context) ([]agentsv1.Agent, error)
	ListByOwner(ctx context.Context, owner string) ([]agentsv1.Agent, error)
	Get(ctx context.Context, name string) (*agentsv1.Agent, error)
	Upsert(ctx context.Context, agent *agentsv1.Agent) error
	Delete(ctx context.Context, name string) error
}

// Store is an AgentStore backed by Agent custom resources in one namespace.
type Store struct {
	client    dynamic.Interface
	namespace string
}

var _ AgentStore = (*Store)(nil)

// New returns a Store scoped to the given namespace.
func New(client dynamic.Interface, namespace string) *Store {
	return &Store{client: client, namespace: namespace}
}

func (s *Store) resource() dynamic.ResourceInterface {
	return s.client.Resource(agentGVR).Namespace(s.namespace)
}

// List returns every agent in the namespace, sorted by name. When the Agent CRD is not
// installed it returns no agents rather than an error, so a config server in an environment
// without the registry (for example prod today) still serves cleanly.
func (s *Store) List(ctx context.Context) ([]agentsv1.Agent, error) {
	list, err := s.resource().List(ctx, metav1.ListOptions{})
	if err != nil {
		if notInstalled(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing agents: %w", err)
	}

	agents := make([]agentsv1.Agent, 0, len(list.Items))
	for i := range list.Items {
		cr, convErr := fromUnstructured(&list.Items[i])
		if convErr != nil {
			return nil, convErr
		}
		agents = append(agents, *cr)
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Name < agents[j].Name })
	return agents, nil
}

// ListByOwner returns the agents whose spec.owner matches the given Okta subject.
func (s *Store) ListByOwner(ctx context.Context, owner string) ([]agentsv1.Agent, error) {
	agents, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	owned := make([]agentsv1.Agent, 0, len(agents))
	for i := range agents {
		if agents[i].Spec.Owner == owner {
			owned = append(owned, agents[i])
		}
	}
	return owned, nil
}

// Get returns the named agent, or nil when it does not exist.
func (s *Store) Get(ctx context.Context, name string) (*agentsv1.Agent, error) {
	u, err := s.resource().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting agent %s: %w", name, err)
	}
	return fromUnstructured(u)
}

// Upsert creates the Agent CR or updates its spec in place. Status is a subresource, so this
// spec write never clobbers what the operator provisioned.
func (s *Store) Upsert(ctx context.Context, agent *agentsv1.Agent) error {
	if agent.Namespace == "" {
		agent.Namespace = s.namespace
	}
	if agent.APIVersion == "" {
		agent.APIVersion = agentsv1.GroupVersion.String()
	}
	if agent.Kind == "" {
		agent.Kind = "Agent"
	}

	desired, err := toUnstructured(agent)
	if err != nil {
		return err
	}

	existing, err := s.resource().Get(ctx, agent.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("getting agent %s: %w", agent.Name, err)
		}
		_, err = s.resource().Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating agent %s: %w", agent.Name, err)
		}
		return nil
	}

	desired.SetResourceVersion(existing.GetResourceVersion())
	_, err = s.resource().Update(ctx, desired, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating agent %s: %w", agent.Name, err)
	}
	return nil
}

// Delete removes the named agent. A missing agent is treated as success.
func (s *Store) Delete(ctx context.Context, name string) error {
	err := s.resource().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting agent %s: %w", name, err)
	}
	return nil
}

// notInstalled reports whether the error means the Agent CRD is absent from the cluster,
// which List treats as "no agents" rather than a failure.
func notInstalled(err error) bool {
	return apierrors.IsNotFound(err) || meta.IsNoMatchError(err)
}

// ToUnstructured converts an Agent CR to the unstructured shape the dynamic client uses.
func ToUnstructured(cr *agentsv1.Agent) (*unstructured.Unstructured, error) {
	return toUnstructured(cr)
}

// FromUnstructured converts a dynamic-client object back into an Agent CR.
func FromUnstructured(u *unstructured.Unstructured) (*agentsv1.Agent, error) {
	return fromUnstructured(u)
}

func toUnstructured(cr *agentsv1.Agent) (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	if err != nil {
		return nil, fmt.Errorf("converting agent to unstructured: %w", err)
	}
	return &unstructured.Unstructured{Object: obj}, nil
}

func fromUnstructured(u *unstructured.Unstructured) (*agentsv1.Agent, error) {
	cr := &agentsv1.Agent{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr)
	if err != nil {
		return nil, fmt.Errorf("converting unstructured to agent: %w", err)
	}
	return cr, nil
}
