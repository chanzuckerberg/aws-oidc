package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// WorkspaceVolumeName is the name of the shared workspace volume in every workspace's pod spec.
const WorkspaceVolumeName = "workspace"

const (
	// uidPrefixLength is how many hex characters of the Agent's uid identify it in the names
	// of its workspace service accounts. The uid is server-assigned and fixed-length, so no
	// agent's prefix can be a prefix of another agent's, which is what makes the wildcard in
	// the IAM trust condition safe.
	uidPrefixLength = 12

	// statefulSetNameMaxLength leaves room for the "-0" ordinal suffix so a workspace's pod name
	// stays a valid DNS label, which the headless service's DNS records require.
	statefulSetNameMaxLength = 61

	// objectNameMaxLength is the RFC 1123 label limit Kubernetes applies to object names.
	objectNameMaxLength = 63
)

// ServiceName is the headless service shared by all of the agent's workspaces. It governs the
// workspace StatefulSets, giving each pod a stable DNS name.
func (a *Agent) ServiceName() string {
	return truncateName("agent-"+sanitizeName(a.Name), objectNameMaxLength)
}

// AWSConfigMapName is the ConfigMap holding the agent's rendered AWS config. Every workspace of
// the agent mounts the same one, since they all hold the same access.
func (a *Agent) AWSConfigMapName() string {
	return truncateName("agent-"+sanitizeName(a.Name)+"-aws-config", objectNameMaxLength)
}

// WorkspaceStatefulSetName is the workload running one workspace. It carries the human-readable
// agent and workspace names, since this is what shows up in kubectl and in pod names.
func (a *Agent) WorkspaceStatefulSetName(workspace string) string {
	return truncateName("agent-"+sanitizeName(a.Name)+"-"+sanitizeName(workspace), statefulSetNameMaxLength)
}

// WorkspaceClaimName is the PVC shared by all workspaces of this agent. Each workspace mounts it
// at its own subPath so they get isolated working trees under the same EFS access point.
func (a *Agent) WorkspaceClaimName() string {
	return truncateName("agent-"+sanitizeName(a.Name)+"-workspace", objectNameMaxLength)
}

// WorkspaceSubPath is the directory within the shared workspace PVC where a workspace's
// own working tree lives. Workspaces never share this subtree.
func (a *Agent) WorkspaceSubPath(workspace string) string {
	return "workspaces/" + sanitizeName(workspace)
}

// SharedWorkspaceSubPath is the directory within the shared workspace PVC that every workspace
// can read and write. Workspaces use it to pass files between themselves.
func (a *Agent) SharedWorkspaceSubPath() string {
	return "shared"
}

// WorkspaceServiceAccountName is the identity one workspace's pod runs as. It is keyed on the
// agent's uid rather than its name so that the IAM trust condition can match every workspace of
// this agent with a single wildcard without also matching another agent's workspaces. An agent
// named "foo-bar" would otherwise produce names matching agent "foo"'s prefix.
func (a *Agent) WorkspaceServiceAccountName(workspace string) string {
	return truncateName(a.identityPrefix()+"-"+sanitizeName(workspace), objectNameMaxLength)
}

// WorkspaceSubjectPattern is the Kubernetes service account subject shared by every workspace of
// this agent, as a pattern for a StringLike IAM trust condition. Adding a workspace therefore
// needs no IAM write.
func (a *Agent) WorkspaceSubjectPattern(namespace string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s-*", namespace, a.identityPrefix())
}

// Workspace returns the named workspace from the spec, or nil when the agent has no such workspace.
func (a *Agent) Workspace(name string) *AgentWorkspace {
	for i := range a.Spec.Workspaces {
		if a.Spec.Workspaces[i].Name == name {
			return &a.Spec.Workspaces[i]
		}
	}
	return nil
}

// identityPrefix is the fixed-length, per-agent prefix every workspace service account name
// starts with. It falls back to a hash of the agent's identity when the uid is unset, so
// tests and dry runs still get a stable, unique prefix.
func (a *Agent) identityPrefix() string {
	token := strings.ReplaceAll(string(a.UID), "-", "")
	if token == "" {
		token = hashHex(a.Namespace + "/" + a.Name)
	}
	if len(token) > uidPrefixLength {
		token = token[:uidPrefixLength]
	}
	return "remote-agent-" + token
}

// sanitizeName lowercases and replaces anything outside [a-z0-9-] so the result is usable as
// an RFC 1123 label. Kubernetes object names allow dots, but a dot in a pod name breaks the
// per-pod DNS records the headless service publishes.
func sanitizeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// truncateName shortens a name to max, replacing the tail with a hash of the full name so two
// long names that share a prefix do not collide.
func truncateName(name string, max int) string {
	if len(name) <= max {
		return name
	}
	const suffixLength = 9 // "-" plus 8 hex characters
	return strings.Trim(name[:max-suffixLength], "-") + "-" + hashHex(name)[:8]
}

func hashHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
