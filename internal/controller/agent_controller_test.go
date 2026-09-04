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

// A suspended workspace is deliberately idle, so it must not hold the condition false forever.
func TestRuntimeConditionIgnoresSuspendedWorkspaces(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}

	condition := runtimeCondition(agent, []agentsv1.WorkspaceStatus{
		{Name: "main", State: agentsv1.WorkspaceStateRunning},
		{Name: "review", State: agentsv1.WorkspaceStateSuspended},
	})
	require.Equal(t, metav1.ConditionTrue, condition.Status)
	require.Equal(t, "AllWorkspacesRunning", condition.Reason)
}

func TestRuntimeConditionReportsFailingWorkspace(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}

	condition := runtimeCondition(agent, []agentsv1.WorkspaceStatus{
		{Name: "main", State: agentsv1.WorkspaceStateRunning},
		{Name: "review", State: agentsv1.WorkspaceStateFailed, Message: "agent is limited to 2 workspaces"},
	})
	require.Equal(t, metav1.ConditionFalse, condition.Status)
	require.Equal(t, "WorkspacesPending", condition.Reason)
	require.Contains(t, condition.Message, "workspace review is failed")
	require.Contains(t, condition.Message, "limited to 2 workspaces")
}
