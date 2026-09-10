package agentpod

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

func TestReconcileMemoryImportAppliesOnceAndCleansStaging(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	revision := strings.Repeat("a", 64)
	repository := agentsv1.Repository("chanzuckerberg/aws-oidc")
	agent.Spec.Claude = &agentsv1.ClaudeConfig{MemoryImports: []agentsv1.ProjectMemoryImport{{
		Repository: repository,
		Revision:   revision,
	}}}
	staging := memoryImportConfigMap(agent, repository, revision)
	reconciler, kubeClient := testReconciler(t, agent, staging)

	runtimeStatus, err := reconciler.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Len(t, runtimeStatus.MemoryImports, 1)
	require.Equal(t, agentsv1.ProjectMemoryImportPending, runtimeStatus.MemoryImports[0].State)

	job := &batchv1.Job{}
	jobKey := types.NamespacedName{
		Namespace: testNamespace,
		Name:      agent.MemoryImportJobName(repository, revision),
	}
	require.NoError(t, kubeClient.Get(ctx, jobKey, job))
	require.Equal(t, "arm64", job.Spec.Template.Spec.NodeSelector["kubernetes.io/arch"])
	container := job.Spec.Template.Spec.Containers[0]
	require.Contains(t, container.Args[0], `mv "$incoming" "$destination"`)
	require.Contains(t, container.Args[0], `projectPath="/workspace/$(basename "$checkout")"`)
	require.Equal(t, agentDataMountPath, container.Env[0].Value)
	require.Equal(t, agentDataMountPath, container.VolumeMounts[1].MountPath)
	require.Equal(t, agent.PersistentDataSubPath(), container.VolumeMounts[1].SubPath)
	require.Equal(t, agent.PersistentVolumeClaimName(),
		job.Spec.Template.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)

	job.Status.Conditions = []batchv1.JobCondition{{
		Type:   batchv1.JobComplete,
		Status: corev1.ConditionTrue,
	}}
	require.NoError(t, kubeClient.Status().Update(ctx, job))
	agent.Status.Runtime = &agentsv1.RuntimeStatus{MemoryImports: runtimeStatus.MemoryImports}

	runtimeStatus, err = reconciler.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Equal(t, agentsv1.ProjectMemoryImportApplied, runtimeStatus.MemoryImports[0].State)
	require.NoError(t, kubeClient.Get(ctx, jobKey, &batchv1.Job{}))

	agent.Status.Runtime.MemoryImports = runtimeStatus.MemoryImports
	runtimeStatus, err = reconciler.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.Equal(t, agentsv1.ProjectMemoryImportApplied, runtimeStatus.MemoryImports[0].State)
	require.True(t, apierrors.IsNotFound(kubeClient.Get(ctx, jobKey, &batchv1.Job{})))
	configMapKey := types.NamespacedName{
		Namespace: testNamespace,
		Name:      agent.MemoryImportConfigMapName(repository, revision),
	}
	require.True(t, apierrors.IsNotFound(kubeClient.Get(ctx, configMapKey, &corev1.ConfigMap{})))

	_, err = reconciler.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.True(t, apierrors.IsNotFound(kubeClient.Get(ctx, jobKey, &batchv1.Job{})))
}

func TestReconcileMemoryImportReportsMissingStaging(t *testing.T) {
	agent := testAgent()
	agent.Spec.Claude = &agentsv1.ClaudeConfig{MemoryImports: []agentsv1.ProjectMemoryImport{{
		Repository: "chanzuckerberg/aws-oidc",
		Revision:   strings.Repeat("b", 64),
	}}}
	reconciler, _ := testReconciler(t, agent)

	runtimeStatus, err := reconciler.Reconcile(context.Background(), agent)
	require.NoError(t, err)
	require.Len(t, runtimeStatus.MemoryImports, 1)
	require.Equal(t, agentsv1.ProjectMemoryImportFailed, runtimeStatus.MemoryImports[0].State)
	require.Contains(t, runtimeStatus.MemoryImports[0].Message, "missing")
}

func TestReconcileMemoryImportPrunesSupersededRevision(t *testing.T) {
	ctx := context.Background()
	agent := testAgent()
	repository := agentsv1.Repository("chanzuckerberg/aws-oidc")
	oldRevision := strings.Repeat("c", 64)
	newRevision := strings.Repeat("d", 64)
	agent.Spec.Claude = &agentsv1.ClaudeConfig{MemoryImports: []agentsv1.ProjectMemoryImport{{
		Repository: repository,
		Revision:   newRevision,
	}}}
	oldConfigMap := memoryImportConfigMap(agent, repository, oldRevision)
	oldJob := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{
		Name:      agent.MemoryImportJobName(repository, oldRevision),
		Namespace: testNamespace,
		Labels: map[string]string{
			LabelAgent:        agent.Name,
			memoryImportLabel: "true",
		},
	}}
	newConfigMap := memoryImportConfigMap(agent, repository, newRevision)
	reconciler, kubeClient := testReconciler(t, agent, oldConfigMap, oldJob, newConfigMap)

	_, err := reconciler.Reconcile(ctx, agent)
	require.NoError(t, err)
	require.True(t, apierrors.IsNotFound(kubeClient.Get(ctx, client.ObjectKeyFromObject(oldConfigMap), &corev1.ConfigMap{})))
	require.True(t, apierrors.IsNotFound(kubeClient.Get(ctx, client.ObjectKeyFromObject(oldJob), &batchv1.Job{})))
	require.NoError(t, kubeClient.Get(ctx, client.ObjectKeyFromObject(newConfigMap), &corev1.ConfigMap{}))
}

func memoryImportConfigMap(agent *agentsv1.Agent, repository agentsv1.Repository, revision string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.MemoryImportConfigMapName(repository, revision),
			Namespace: testNamespace,
			Labels: map[string]string{
				LabelAgent:        agent.Name,
				memoryImportLabel: "true",
			},
		},
		Data: map[string]string{"MEMORY.md": "# Memory\n"},
	}
}
