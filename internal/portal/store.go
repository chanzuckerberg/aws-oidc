// Package portal implements the agent-registry control plane UI and API: a person logs
// in, sees the AWS access they already have, registers agents, and grants each agent a
// subset of that access.
//
// This is the hackathon iteration. Agents are stored as YAML in a single ConfigMap and
// there is no operator yet, so grants are recorded but not provisioned into IAM. The
// plan's target is Agent custom resources reconciled by an operator; the AgentStore
// interface is the seam that swap targets later.
package portal

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"gopkg.in/yaml.v2"
)

// Grant is a single account plus role an agent may assume. It must always be a subset of
// the owner's own access, enforced when a grant is written.
type Grant struct {
	AccountID    string `yaml:"account_id"`
	AccountAlias string `yaml:"account_alias"`
	RoleARN      string `yaml:"role_arn"`
	RoleName     string `yaml:"role_name"`
}

// Key uniquely identifies a grant by account and role.
func (g Grant) Key() string {
	return g.AccountID + "|" + g.RoleARN
}

// Agent is a registered agent and the access granted to it. Owner is the Okta subject of
// the person who registered it.
type Agent struct {
	Name       string    `yaml:"name"`
	Owner      string    `yaml:"owner"`
	OwnerEmail string    `yaml:"owner_email"`
	Grants     []Grant   `yaml:"grants"`
	CreatedAt  time.Time `yaml:"created_at"`
	UpdatedAt  time.Time `yaml:"updated_at"`
}

// AgentStore persists agents.
type AgentStore interface {
	List(ctx context.Context) ([]Agent, error)
	ListByOwner(ctx context.Context, owner string) ([]Agent, error)
	Get(ctx context.Context, name string) (*Agent, error)
	Upsert(ctx context.Context, agent Agent) error
	Delete(ctx context.Context, name string) error
}

// configMapStore stores the whole registry as YAML under one key in one ConfigMap.
type configMapStore struct {
	client    kubernetes.Interface
	namespace string
	name      string
	key       string
}

var _ AgentStore = (*configMapStore)(nil)

// NewConfigMapStore returns an AgentStore backed by a single ConfigMap.
func NewConfigMapStore(client kubernetes.Interface, namespace, name, key string) *configMapStore {
	return &configMapStore{client: client, namespace: namespace, name: name, key: key}
}

type registryFile struct {
	Agents []Agent `yaml:"agents"`
}

// read returns the current ConfigMap (nil if it does not exist yet) and the agents in it.
func (s *configMapStore) read(ctx context.Context) (*corev1.ConfigMap, []Agent, error) {
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("getting configmap %s/%s: %w", s.namespace, s.name, err)
	}

	raw, ok := cm.Data[s.key]
	if !ok || raw == "" {
		return cm, nil, nil
	}

	file := registryFile{}
	err = yaml.Unmarshal([]byte(raw), &file)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshalling agent registry %s/%s key %q: %w", s.namespace, s.name, s.key, err)
	}
	return cm, file.Agents, nil
}

func (s *configMapStore) List(ctx context.Context) ([]Agent, error) {
	_, agents, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].Name < agents[j].Name })
	return agents, nil
}

func (s *configMapStore) ListByOwner(ctx context.Context, owner string) ([]Agent, error) {
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

func (s *configMapStore) Get(ctx context.Context, name string) (*Agent, error) {
	agents, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range agents {
		if agents[i].Name == name {
			found := agents[i]
			return &found, nil
		}
	}
	return nil, nil
}

func (s *configMapStore) write(ctx context.Context, cm *corev1.ConfigMap, agents []Agent) error {
	sort.Slice(agents, func(i, j int) bool { return agents[i].Name < agents[j].Name })
	data, err := yaml.Marshal(registryFile{Agents: agents})
	if err != nil {
		return fmt.Errorf("marshalling agent registry: %w", err)
	}

	configMaps := s.client.CoreV1().ConfigMaps(s.namespace)
	if cm == nil {
		_, err = configMaps.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: s.name, Namespace: s.namespace},
			Data:       map[string]string{s.key: string(data)},
		}, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating configmap %s/%s: %w", s.namespace, s.name, err)
		}
		return nil
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[s.key] = string(data)
	_, err = configMaps.Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating configmap %s/%s: %w", s.namespace, s.name, err)
	}
	return nil
}

func (s *configMapStore) Upsert(ctx context.Context, agent Agent) error {
	cm, agents, err := s.read(ctx)
	if err != nil {
		return err
	}

	found := false
	for i := range agents {
		if agents[i].Name == agent.Name {
			agents[i] = agent
			found = true
			break
		}
	}
	if !found {
		agents = append(agents, agent)
	}
	return s.write(ctx, cm, agents)
}

func (s *configMapStore) Delete(ctx context.Context, name string) error {
	cm, agents, err := s.read(ctx)
	if err != nil {
		return err
	}

	kept := make([]Agent, 0, len(agents))
	for _, a := range agents {
		if a.Name != name {
			kept = append(kept, a)
		}
	}
	return s.write(ctx, cm, kept)
}
