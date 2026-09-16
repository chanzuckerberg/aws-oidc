package agentpod

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const (
	memoryImportLabel     = "agents.czi.team/memory-import"
	memoryImportDataMount = "/import"
)

func (r *Reconciler) reconcileMemoryImports(ctx context.Context, agent *agentsv1.Agent) ([]agentsv1.ProjectMemoryImportStatus, error) {
	var imports []agentsv1.ProjectMemoryImport
	if agent.Spec.Claude != nil {
		imports = agent.Spec.Claude.MemoryImports
	}

	statuses := make([]agentsv1.ProjectMemoryImportStatus, 0, len(imports))
	keep := make(map[string]bool, len(imports))
	for _, memoryImport := range imports {
		jobName := agent.MemoryImportJobName(memoryImport.Repository, memoryImport.Revision)
		configMapName := agent.MemoryImportConfigMapName(memoryImport.Repository, memoryImport.Revision)
		keep[jobName] = true
		keep[configMapName] = true

		observed := observedMemoryImport(agent, memoryImport)
		if observed.State == agentsv1.ProjectMemoryImportApplied {
			err := r.deleteMemoryImportArtifacts(ctx, jobName, configMapName)
			if err != nil {
				return nil, err
			}
			delete(keep, jobName)
			delete(keep, configMapName)
			statuses = append(statuses, observed)
			continue
		}

		configMap := &corev1.ConfigMap{}
		err := r.Get(ctx, client.ObjectKey{Namespace: r.Namespace, Name: configMapName}, configMap)
		if apierrors.IsNotFound(err) {
			statuses = append(statuses, agentsv1.ProjectMemoryImportStatus{
				Repository: memoryImport.Repository,
				Revision:   memoryImport.Revision,
				State:      agentsv1.ProjectMemoryImportFailed,
				Message:    "staged memory files are missing",
			})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("getting memory import %s: %w", configMapName, err)
		}

		job, err := r.ensureMemoryImportJob(ctx, agent, memoryImport)
		if err != nil {
			return nil, err
		}
		status := agentsv1.ProjectMemoryImportStatus{
			Repository: memoryImport.Repository,
			Revision:   memoryImport.Revision,
			State:      agentsv1.ProjectMemoryImportPending,
		}
		for _, condition := range job.Status.Conditions {
			switch {
			case condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue:
				status.State = agentsv1.ProjectMemoryImportApplied
			case condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue:
				status.State = agentsv1.ProjectMemoryImportFailed
				status.Message = condition.Message
			}
		}
		statuses = append(statuses, status)
	}

	err := r.pruneMemoryImportArtifacts(ctx, agent, keep)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func observedMemoryImport(agent *agentsv1.Agent, memoryImport agentsv1.ProjectMemoryImport) agentsv1.ProjectMemoryImportStatus {
	if agent.Status.Runtime != nil {
		for _, status := range agent.Status.Runtime.MemoryImports {
			if status.Repository == memoryImport.Repository && status.Revision == memoryImport.Revision {
				return status
			}
		}
	}
	return agentsv1.ProjectMemoryImportStatus{
		Repository: memoryImport.Repository,
		Revision:   memoryImport.Revision,
		State:      agentsv1.ProjectMemoryImportPending,
	}
}

func (r *Reconciler) ensureMemoryImportJob(ctx context.Context, agent *agentsv1.Agent, memoryImport agentsv1.ProjectMemoryImport) (*batchv1.Job, error) {
	name := agent.MemoryImportJobName(memoryImport.Repository, memoryImport.Revision)
	job := &batchv1.Job{}
	err := r.Get(ctx, client.ObjectKey{Namespace: r.Namespace, Name: name}, job)
	if err == nil {
		return job, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting memory import job %s: %w", name, err)
	}

	backoffLimit := int32(1)
	repositoryName := strings.SplitN(string(memoryImport.Repository), "/", 2)[1]
	labels := agentLabels(agent)
	labels[memoryImportLabel] = "true"
	job = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.Namespace, Labels: labels},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					RestartPolicy:                corev1.RestartPolicyNever,
					AutomountServiceAccountToken: ptr(false),
					NodeSelector:                 nodeSelector(agent.Spec.Runtime.NodeSelector),
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: ptr(int64(1000)),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Name:    "memory-import",
						Image:   r.image(agent),
						Command: []string{"/bin/bash", "-c"},
						Args: []string{`set -euo pipefail
checkout="${AGENT_DATA_ROOT}/${REPOSITORY_NAME}"
for attempt in {1..60}; do
  if [[ -d "${checkout}/.git" ]]; then break; fi
  sleep 5
done
[[ -d "${checkout}/.git" ]]
projectPath="/workspace/$(basename "$checkout")"
projectDirectory="$(printf '%s' "$projectPath" | LC_ALL=C tr -c 'A-Za-z0-9-' '-')"
destination="${AGENT_DATA_ROOT}/.claude/projects/${projectDirectory}/memory"
parent="$(dirname "$destination")"
incoming="${parent}/.memory-${MEMORY_REVISION}"
previous="${parent}/.memory-previous-${MEMORY_REVISION}"
restore() {
  if [[ ! -e "$destination" && -e "$previous" ]]; then mv "$previous" "$destination"; fi
  rm -rf "$incoming"
}
trap restore EXIT
rm -rf "$incoming" "$previous"
mkdir -p "$incoming"
cp /import/*.md "$incoming/"
if [[ -d "$destination" ]]; then mv "$destination" "$previous"; fi
mv "$incoming" "$destination"
rm -rf "$previous"
trap - EXIT`},
						Env: []corev1.EnvVar{
							{Name: "AGENT_DATA_ROOT", Value: agentDataMountPath},
							{Name: "MEMORY_REVISION", Value: memoryImport.Revision},
							{Name: "REPOSITORY_NAME", Value: repositoryName},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("32Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: ptr(false),
							Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
							RunAsUser:                ptr(int64(1000)),
							RunAsGroup:               ptr(int64(1000)),
						},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "memory-import", MountPath: memoryImportDataMount, ReadOnly: true},
							{
								Name:      agentsv1.DataVolumeName,
								MountPath: agentDataMountPath,
								SubPath:   agent.PersistentDataSubPath(),
							},
						},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "memory-import",
							VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: agent.MemoryImportConfigMapName(memoryImport.Repository, memoryImport.Revision),
								},
							}},
						},
						{
							Name: agentsv1.DataVolumeName,
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: agent.PersistentVolumeClaimName(),
							}},
						},
					},
				},
			},
		},
	}
	err = controllerutil.SetControllerReference(agent, job, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("setting memory import job owner: %w", err)
	}
	err = r.Create(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("creating memory import job %s: %w", name, err)
	}
	return job, nil
}

func (r *Reconciler) pruneMemoryImportArtifacts(ctx context.Context, agent *agentsv1.Agent, keep map[string]bool) error {
	jobs := &batchv1.JobList{}
	err := r.List(ctx, jobs, client.InNamespace(r.Namespace), client.MatchingLabels{
		LabelAgent:        agent.Name,
		memoryImportLabel: "true",
	})
	if err != nil {
		return fmt.Errorf("listing memory import jobs: %w", err)
	}
	for i := range jobs.Items {
		if keep[jobs.Items[i].Name] {
			continue
		}
		err = ignoreNotFound(r.Delete(ctx, &jobs.Items[i]))
		if err != nil {
			return fmt.Errorf("deleting memory import job %s: %w", jobs.Items[i].Name, err)
		}
	}

	configMaps := &corev1.ConfigMapList{}
	err = r.List(ctx, configMaps, client.InNamespace(r.Namespace), client.MatchingLabels{
		LabelAgent:        agent.Name,
		memoryImportLabel: "true",
	})
	if err != nil {
		return fmt.Errorf("listing memory import config maps: %w", err)
	}
	for i := range configMaps.Items {
		if keep[configMaps.Items[i].Name] {
			continue
		}
		err = ignoreNotFound(r.Delete(ctx, &configMaps.Items[i]))
		if err != nil {
			return fmt.Errorf("deleting memory import config map %s: %w", configMaps.Items[i].Name, err)
		}
	}
	return nil
}

func (r *Reconciler) deleteMemoryImportArtifacts(ctx context.Context, jobName, configMapName string) error {
	objects := []client.Object{
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: r.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: r.Namespace}},
	}
	for _, object := range objects {
		err := ignoreNotFound(r.Delete(ctx, object))
		if err != nil {
			return fmt.Errorf("deleting memory import artifact %s: %w", object.GetName(), err)
		}
	}
	return nil
}
