package agentpod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const testNamespace = "argus-aws-oidc-rdev"

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, agentsv1.AddToScheme(scheme))
	return scheme
}

func testReconciler(t *testing.T, objects ...client.Object) (*Reconciler, client.Client) {
	t.Helper()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	return New(c, scheme, Config{
		Namespace:    testNamespace,
		DefaultImage: "ubuntu:24.04",
	}), c
}

// testAgent is an agent with one provisioned grant and two threads, one of them suspended.
func testAgent() *agentsv1.Agent {
	agent := &agentsv1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bot",
			Namespace: testNamespace,
			UID:       "0f8fad5b-d9cb-469f-a165-70867728950e",
		},
		Spec: agentsv1.AgentSpec{
			Owner:      "00uSUB123",
			OwnerEmail: "jheath@chanzuckerberg.com",
			Grants: []agentsv1.Grant{{AWS: &agentsv1.AWSGrant{
				AccountID:    "111111111111",
				AccountAlias: "playground",
				RoleARN:      "arn:aws:iam::111111111111:role/readonly",
				RoleName:     "readonly",
			}}},
			Runtime: &agentsv1.AgentRuntime{},
			Threads: []agentsv1.AgentThread{
				{Name: "main"},
				{Name: "review", Suspended: true},
			},
		},
		Status: agentsv1.AgentStatus{
			Grants: []agentsv1.GrantStatus{{
				Provider: "aws",
				State:    agentsv1.GrantStateProvisioned,
				AWS: &agentsv1.AWSGrantStatus{
					AccountID: "111111111111",
					RoleARN:   "arn:aws:iam::111111111111:role/agents/jheath-agent-bot-readonly",
				},
			}},
		},
	}
	return agent
}

func TestReconcileCreatesOneStatefulSetPerThread(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	statuses, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Len(t, statuses, 2)

	// A thread with no ready pod yet is pending; a suspended one reports suspended rather than
	// waiting forever on a pod that will not arrive.
	require.Equal(t, agentsv1.ThreadStatePending, statuses[0].State)
	require.Equal(t, agentsv1.ThreadStateSuspended, statuses[1].State)

	sets := &appsv1.StatefulSetList{}
	require.NoError(t, c.List(ctx, sets, client.InNamespace(testNamespace)))
	require.Len(t, sets.Items, 2)

	byName := map[string]appsv1.StatefulSet{}
	for _, set := range sets.Items {
		byName[set.Name] = set
	}

	main := byName["agent-bot-main"]
	require.Equal(t, int32(1), *main.Spec.Replicas)
	require.Equal(t, "agent-bot", main.Spec.ServiceName)
	// Each thread runs as its own service account, which is what the role trust matches and
	// what makes a thread's AWS activity attributable.
	require.Equal(t, "agent-0f8fad5bd9cb-main", main.Spec.Template.Spec.ServiceAccountName)

	// Suspended means zero replicas, not deleted, so the workspace survives.
	require.Equal(t, int32(0), *byName["agent-bot-review"].Spec.Replicas)

	accounts := &corev1.ServiceAccountList{}
	require.NoError(t, c.List(ctx, accounts, client.InNamespace(testNamespace)))
	require.Len(t, accounts.Items, 2)
}

func TestReconcileMountsTokenAndAWSConfig(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	set := &appsv1.StatefulSet{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-main"}, set))

	pod := set.Spec.Template.Spec
	var projected *corev1.ServiceAccountTokenProjection
	for _, volume := range pod.Volumes {
		if volume.Projected != nil && len(volume.Projected.Sources) == 1 {
			projected = volume.Projected.Sources[0].ServiceAccountToken
		}
	}
	require.NotNil(t, projected, "the token volume is explicit because IRSA's webhook does not inject one")
	require.Equal(t, "sts.amazonaws.com", projected.Audience)

	// The projected token is the pod's only credential, so agent code cannot reach the
	// Kubernetes API.
	require.False(t, *pod.AutomountServiceAccountToken)
	require.False(t, *pod.Containers[0].SecurityContext.AllowPrivilegeEscalation)

	env := map[string]string{}
	for _, e := range pod.Containers[0].Env {
		env[e.Name] = e.Value
	}
	require.Equal(t, awsConfigFilePath, env["AWS_CONFIG_FILE"])
	require.Equal(t, "agent-scoped", env["AWS_PROFILE"])
	require.Equal(t, "main", env["AGENT_THREAD"])

	// The rendered config points every profile at the projected token, and all threads of the
	// agent share it.
	configMap := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-aws-config"}, configMap))
	require.Contains(t, configMap.Data["config"], "web_identity_token_file = "+tokenFilePath)
	require.Contains(t, configMap.Data["config"], "arn:aws:iam::111111111111:role/agents/jheath-agent-bot-readonly")
	require.Contains(t, configMap.Data["config"], "[profile agent-scoped]")
	// Profile names come from the same builder the laptop rendering uses, so a prompt written
	require.Contains(t, configMap.Data["config"], "[profile playground-readonly]")
	// Sessions are named after the agent, so CloudTrail says which agent acted.
	require.Regexp(t, `role_session_name\s+= agent-bot/jheath@chanzuckerberg\.com`, configMap.Data["config"])
}

func TestReconcileIsIdempotent(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	before := &appsv1.StatefulSet{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-main"}, before))

	_, err = r.Reconcile(ctx, agent)
	require.NoError(t, err)

	after := &appsv1.StatefulSet{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-main"}, after))
	require.Equal(t, before.ResourceVersion, after.ResourceVersion, "a steady-state resync must not write")
}

// Removing a thread from the spec has to delete its objects. Owner references do not cover
// this, because the agent itself still exists.
func TestReconcilePrunesRemovedThread(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	agent.Spec.Threads = []agentsv1.AgentThread{{Name: "main"}}
	statuses, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Len(t, statuses, 1)

	sets := &appsv1.StatefulSetList{}
	require.NoError(t, c.List(ctx, sets, client.InNamespace(testNamespace)))
	require.Len(t, sets.Items, 1)
	require.Equal(t, "agent-bot-main", sets.Items[0].Name)

	accounts := &corev1.ServiceAccountList{}
	require.NoError(t, c.List(ctx, accounts, client.InNamespace(testNamespace), client.MatchingLabels{LabelAgent: "bot"}))
	require.Len(t, accounts.Items, 1)
}

// Disabling the runtime removes everything, including the objects the threads shared.
func TestReconcileWithoutRuntimeRemovesEverything(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	agent.Spec.Runtime = nil
	statuses, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Empty(t, statuses)

	sets := &appsv1.StatefulSetList{}
	require.NoError(t, c.List(ctx, sets, client.InNamespace(testNamespace)))
	require.Empty(t, sets.Items)

	services := &corev1.ServiceList{}
	require.NoError(t, c.List(ctx, services, client.InNamespace(testNamespace)))
	require.Empty(t, services.Items)

	configMaps := &corev1.ConfigMapList{}
	require.NoError(t, c.List(ctx, configMaps, client.InNamespace(testNamespace)))
	require.Empty(t, configMaps.Items)
}

func TestReconcileRefusesThreadsBeyondTheLimit(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	agent.Spec.Threads = []agentsv1.AgentThread{{Name: "a"}, {Name: "b"}, {Name: "c"}}

	r, c := testReconciler(t, agent)
	r.MaxThreads = 2

	statuses, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Len(t, statuses, 3)
	require.Equal(t, agentsv1.ThreadStateFailed, statuses[2].State)
	require.Contains(t, statuses[2].Message, "limited to 2 threads")

	sets := &appsv1.StatefulSetList{}
	require.NoError(t, c.List(ctx, sets, client.InNamespace(testNamespace)))
	require.Len(t, sets.Items, 2)
}

// All threads share one ReadWriteMany PVC. Reconciling two threads creates exactly one claim,
// owned by the agent and not bound to any individual thread.
func TestReconcileCreatesOneWorkspacePerAgent(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	claims := &corev1.PersistentVolumeClaimList{}
	require.NoError(t, c.List(ctx, claims, client.InNamespace(testNamespace)))
	require.Len(t, claims.Items, 1, "one PVC regardless of thread count")

	claim := claims.Items[0]
	require.Equal(t, "agent-bot-workspace", claim.Name)
	require.Equal(t, []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}, claim.Spec.AccessModes)
	require.Equal(t, "efs-agent-workspaces", *claim.Spec.StorageClassName)
	require.Len(t, claim.OwnerReferences, 1)
	require.Equal(t, "bot", claim.OwnerReferences[0].Name, "agent owns the PVC for GC")
}

// Each thread mounts the shared PVC at its own subPath for an isolated working tree, and at
// the shared subPath for cross-thread file exchange.
func TestReconcileThreadsGetIsolatedSubPaths(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	set := &appsv1.StatefulSet{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-main"}, set))

	pod := set.Spec.Template.Spec
	require.Empty(t, set.Spec.VolumeClaimTemplates, "no per-thread claim templates; one shared PVC is used instead")

	var workspaceVolume *corev1.PersistentVolumeClaimVolumeSource
	for _, v := range pod.Volumes {
		if v.Name == agentsv1.WorkspaceVolumeName {
			workspaceVolume = v.PersistentVolumeClaim
		}
	}
	require.NotNil(t, workspaceVolume)
	require.Equal(t, agent.WorkspaceClaimName(), workspaceVolume.ClaimName)

	mounts := map[string]corev1.VolumeMount{}
	for _, m := range pod.Containers[0].VolumeMounts {
		if m.Name == agentsv1.WorkspaceVolumeName {
			mounts[m.MountPath] = m
		}
	}
	require.Equal(t, "threads/main", mounts[workspaceMountPath].SubPath)
	require.Equal(t, "shared", mounts[sharedMountPath].SubPath)

	// Pod security context matches the EFS access point uid/gid so writes land as uid 1000.
	require.Equal(t, int64(1000), *pod.SecurityContext.RunAsUser)
	require.Equal(t, int64(1000), *pod.SecurityContext.FSGroup)
}

// Removing a thread removes its StatefulSet and ServiceAccount but keeps the shared PVC.
// The thread's subdirectory inside the volume is intentionally left behind — silently deleting
// a person's work on a spec edit is worse than leaving it for them to clean up.
func TestReconcileThreadRemovalKeepsWorkspace(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	r, c := testReconciler(t, agent)

	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	agent.Spec.Threads = []agentsv1.AgentThread{{Name: "main"}}
	_, err = r.Reconcile(ctx, agent)
	require.NoError(t, err)

	claims := &corev1.PersistentVolumeClaimList{}
	require.NoError(t, c.List(ctx, claims, client.InNamespace(testNamespace)))
	require.Len(t, claims.Items, 1, "PVC must survive thread removal")

	sets := &appsv1.StatefulSetList{}
	require.NoError(t, c.List(ctx, sets, client.InNamespace(testNamespace)))
	require.Len(t, sets.Items, 1, "removed thread's StatefulSet is pruned")
}

func TestReconcileReportsRunningThread(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	agent.Spec.Threads = []agentsv1.AgentThread{{Name: "main"}}

	r, c := testReconciler(t, agent)
	_, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)

	set := &appsv1.StatefulSet{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: "agent-bot-main"}, set))
	set.Status.ReadyReplicas = 1
	require.NoError(t, c.Status().Update(ctx, set))

	statuses, err := r.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Equal(t, agentsv1.ThreadStateRunning, statuses[0].State)
	require.Equal(t, int32(1), statuses[0].ReadyReplicas)
}
