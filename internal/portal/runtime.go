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

type tailscaleForm struct {
	Enabled bool
	SSHUser string
}

func tailscaleFormFromAgent(agent *agentsv1.Agent) tailscaleForm {
	if agent == nil || agent.Spec.Tailscale == nil {
		return tailscaleForm{}
	}
	return tailscaleForm{
		Enabled: true,
		SSHUser: agent.Spec.Tailscale.SSHUser,
	}
}

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

var validSSHUserRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

func validSSHUser(s string) bool {
	return validSSHUserRe.MatchString(s)
}

const (
	defaultCPU         = "500m"
	defaultMemory      = "1Gi"
	defaultStorageSize = "50Gi"
)

type AgentLimits struct {
	MaxCPU              string
	MaxMemory           string
	MaxStorage          string
	DefaultImage        string
	DefaultStorageClass string
}

func (l AgentLimits) defaults() AgentLimits {
	if l.MaxCPU == "" {
		l.MaxCPU = "4"
	}
	if l.MaxMemory == "" {
		l.MaxMemory = "16Gi"
	}
	if l.MaxStorage == "" {
		l.MaxStorage = "500Gi"
	}
	return l
}

func defaultRuntime(limits AgentLimits) *agentsv1.AgentRuntime {
	limits = limits.defaults()
	sizing := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(defaultCPU),
		corev1.ResourceMemory: resource.MustParse(defaultMemory),
	}
	storageSize := resource.MustParse(defaultStorageSize)
	return &agentsv1.AgentRuntime{
		Image:        limits.DefaultImage,
		StorageClass: limits.DefaultStorageClass,
		StorageSize:  &storageSize,
		Resources:    corev1.ResourceRequirements{Requests: sizing, Limits: sizing.DeepCopy()},
	}
}

type runtimeForm struct {
	CPU          string
	Memory       string
	Image        string
	StorageClass string
	StorageSize  string
	State        string
	Message      string
	Limits       AgentLimits
}

func runtimeFromAgent(agent *agentsv1.Agent, limits AgentLimits) runtimeForm {
	form := runtimeForm{
		CPU:          defaultCPU,
		Memory:       defaultMemory,
		StorageSize:  defaultStorageSize,
		StorageClass: limits.DefaultStorageClass,
		Image:        limits.DefaultImage,
		Limits:       limits,
	}
	if agent == nil || agent.Spec.Runtime == nil {
		return form
	}

	runtime := agent.Spec.Runtime
	if cpu := runtime.Resources.Requests.Cpu(); !cpu.IsZero() {
		form.CPU = cpu.String()
	}
	if memory := runtime.Resources.Requests.Memory(); !memory.IsZero() {
		form.Memory = memory.String()
	}
	if runtime.Image != "" {
		form.Image = runtime.Image
	}
	if runtime.StorageClass != "" {
		form.StorageClass = runtime.StorageClass
	}
	if runtime.StorageSize != nil {
		form.StorageSize = runtime.StorageSize.String()
	}
	if agent.Status.Runtime != nil {
		form.State = string(agent.Status.Runtime.State)
		form.Message = agent.Status.Runtime.Message
	}
	return form
}

func runtimeFromForm(r *http.Request, limits AgentLimits) runtimeForm {
	return runtimeForm{
		CPU:          r.FormValue("cpu"),
		Memory:       r.FormValue("memory"),
		Image:        r.FormValue("image"),
		StorageClass: r.FormValue("storage-class"),
		StorageSize:  r.FormValue("storage-size"),
		Limits:       limits,
	}
}

func parseRuntime(r *http.Request, current *agentsv1.Agent, limits AgentLimits, isAdmin bool) (*agentsv1.AgentRuntime, error) {
	limits = limits.defaults()

	cpu, err := parseQuantity(r, "cpu", "CPU", defaultCPU, limits.MaxCPU)
	if err != nil {
		return nil, err
	}
	memory, err := parseQuantity(r, "memory", "Memory", defaultMemory, limits.MaxMemory)
	if err != nil {
		return nil, err
	}
	storageSize, err := parseQuantity(r, "storage-size", "Storage size", defaultStorageSize, limits.MaxStorage)
	if err != nil {
		return nil, err
	}

	var image, storageClass string
	suspended := false
	if isAdmin {
		image = strings.TrimSpace(r.FormValue("image"))
		storageClass = strings.TrimSpace(r.FormValue("storage-class"))
	}
	if current != nil && current.Spec.Runtime != nil {
		suspended = current.Spec.Runtime.Suspended
		if !isAdmin {
			image = current.Spec.Runtime.Image
			storageClass = current.Spec.Runtime.StorageClass
		}
	}

	sizing := corev1.ResourceList{corev1.ResourceCPU: cpu, corev1.ResourceMemory: memory}
	return &agentsv1.AgentRuntime{
		Image:        image,
		StorageClass: storageClass,
		StorageSize:  &storageSize,
		Suspended:    suspended,
		Resources:    corev1.ResourceRequirements{Requests: sizing, Limits: sizing.DeepCopy()},
	}, nil
}

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
