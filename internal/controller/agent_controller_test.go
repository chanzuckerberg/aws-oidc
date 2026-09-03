package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

func TestRuntimeConditionWithoutRuntime(t *testing.T) {
	agent := &agentsv1.Agent{}

	condition := runtimeCondition(agent, nil)
	require.Equal(t, metav1.ConditionTrue, condition.Status)
	require.Equal(t, "NoRuntime", condition.Reason)
}

func TestSetManagedMetadataTracksAgentAsArgoRoot(t *testing.T) {
	agent := &agentsv1.Agent{}
	reconciler := &AgentReconciler{
		ArgoCDTrackingID: "test-app:apps/Deployment:test/test-app-stack-operator",
	}

	require.True(t, reconciler.setManagedMetadata(agent))
	require.Equal(t, reconciler.ArgoCDTrackingID, agent.Annotations[argoCDTrackingIDAnnotation])
	require.Contains(t, agent.Finalizers, agentFinalizer)
	require.False(t, reconciler.setManagedMetadata(agent))
}

// A suspended thread is deliberately idle, so it must not hold the condition false forever.
func TestRuntimeConditionIgnoresSuspendedThreads(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}

	condition := runtimeCondition(agent, []agentsv1.ThreadStatus{
		{Name: "main", State: agentsv1.ThreadStateRunning},
		{Name: "review", State: agentsv1.ThreadStateSuspended},
	})
	require.Equal(t, metav1.ConditionTrue, condition.Status)
	require.Equal(t, "AllThreadsRunning", condition.Reason)
}

func TestRuntimeConditionReportsFailingThread(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}

	condition := runtimeCondition(agent, []agentsv1.ThreadStatus{
		{Name: "main", State: agentsv1.ThreadStateRunning},
		{Name: "review", State: agentsv1.ThreadStateFailed, Message: "agent is limited to 2 threads"},
	})
	require.Equal(t, metav1.ConditionFalse, condition.Status)
	require.Equal(t, "ThreadsPending", condition.Reason)
	require.Contains(t, condition.Message, "thread review is failed")
	require.Contains(t, condition.Message, "limited to 2 threads")
}
