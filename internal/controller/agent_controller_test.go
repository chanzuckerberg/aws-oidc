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

func TestRuntimeConditionAcceptsSuspendedAgent(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{Suspended: true}

	condition := runtimeCondition(agent, &agentsv1.RuntimeStatus{State: agentsv1.RuntimeStateSuspended})
	require.Equal(t, metav1.ConditionTrue, condition.Status)
	require.Equal(t, "RuntimeSuspended", condition.Reason)
}

func TestRuntimeConditionReportsFailingAgent(t *testing.T) {
	agent := &agentsv1.Agent{}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}

	condition := runtimeCondition(agent, &agentsv1.RuntimeStatus{
		State:   agentsv1.RuntimeStateFailed,
		Message: "statefulset unavailable",
	})
	require.Equal(t, metav1.ConditionFalse, condition.Status)
	require.Equal(t, "RuntimePending", condition.Reason)
	require.Contains(t, condition.Message, "statefulset unavailable")
}
