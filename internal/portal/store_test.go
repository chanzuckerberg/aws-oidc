package portal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapStore(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	store := NewConfigMapStore(client, "test-ns", "agent-registry", "agents.yaml")

	// Empty registry (ConfigMap does not exist yet).
	agents, err := store.List(ctx)
	require.NoError(t, err)
	require.Empty(t, agents)

	got, err := store.Get(ctx, "missing")
	require.NoError(t, err)
	require.Nil(t, got)

	// Create.
	bot := Agent{
		Name:       "data-bot",
		Owner:      "sub-1",
		OwnerEmail: "a@example.com",
		Grants: []Grant{
			{AccountID: "111", AccountAlias: "prod", RoleARN: "arn:aws:iam::111:role/agents/data-bot-ro", RoleName: "agents/data-bot-ro"},
		},
	}
	require.NoError(t, store.Upsert(ctx, bot))

	got, err = store.Get(ctx, "data-bot")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "sub-1", got.Owner)
	require.Len(t, got.Grants, 1)

	// Owner scoping.
	owned, err := store.ListByOwner(ctx, "sub-1")
	require.NoError(t, err)
	require.Len(t, owned, 1)

	none, err := store.ListByOwner(ctx, "someone-else")
	require.NoError(t, err)
	require.Empty(t, none)

	// A second agent for a different owner does not leak across owners.
	require.NoError(t, store.Upsert(ctx, Agent{Name: "other-bot", Owner: "sub-2"}))
	all, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)
	owned, err = store.ListByOwner(ctx, "sub-1")
	require.NoError(t, err)
	require.Len(t, owned, 1)

	// Update in place.
	bot.Grants = nil
	require.NoError(t, store.Upsert(ctx, bot))
	got, err = store.Get(ctx, "data-bot")
	require.NoError(t, err)
	require.Empty(t, got.Grants)

	// Delete.
	require.NoError(t, store.Delete(ctx, "data-bot"))
	got, err = store.Get(ctx, "data-bot")
	require.NoError(t, err)
	require.Nil(t, got)

	all, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)
}
