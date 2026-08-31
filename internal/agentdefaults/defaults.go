// Package agentdefaults loads operator and portal defaults from a YAML file mounted from a
// ConfigMap. Updating the ConfigMap propagates to both processes within ~60 seconds (kubelet
// sync period) without a restart.
package agentdefaults

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultTTL = 30 * time.Second

// Defaults holds every tunable default for the agent operator and portal. Fields are optional:
// a zero value means "no default from the ConfigMap; fall back to the process flag or
// built-in constant."
type Defaults struct {
	// Image is the container image for agent thread pods.
	Image string `yaml:"image"`
	// Command overrides the image entrypoint when set.
	Command []string `yaml:"command"`
	// StorageClass is the ReadWriteMany storage class for agent workspace PVCs.
	StorageClass string `yaml:"storageClass"`
	// WorkspaceSize is the PVC storage request (an EFS placeholder; the filesystem is elastic).
	WorkspaceSize string `yaml:"workspaceSize"`
	// CPU is the default CPU request/limit placed on each thread container.
	CPU string `yaml:"cpu"`
	// Memory is the default memory request/limit placed on each thread container.
	Memory string `yaml:"memory"`
	// MaxCPU is the ceiling the portal enforces on CPU requests.
	MaxCPU string `yaml:"maxCPU"`
	// MaxMemory is the ceiling the portal enforces on memory requests.
	MaxMemory string `yaml:"maxMemory"`
	// MaxWorkspace is the ceiling the portal enforces on workspace size requests.
	MaxWorkspace string `yaml:"maxWorkspace"`
	// MaxThreads is the maximum number of threads any agent may run.
	MaxThreads int `yaml:"maxThreads"`
}

// Loader reads Defaults from a YAML file and caches the result. Re-reads happen in the
// background so callers always get a recent value without blocking on disk I/O.
type Loader struct {
	path string
	ttl  time.Duration

	mu        sync.RWMutex
	cached    *Defaults
	fetchedAt time.Time
}

// NewLoader returns a Loader that reads from path. Pass an empty path to get a Loader that
// always returns an empty Defaults (all zero values — callers fall back to their own
// constants).
func NewLoader(path string) *Loader {
	return &Loader{path: path, ttl: defaultTTL}
}

// Load returns the current defaults. If the cached value is fresh it is returned without I/O.
// A stale or missing cache triggers a synchronous re-read. Errors reading the file are
// returned to the caller; the stale cached value is kept so a transient I/O error does not
// reset all defaults to zero.
func (l *Loader) Load() (*Defaults, error) {
	if l.path == "" {
		return &Defaults{}, nil
	}

	l.mu.RLock()
	if l.cached != nil && time.Since(l.fetchedAt) < l.ttl {
		d := l.cached
		l.mu.RUnlock()
		return d, nil
	}
	l.mu.RUnlock()

	return l.reload()
}

func (l *Loader) reload() (*Defaults, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cached != nil && time.Since(l.fetchedAt) < l.ttl {
		return l.cached, nil
	}

	data, err := os.ReadFile(l.path)
	if err != nil {
		if l.cached != nil {
			return l.cached, fmt.Errorf("re-reading defaults %s (returning stale): %w", l.path, err)
		}
		return &Defaults{}, fmt.Errorf("reading defaults %s: %w", l.path, err)
	}

	var d Defaults
	err = yaml.Unmarshal(data, &d)
	if err != nil {
		if l.cached != nil {
			return l.cached, fmt.Errorf("parsing defaults %s (returning stale): %w", l.path, err)
		}
		return &Defaults{}, fmt.Errorf("parsing defaults %s: %w", l.path, err)
	}

	l.cached = &d
	l.fetchedAt = time.Now()
	return &d, nil
}
