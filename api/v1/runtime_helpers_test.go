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

func TestServiceAccountSubjectsDoNotOverlap(t *testing.T) {
	foo := agentNamed("foo", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fooBar := agentNamed("foo-bar", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	require.NotEqual(t, foo.ServiceAccountSubject("agents"), fooBar.ServiceAccountSubject("agents"))
	require.Equal(t, "system:serviceaccount:agents:remote-agent-aaaaaaaaaaaa", foo.ServiceAccountSubject("agents"))
}

func TestRuntimeObjectNamesFitKubernetesLimits(t *testing.T) {
	long := agentNamed(strings.Repeat("very-long-agent-name-", 8), "")

	require.LessOrEqual(t, len(long.StatefulSetName()), 61)
	require.LessOrEqual(t, len(long.ServiceAccountName()), 63)
	require.LessOrEqual(t, len(long.PersistentVolumeClaimName()), 63)
	require.LessOrEqual(t, len(long.AWSConfigMapName()), 63)
	require.LessOrEqual(t, len(long.ServiceName()), 63)
	require.Equal(t, long.StatefulSetName(), agentNamed(long.Name, "").StatefulSetName())
}

func TestRuntimeObjectNamesSanitizeAgentName(t *testing.T) {
	agent := agentNamed("My.Agent", "")

	require.Equal(t, "agent-my-agent", agent.StatefulSetName())
	require.Equal(t, "agent-my-agent-workspace", agent.PersistentVolumeClaimName())
	require.Equal(t, "workspaces/main", agent.PersistentDataSubPath())
}
