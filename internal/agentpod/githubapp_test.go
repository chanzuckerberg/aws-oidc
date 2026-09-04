package agentpod

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testPrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

func TestEnsureGitHubAppSecret(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	keyPEM := testPrivateKeyPEM(t)

	name, err := EnsureGitHubAppSecret(ctx, c, testNamespace, keyPEM)
	require.NoError(t, err)
	require.Equal(t, GitHubAppSecretName, name)

	secret := &corev1.Secret{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: GitHubAppSecretName}, secret))
	require.Equal(t, corev1.SecretTypeOpaque, secret.Type)
	// Exactly one key, so a workspace pod mounting this Secret cannot reach anything else.
	require.Len(t, secret.Data, 1)
	require.Equal(t, keyPEM, string(secret.Data[githubAppKeyFileName]))

	// Rotation is a restart, so writing the same Secret again has to replace the value.
	rotated := testPrivateKeyPEM(t)
	_, err = EnsureGitHubAppSecret(ctx, c, testNamespace, rotated)
	require.NoError(t, err)
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: GitHubAppSecretName}, secret))
	require.Equal(t, rotated, string(secret.Data[githubAppKeyFileName]))
}

// `argus set secret` carries the key through a shell, which does not always preserve a
// multi-line value, so a base64-wrapped PEM is accepted as well.
func TestEnsureGitHubAppSecretAcceptsBase64(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	keyPEM := testPrivateKeyPEM(t)

	_, err := EnsureGitHubAppSecret(ctx, c, testNamespace, base64.StdEncoding.EncodeToString([]byte(keyPEM)))
	require.NoError(t, err)

	secret := &corev1.Secret{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: GitHubAppSecretName}, secret))
	require.Equal(t, keyPEM, string(secret.Data[githubAppKeyFileName]))
}

// A mangled key must stop the operator at startup. Accepting it would strand the failure
// inside an agent session hours later, where it looks like a GitHub outage.
func TestEnsureGitHubAppSecretRejectsGarbage(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	_, err := EnsureGitHubAppSecret(ctx, c, testNamespace, "not a key")
	require.ErrorContains(t, err, "neither PEM nor base64-encoded PEM")

	_, err = EnsureGitHubAppSecret(ctx, c, testNamespace, base64.StdEncoding.EncodeToString([]byte("not a key")))
	require.ErrorContains(t, err, "not PEM")

	secrets := &corev1.SecretList{}
	require.NoError(t, c.List(ctx, secrets))
	require.Empty(t, secrets.Items, "a rejected key must not leave a Secret behind")
}
