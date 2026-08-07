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

// The IAM trust condition matches thread service accounts with a wildcard, so one agent's
// prefix must never be a prefix of another's. A name-based prefix fails this: agent "foo-bar"
// would produce names matching agent "foo"'s pattern, handing it another owner's access.
func TestThreadSubjectPatternsDoNotOverlap(t *testing.T) {
	foo := agentNamed("foo", "0f8fad5b-d9cb-469f-a165-70867728950e")
	fooBar := agentNamed("foo-bar", "7c9e6679-7425-40de-944b-e07fc1f90ae7")

	pattern := strings.TrimSuffix(foo.ThreadSubjectPattern("agents"), "*")
	require.NotEmpty(t, pattern)

	require.True(t, strings.HasPrefix("system:serviceaccount:agents:"+foo.ThreadServiceAccountName("main"), pattern))
	require.False(t, strings.HasPrefix("system:serviceaccount:agents:"+fooBar.ThreadServiceAccountName("main"), pattern),
		"another agent's thread must not match this agent's subject pattern")
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

	// A thread's pod is named <statefulset>-0 and has to remain a DNS label.
	sts := long.ThreadStatefulSetName("main")
	require.LessOrEqual(t, len(sts+"-0"), 63)
	require.LessOrEqual(t, len(long.ThreadServiceAccountName("main")), 63)
	require.LessOrEqual(t, len(long.ServiceName()), 63)
	require.LessOrEqual(t, len(long.AWSConfigMapName()), 63)

	// A dot is legal in a CR name but breaks the per-pod DNS records the headless service
	// publishes, so it is replaced.
	dotted := agentNamed("my.agent", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "agent-my-agent-main", dotted.ThreadStatefulSetName("main"))
	require.NotContains(t, dotted.ServiceName(), ".")
}

func TestWorkspaceClaimMatchesStatefulSetOrdinal(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	require.Equal(t, "workspace-agent-bot-main-0", agent.ThreadWorkspaceClaimName("main"))
}

func TestThreadLookup(t *testing.T) {
	agent := agentNamed("bot", "0f8fad5b-d9cb-469f-a165-70867728950e")
	agent.Spec.Threads = []AgentThread{{Name: "main"}, {Name: "review", Suspended: true}}

	require.Nil(t, agent.Thread("missing"))
	require.True(t, agent.Thread("review").Suspended)
}
