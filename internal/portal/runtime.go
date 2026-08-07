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

// threadNameRe matches the CRD's thread name pattern, so the portal rejects a bad name with a
// message instead of surfacing an API server validation error.
var threadNameRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

const (
	// firstThreadName is the thread an agent gets when its owner turns on the runtime without
	// naming one, so enabling the runtime is a single click.
	firstThreadName = "main"

	defaultCPU     = "500m"
	defaultMemory  = "1Gi"
	defaultStorage = "20Gi"

	threadNameMaxLength = 24
)

// AgentLimits caps what an owner may ask for when running an agent in the cluster. The portal
// is the only place a person sets these, so this is where the ceiling belongs.
type AgentLimits struct {
	MaxCPU     string
	MaxMemory  string
	MaxStorage string
	MaxThreads int
}

// defaults fills the unset caps and is where the numbers live.
func (l AgentLimits) defaults() AgentLimits {
	if l.MaxCPU == "" {
		l.MaxCPU = "4"
	}
	if l.MaxMemory == "" {
		l.MaxMemory = "16Gi"
	}
	if l.MaxStorage == "" {
		l.MaxStorage = "200Gi"
	}
	if l.MaxThreads <= 0 {
		l.MaxThreads = 5
	}
	return l
}

// runtimeForm is the runtime section of the agent form, both what is rendered and what a
// submission is read back into so a rejected form comes back with the values the person typed.
type runtimeForm struct {
	Enabled bool
	CPU     string
	Memory  string
	Storage string
	Threads []threadForm
	// NewThread is the name typed into the add-a-thread box, kept so a rejected form does not
	// lose it.
	NewThread string
	Limits    AgentLimits
}

// threadForm is one row of the threads table, carrying the operator's report of the thread so
// the owner sees whether it is actually running.
type threadForm struct {
	Name      string
	Suspended bool
	State     string
	Message   string
}

// runtimeFromAgent builds the form's runtime section from a stored agent. A nil agent, or one
// that does not run in the cluster, yields the defaults with the runtime off.
func runtimeFromAgent(agent *agentsv1.Agent, limits AgentLimits) runtimeForm {
	form := runtimeForm{
		CPU:     defaultCPU,
		Memory:  defaultMemory,
		Storage: defaultStorage,
		Limits:  limits,
	}
	if agent == nil || agent.Spec.Runtime == nil {
		return form
	}
	runtime := agent.Spec.Runtime

	form.Enabled = true
	if cpu := runtime.Resources.Requests.Cpu(); !cpu.IsZero() {
		form.CPU = cpu.String()
	}
	if memory := runtime.Resources.Requests.Memory(); !memory.IsZero() {
		form.Memory = memory.String()
	}
	if runtime.Storage != nil && runtime.Storage.Size != "" {
		form.Storage = runtime.Storage.Size
	}

	states := make(map[string]agentsv1.ThreadStatus, len(agent.Status.Threads))
	for _, status := range agent.Status.Threads {
		states[status.Name] = status
	}
	for _, thread := range agent.Spec.Threads {
		status := states[thread.Name]
		form.Threads = append(form.Threads, threadForm{
			Name:      thread.Name,
			Suspended: thread.Suspended,
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
		Enabled:   r.FormValue("runtime") != "",
		CPU:       r.FormValue("cpu"),
		Memory:    r.FormValue("memory"),
		Storage:   r.FormValue("storage"),
		NewThread: r.FormValue("new-thread"),
		Limits:    limits,
	}

	suspended := selection(r.Form["suspended"])
	removed := selection(r.Form["remove-thread"])
	for _, name := range r.Form["thread"] {
		if removed[name] {
			continue
		}
		form.Threads = append(form.Threads, threadForm{Name: name, Suspended: suspended[name]})
	}
	return form
}

// parseRuntime turns a submission into the agent's desired runtime and threads. It returns nil
// when the owner has not turned the runtime on, which is what tells the operator the agent has
// no pods.
//
// Sizing is applied as both requests and limits, so a thread gets the CPU and memory it asked
// for and cannot burst past it. That makes an agent's cost predictable, which matters when the
// people creating them are not the people paying for them.
func parseRuntime(r *http.Request, current *agentsv1.Agent, limits AgentLimits) (*agentsv1.AgentRuntime, []agentsv1.AgentThread, error) {
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
	storage, err := parseQuantity(r, "storage", "Storage", defaultStorage, limits.MaxStorage)
	if err != nil {
		return nil, nil, err
	}

	// A provisioned volume cannot shrink, so a smaller request would be silently ignored.
	// Refusing it says so instead of leaving the form disagreeing with reality.
	err = rejectShrink(current, storage)
	if err != nil {
		return nil, nil, err
	}

	threads, err := parseThreads(r, limits.MaxThreads)
	if err != nil {
		return nil, nil, err
	}

	sizing := corev1.ResourceList{corev1.ResourceCPU: cpu, corev1.ResourceMemory: memory}
	runtime := &agentsv1.AgentRuntime{
		Resources: corev1.ResourceRequirements{Requests: sizing, Limits: sizing.DeepCopy()},
		Storage:   &agentsv1.AgentStorage{Size: storage.String()},
	}
	return runtime, threads, nil
}

// parseThreads reads the threads table: the rows that came back, minus the ones marked for
// removal, plus a newly named one.
func parseThreads(r *http.Request, maxThreads int) ([]agentsv1.AgentThread, error) {
	suspended := selection(r.Form["suspended"])
	removed := selection(r.Form["remove-thread"])

	threads := make([]agentsv1.AgentThread, 0, len(r.Form["thread"])+1)
	seen := map[string]bool{}

	add := func(name string) error {
		if seen[name] {
			return fmt.Errorf("there is already a thread named %q", name)
		}
		seen[name] = true
		threads = append(threads, agentsv1.AgentThread{Name: name, Suspended: suspended[name]})
		return nil
	}

	for _, name := range r.Form["thread"] {
		if removed[name] {
			continue
		}
		err := add(name)
		if err != nil {
			return nil, err
		}
	}

	newThread := strings.ToLower(strings.TrimSpace(r.FormValue("new-thread")))
	if newThread != "" {
		if len(newThread) > threadNameMaxLength {
			return nil, fmt.Errorf("thread names are limited to %d characters", threadNameMaxLength)
		}
		if !threadNameRe.MatchString(newThread) {
			return nil, fmt.Errorf("thread name %q must use only lowercase letters, numbers, and dashes", newThread)
		}
		err := add(newThread)
		if err != nil {
			return nil, err
		}
	}

	// Turning the runtime on with no threads means one thread, rather than an agent that runs
	// nothing.
	if len(threads) == 0 {
		threads = append(threads, agentsv1.AgentThread{Name: firstThreadName})
	}

	if len(threads) > maxThreads {
		return nil, fmt.Errorf("an agent is limited to %d threads", maxThreads)
	}
	return threads, nil
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

func rejectShrink(current *agentsv1.Agent, storage resource.Quantity) error {
	if current == nil || current.Spec.Runtime == nil || current.Spec.Runtime.Storage == nil {
		return nil
	}
	existing, err := resource.ParseQuantity(current.Spec.Runtime.Storage.Size)
	if err != nil {
		return nil
	}
	if storage.Cmp(existing) < 0 {
		return fmt.Errorf("storage cannot be reduced below the current %s", existing.String())
	}
	return nil
}

func selection(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
