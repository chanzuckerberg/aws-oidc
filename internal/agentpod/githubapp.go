package agentpod

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GitHubAppSecretName is the Secret workspace pods mount the GitHub App's private key from.
//
// Argus delivers the key to the operator as an environment variable, which every service in
// the stack shares. Handing a workspace pod that whole environment would give an agent the
// stack's other secrets, so the operator republishes this one key on its own and projects
// nothing else.
const GitHubAppSecretName = "agent-github-app"

// EnsureGitHubAppSecret writes the GitHub App's private key where workspace pods can mount it,
// and returns the Secret's name.
//
// It runs once at operator startup rather than on every reconcile. The operator's own copy of
// the key comes from its environment, which Kubernetes fixes when the pod starts, so a
// rotation only reaches the operator on a restart and re-writing between restarts would
// achieve nothing.
func EnsureGitHubAppSecret(ctx context.Context, c client.Client, namespace, privateKey string) (string, error) {
	pemBytes, err := normalizePrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      GitHubAppSecretName,
		Namespace: namespace,
	}}

	_, err = controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		secret.Labels = map[string]string{labelManagedBy: managedByValue}
		secret.Type = corev1.SecretTypeOpaque
		secret.Data = map[string][]byte{githubAppKeyFileName: pemBytes}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ensuring secret %s: %w", GitHubAppSecretName, err)
	}
	return GitHubAppSecretName, nil
}

// normalizePrivateKey accepts the key either as a PEM or as base64-encoded PEM, because a
// multi-line value does not always survive the shell that set it. Rejecting anything that is
// not a PEM here turns a mangled key into a startup failure rather than a git push that fails
// hours later inside an agent session.
func normalizePrivateKey(privateKey string) ([]byte, error) {
	trimmed := strings.TrimSpace(privateKey)

	block, _ := pem.Decode([]byte(trimmed))
	if block != nil {
		return []byte(trimmed + "\n"), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("github app private key is neither PEM nor base64-encoded PEM")
	}
	block, _ = pem.Decode(decoded)
	if block == nil {
		return nil, fmt.Errorf("github app private key decodes from base64 but is not PEM")
	}
	return append([]byte(strings.TrimSpace(string(decoded))), '\n'), nil
}
