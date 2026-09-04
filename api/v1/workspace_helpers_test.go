package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func agentNamed(name string, uid types.UID) *Agent {
	return &Agent{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "agents", UID: uid}}
}

// The IAM trust condition matches workspace service accounts with a wildcard, so one agent's
// prefix must never be a prefix of another's. A name-based prefix fails this: agent "foo-bar"
// would produce names matching agent "foo"'s pattern, handing it another owner's access.
func TestWorkspaceSubjectPatternsDoNotOverlap(t *testing.T) {
	foo := agentNamed("foo", "0f8fad5b-d9cb-469f-a165-70867728950e")
	fooBar := agentNamed("foo-bar", "7c9e6679-7425-40de-944b-e07fc1f90ae7")

	pattern := strings.TrimSuffix(foo.WorkspaceSubjectPattern("agents"), "*")
	require.NotEmpty(t, pattern)

	require.True(t, strings.HasPrefix("system:serviceaccount:agents:"+foo.WorkspaceServiceAccountName("main"), pattern))
	require.False(t, strings.HasPrefix("system:serviceaccount:agents:"+fooBar.WorkspaceServiceAccountName("main"), pattern),
		"another agent's workspace must not match this agent's subject pattern")
}

// Without a uid (a CR the API server has not assigned one to yet) the prefix still has to be
// stable and unique, or two agents would share a trust scope.
func TestIdentityPrefixFallsBackToHash(t *testing.T) {
	a := agentNamed("foo", "")
	b := agentNamed("bar", "")

	require.Equal(t, a.identityPrefix(), agentNamed("foo", "").identityPrefix())
	require.NotEqual(t, a.identityPrefix(), b.identityPrefix())
}

func TestObjectNamesStayValidLabels(t *testing.T) {
	long := agentNamed(strings.Repeat("verylongagentname", 6), "0f8fad5b-d9cb-469f-a165-70867728950e")

	// A workspace's pod is named <statefulset>-0 and has to remain a DNS label.
	sts := long.WorkspaceStatefulSetName("main")
	require.LessOrEqual(t, len(sts+"-0"), 63)
	require.LessOrEqual(t, len(long.WorkspaceServiceAccountName("main")), 63)
	require.LessOrEqual(t, len(long.ServiceName()), 63)
	require.LessOrEqual(t, len(long.AWSConfigMapName()), 63)

	// A dot is legal in a CR name but breaks the per-pod DNS records the headless service
	// publishes, so it is replaced.
	dotted := agentNamed("my.agent", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "agent-my-agent-main", dotted.WorkspaceStatefulSetName("main"))
	require.NotContains(t, dotted.ServiceName(), ".")
}

func TestWorkspaceClaimName(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "agent-bot-workspace", agent.WorkspaceClaimName())
}

func TestWorkspaceSubPath(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "workspaces/main", agent.WorkspaceSubPath("main"))
	require.Equal(t, "workspaces/my-workspace", agent.WorkspaceSubPath("my-workspace"))
}

func TestSharedWorkspaceSubPath(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "shared", agent.SharedWorkspaceSubPath())
}

func TestWorkspaceLookup(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	agent.Spec.Workspaces = []AgentWorkspace{{Name: "main"}, {Name: "review", Suspended: true}}

	require.Nil(t, agent.Workspace("missing"))
	require.True(t, agent.Workspace("review").Suspended)
}
