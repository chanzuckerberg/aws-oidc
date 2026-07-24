// Package portal implements the agent-registry control plane UI and API: a person logs in,
// sees the AWS access they already have, registers agents, and grants each agent a subset
// of that access.
//
// Agents are stored as Agent custom resources (agents.czi.team/v1), one per agent, through
// the AgentStore seam. The portal writes the desired grants to a CR's spec; the operator
// reconciles them. There is no database and no ConfigMap registry.
package portal

import (
	"context"
	"time"
)

// Grant is a single account plus role an agent may assume. It must always be a subset of
// the owner's own access, enforced when a grant is written. It maps to the AWS section of
// an Agent CR grant.
type Grant struct {
	AccountID    string
	AccountAlias string
	RoleARN      string
	RoleName     string
}

// Key uniquely identifies a grant by account and role.
func (g Grant) Key() string {
	return g.AccountID + "|" + g.RoleARN
}

// Agent is a registered agent and the access granted to it. Owner is the Okta subject of
// the person who registered it.
type Agent struct {
	Name       string
	Owner      string
	OwnerEmail string
	Grants     []Grant
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// AgentStore persists agents.
type AgentStore interface {
	List(ctx context.Context) ([]Agent, error)
	ListByOwner(ctx context.Context, owner string) ([]Agent, error)
	Get(ctx context.Context, name string) (*Agent, error)
	Upsert(ctx context.Context, agent Agent) error
	Delete(ctx context.Context, name string) error
}
