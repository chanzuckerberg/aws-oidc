package portal

import (
	"context"
	"fmt"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// agentGVR is the Agent custom resource this store reads and writes.
var agentGVR = agentsv1.GroupVersion.WithResource("agents")

// crStore persists agents as Agent custom resources, one CR per agent. It uses the dynamic
// client so the portal does not need a generated clientset or the operator's
// controller-runtime dependency.
type crStore struct {
	client    dynamic.Interface
	namespace string
}

var _ AgentStore = (*crStore)(nil)

// NewCRStore returns an AgentStore backed by Agent custom resources in the given namespace.
func NewCRStore(client dynamic.Interface, namespace string) *crStore {
	return &crStore{client: client, namespace: namespace}
}

func (s *crStore) resource() dynamic.ResourceInterface {
	return s.client.Resource(agentGVR).Namespace(s.namespace)
}

func (s *crStore) List(ctx context.Context) ([]Agent, error) {
	list, err := s.resource().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}

	agents := make([]Agent, 0, len(list.Items))
	for i := range list.Items {
		cr, err := agentFromUnstructured(&list.Items[i])
		if err != nil {
			return nil, err
		}
		agents = append(agents, agentFromCR(cr))
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Name < agents[j].Name })
	return agents, nil
}

func (s *crStore) ListByOwner(ctx context.Context, owner string) ([]Agent, error) {
	agents, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	owned := make([]Agent, 0, len(agents))
	for _, a := range agents {
		if a.Owner == owner {
			owned = append(owned, a)
		}
	}
	return owned, nil
}

func (s *crStore) Get(ctx context.Context, name string) (*Agent, error) {
	u, err := s.resource().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting agent %s: %w", name, err)
	}
	cr, err := agentFromUnstructured(u)
	if err != nil {
		return nil, err
	}
	a := agentFromCR(cr)
	return &a, nil
}

// Upsert creates the Agent CR or updates its spec in place. Status is a subresource, so this
// spec write never clobbers what the operator provisioned.
func (s *crStore) Upsert(ctx context.Context, agent Agent) error {
	desired, err := agentToUnstructured(agentToCR(agent, s.namespace))
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

func (s *crStore) Delete(ctx context.Context, name string) error {
	err := s.resource().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting agent %s: %w", name, err)
	}
	return nil
}

// agentToCR maps the portal's domain agent to an Agent CR, putting each grant in the AWS
// section of the CR grant union.
func agentToCR(a Agent, namespace string) *agentsv1.Agent {
	grants := make([]agentsv1.Grant, 0, len(a.Grants))
	for _, g := range a.Grants {
		grants = append(grants, agentsv1.Grant{
			AWS: &agentsv1.AWSGrant{
				AccountID:    g.AccountID,
				AccountAlias: g.AccountAlias,
				RoleARN:      g.RoleARN,
				RoleName:     g.RoleName,
			},
		})
	}
	return &agentsv1.Agent{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentsv1.GroupVersion.String(),
			Kind:       "Agent",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: namespace,
		},
		Spec: agentsv1.AgentSpec{
			DisplayName: a.Name,
			Owner:       a.Owner,
			OwnerEmail:  a.OwnerEmail,
			Grants:      grants,
		},
	}
}

// agentFromCR maps an Agent CR back to the portal's domain agent, reading the AWS section of
// each grant. Grants without an AWS section are skipped since the portal only renders AWS.
func agentFromCR(cr *agentsv1.Agent) Agent {
	grants := make([]Grant, 0, len(cr.Spec.Grants))
	for _, g := range cr.Spec.Grants {
		if g.AWS == nil {
			continue
		}
		grants = append(grants, Grant{
			AccountID:    g.AWS.AccountID,
			AccountAlias: g.AWS.AccountAlias,
			RoleARN:      g.AWS.RoleARN,
			RoleName:     g.AWS.RoleName,
		})
	}
	return Agent{
		Name:       cr.Name,
		Owner:      cr.Spec.Owner,
		OwnerEmail: cr.Spec.OwnerEmail,
		Grants:     grants,
		CreatedAt:  cr.CreationTimestamp.Time,
	}
}

func agentToUnstructured(cr *agentsv1.Agent) (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	if err != nil {
		return nil, fmt.Errorf("converting agent to unstructured: %w", err)
	}
	return &unstructured.Unstructured{Object: obj}, nil
}

func agentFromUnstructured(u *unstructured.Unstructured) (*agentsv1.Agent, error) {
	cr := &agentsv1.Agent{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, cr)
	if err != nil {
		return nil, fmt.Errorf("converting unstructured to agent: %w", err)
	}
	return cr, nil
}
