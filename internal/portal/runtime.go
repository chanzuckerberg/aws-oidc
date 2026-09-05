package portal

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

// tailscaleForm carries the state for the Tailscale page.
type tailscaleForm struct {
	Enabled bool
	SSHUser string
}

// tailscaleFormFromAgent builds the Tailscale form state from a stored agent.
func tailscaleFormFromAgent(agent *agentsv1.Agent) tailscaleForm {
	if agent == nil || agent.Spec.Tailscale == nil {
		return tailscaleForm{}
	}
	return tailscaleForm{
		Enabled: true,
		SSHUser: agent.Spec.Tailscale.SSHUser,
	}
}

// deriveTailscaleUser returns the email local part to use as the SSH username. It rejects
// empty values and "root".
func deriveTailscaleUser(email string) (string, error) {
	local, _, _ := strings.Cut(email, "@")
	local, _, _ = strings.Cut(local, "+")
	local = strings.TrimSpace(local)
	if local == "" {
		return "", fmt.Errorf("cannot determine SSH user: owner email is empty or malformed")
	}
	if local == "root" {
		return "", fmt.Errorf("root is not allowed as the SSH user")
	}
	return local, nil
}

// validSSHUserRe matches the CRD's sshUser pattern: starts with a letter or underscore,
// followed by lowercase letters, digits, underscores or hyphens.
var validSSHUserRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// validSSHUser reports whether s is an acceptable run-as username.
func validSSHUser(s string) bool {
	return validSSHUserRe.MatchString(s)
}

// workspaceNameRe matches the CRD's workspace name pattern, so the portal rejects a bad name with a
// message instead of surfacing an API server validation error.
var workspaceNameRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

const (
	// firstWorkspaceName is the workspace an agent gets when its owner turns on the runtime without
	// naming one, so enabling the runtime is a single click.
	firstWorkspaceName = "main"

	defaultCPU           = "500m"
	defaultMemory        = "1Gi"
	defaultWorkspaceSize = "50Gi"

	workspaceNameMaxLength = 24
)

// AgentLimits caps what an owner may ask for when running an agent in the cluster. The portal
// is the only place a person sets these, so this is where the ceiling belongs.
type AgentLimits struct {
	MaxCPU              string
	MaxMemory           string
	MaxWorkspaces       int
	MaxWorkspace        string
	DefaultImage        string
	DefaultStorageClass string
}

// defaults fills the unset caps and is where the numbers live.
func (l AgentLimits) defaults() AgentLimits {
	if l.MaxCPU == "" {
		l.MaxCPU = "4"
	}
	if l.MaxMemory == "" {
		l.MaxMemory = "16Gi"
	}
	if l.MaxWorkspaces <= 0 {
		l.MaxWorkspaces = 5
	}
	if l.MaxWorkspace == "" {
		l.MaxWorkspace = "500Gi"
	}
	return l
}

// defaultRuntime is what a new agent runs with before its owner has opened the Runtime page, so
// creating an agent already gives it a workspace pod rather than an inert record.
func defaultRuntime(limits AgentLimits) (*agentsv1.AgentRuntime, []agentsv1.AgentWorkspace) {
	limits = limits.defaults()
	sizing := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(defaultCPU),
		corev1.ResourceMemory: resource.MustParse(defaultMemory),
	}
	workspaceSize := resource.MustParse(defaultWorkspaceSize)
	runtime := &agentsv1.AgentRuntime{
		Image:         limits.DefaultImage,
		StorageClass:  limits.DefaultStorageClass,
		WorkspaceSize: &workspaceSize,
		Resources:     corev1.ResourceRequirements{Requests: sizing, Limits: sizing.DeepCopy()},
	}
	return runtime, []agentsv1.AgentWorkspace{{Name: firstWorkspaceName}}
}

// runtimeForm is the runtime section of the agent form, both what is rendered and what a
// submission is read back into so a rejected form comes back with the values the person typed.
type runtimeForm struct {
	Enabled       bool
	CPU           string
	Memory        string
	Image         string
	StorageClass  string
	WorkspaceSize string
	Workspaces    []workspaceForm
	// NewWorkspace is the name typed into the add-a-workspace box, kept so a rejected form does not
	// lose it.
	NewWorkspace string
	Limits       AgentLimits
}

// workspaceForm is one row of the workspaces table, carrying the operator's report of the workspace so
// the owner sees whether it is actually running.
type workspaceForm struct {
	Name      string
	Suspended bool
	State     string
	Message   string
}

// runtimeFromAgent builds the form's runtime section from a stored agent. A nil agent, or one
// that does not run in the cluster, yields the defaults with the runtime off.
func runtimeFromAgent(agent *agentsv1.Agent, limits AgentLimits) runtimeForm {
	form := runtimeForm{
		CPU:           defaultCPU,
		Memory:        defaultMemory,
		WorkspaceSize: defaultWorkspaceSize,
		StorageClass:  limits.DefaultStorageClass,
		Image:         limits.DefaultImage,
		Limits:        limits,
	}
	if agent == nil || agent.Spec.Runtime == nil {
		return form
	}
	agentRuntime := agent.Spec.Runtime

	form.Enabled = true
	if cpu := agentRuntime.Resources.Requests.Cpu(); !cpu.IsZero() {
		form.CPU = cpu.String()
	}
	if memory := agentRuntime.Resources.Requests.Memory(); !memory.IsZero() {
		form.Memory = memory.String()
	}
	if agentRuntime.Image != "" {
		form.Image = agentRuntime.Image
	}
	if agentRuntime.StorageClass != "" {
		form.StorageClass = agentRuntime.StorageClass
	}
	if agentRuntime.WorkspaceSize != nil {
		form.WorkspaceSize = agentRuntime.WorkspaceSize.String()
	}

	states := make(map[string]agentsv1.WorkspaceStatus, len(agent.Status.Workspaces))
	for _, status := range agent.Status.Workspaces {
		states[status.Name] = status
	}
	for _, workspace := range agent.Spec.Workspaces {
		status := states[workspace.Name]
		form.Workspaces = append(form.Workspaces, workspaceForm{
			Name:      workspace.Name,
			Suspended: workspace.Suspended,
			State:     string(status.State),
			Message:   status.Message,
		})
	}
	return form
}

// runtimeFromForm reads the runtime section back out of a submission, for re-rendering a form
// that failed validation.
func runtimeFromForm(r *http.Request, limits AgentLimits) runtimeForm {
	form := runtimeForm{
		Enabled:       r.FormValue("runtime") != "",
		CPU:           r.FormValue("cpu"),
		Memory:        r.FormValue("memory"),
		Image:         r.FormValue("image"),
		StorageClass:  r.FormValue("storage-class"),
		WorkspaceSize: r.FormValue("workspace-size"),
		NewWorkspace:  r.FormValue("new-workspace"),
		Limits:        limits,
	}

	suspended := selection(r.Form["suspended"])
	removed := selection(r.Form["remove-workspace"])
	for _, name := range r.Form["workspace"] {
		if removed[name] {
			continue
		}
		form.Workspaces = append(form.Workspaces, workspaceForm{Name: name, Suspended: suspended[name]})
	}
	return form
}

// parseRuntime turns a submission into the agent's desired runtime and workspaces. It returns nil
// when the owner has not turned the runtime on, which is what tells the operator the agent has
// no pods.
//
// Sizing is applied as both requests and limits, so a workspace gets the CPU and memory it asked
// for and cannot burst past it. That makes an agent's cost predictable, which matters when the
// people creating them are not the people paying for them.
func parseRuntime(r *http.Request, current *agentsv1.Agent, limits AgentLimits, isAdmin bool) (*agentsv1.AgentRuntime, []agentsv1.AgentWorkspace, error) {
	if r.FormValue("runtime") == "" {
		return nil, nil, nil
	}
	limits = limits.defaults()

	cpu, err := parseQuantity(r, "cpu", "CPU", defaultCPU, limits.MaxCPU)
	if err != nil {
		return nil, nil, err
	}
	memory, err := parseQuantity(r, "memory", "Memory", defaultMemory, limits.MaxMemory)
	if err != nil {
		return nil, nil, err
	}
	workspaceSize, err := parseQuantity(r, "workspace-size", "Workspace size", defaultWorkspaceSize, limits.MaxWorkspace)
	if err != nil {
		return nil, nil, err
	}

	workspaces, err := parseWorkspaces(r, limits.MaxWorkspaces)
	if err != nil {
		return nil, nil, err
	}

	var image, storageClass string
	if isAdmin {
		image = strings.TrimSpace(r.FormValue("image"))
		storageClass = strings.TrimSpace(r.FormValue("storage-class"))
	} else if current != nil && current.Spec.Runtime != nil {
		image = current.Spec.Runtime.Image
		storageClass = current.Spec.Runtime.StorageClass
	}

	sizing := corev1.ResourceList{corev1.ResourceCPU: cpu, corev1.ResourceMemory: memory}
	agentRuntime := &agentsv1.AgentRuntime{
		Image:         image,
		StorageClass:  storageClass,
		WorkspaceSize: &workspaceSize,
		Resources:     corev1.ResourceRequirements{Requests: sizing, Limits: sizing.DeepCopy()},
	}
	return agentRuntime, workspaces, nil
}

// parseWorkspaces reads the workspaces table: the rows that came back, minus the ones marked for
// removal, plus a newly named one.
func parseWorkspaces(r *http.Request, maxWorkspaces int) ([]agentsv1.AgentWorkspace, error) {
	suspended := selection(r.Form["suspended"])
	removed := selection(r.Form["remove-workspace"])

	workspaces := make([]agentsv1.AgentWorkspace, 0, len(r.Form["workspace"])+1)
	seen := map[string]bool{}

	add := func(name string) error {
		if seen[name] {
			return fmt.Errorf("there is already a workspace named %q", name)
		}
		seen[name] = true
		workspaces = append(workspaces, agentsv1.AgentWorkspace{Name: name, Suspended: suspended[name]})
		return nil
	}

	for _, name := range r.Form["workspace"] {
		if removed[name] {
			continue
		}
		err := add(name)
		if err != nil {
			return nil, err
		}
	}

	newWorkspace := strings.ToLower(strings.TrimSpace(r.FormValue("new-workspace")))
	if newWorkspace != "" {
		if len(newWorkspace) > workspaceNameMaxLength {
			return nil, fmt.Errorf("workspace names are limited to %d characters", workspaceNameMaxLength)
		}
		if !workspaceNameRe.MatchString(newWorkspace) {
			return nil, fmt.Errorf("workspace name %q must use only lowercase letters, numbers, and dashes", newWorkspace)
		}
		err := add(newWorkspace)
		if err != nil {
			return nil, err
		}
	}

	// Turning the runtime on with no workspaces means one workspace, rather than an agent that runs
	// nothing.
	if len(workspaces) == 0 {
		workspaces = append(workspaces, agentsv1.AgentWorkspace{Name: firstWorkspaceName})
	}

	if len(workspaces) > maxWorkspaces {
		return nil, fmt.Errorf("an agent is limited to %d workspaces", maxWorkspaces)
	}
	return workspaces, nil
}

// parseQuantity reads one sizing field, falling back to the default when it is blank and
// refusing anything above the cap.
func parseQuantity(r *http.Request, field, label, fallback, limit string) (resource.Quantity, error) {
	raw := strings.TrimSpace(r.FormValue(field))
	if raw == "" {
		raw = fallback
	}

	value, err := resource.ParseQuantity(raw)
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("%s %q is not a valid quantity (for example %s)", label, raw, fallback)
	}
	if value.Sign() <= 0 {
		return resource.Quantity{}, fmt.Errorf("%s must be greater than zero", label)
	}

	ceiling, err := resource.ParseQuantity(limit)
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("parsing %s limit %q: %w", label, limit, err)
	}
	if value.Cmp(ceiling) > 0 {
		return resource.Quantity{}, fmt.Errorf("%s is limited to %s", label, ceiling.String())
	}
	return value, nil
}

func selection(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
