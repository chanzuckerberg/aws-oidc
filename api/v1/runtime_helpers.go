package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	DataVolumeName           = "data"
	statefulSetNameMaxLength = 61
	objectNameMaxLength      = 63
)

func (a *Agent) ServiceName() string {
	return truncateName("agent-"+sanitizeName(a.Name), objectNameMaxLength)
}

func (a *Agent) AWSConfigMapName() string {
	return truncateName("agent-"+sanitizeName(a.Name)+"-aws-config", objectNameMaxLength)
}

func (a *Agent) StatefulSetName() string {
	return truncateName("agent-"+sanitizeName(a.Name), statefulSetNameMaxLength)
}

func (a *Agent) PersistentVolumeClaimName() string {
	return truncateName("agent-"+sanitizeName(a.Name)+"-workspace", objectNameMaxLength)
}

func (a *Agent) PersistentDataSubPath() string {
	return "workspaces/main"
}

func (a *Agent) ServiceAccountName() string {
	token := strings.ReplaceAll(string(a.UID), "-", "")
	if token == "" {
		token = hashHex(a.Namespace + "/" + a.Name)
	}
	if len(token) > 12 {
		token = token[:12]
	}
	return truncateName("remote-agent-"+token, objectNameMaxLength)
}

func (a *Agent) ServiceAccountSubject(namespace string) string {
	return fmt.Sprintf("system:serviceaccount:%s:%s", namespace, a.ServiceAccountName())
}

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

func truncateName(name string, max int) string {
	if len(name) <= max {
		return name
	}
	const suffixLength = 9
	return strings.Trim(name[:max-suffixLength], "-") + "-" + hashHex(name)[:8]
}

func hashHex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
