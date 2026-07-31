// Package configmap reads and writes the aws-oidc rolemap, stored as a Kubernetes
// ConfigMap. The cronjob writes it; the config server reads it on each request.
package configmap

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// NewInClusterConfig loads the pod's in-cluster REST config and returns it along with the
// pod's own namespace. Callers that need more than the typed client (for example a dynamic
// client for custom resources) build their clients from this config.
func NewInClusterConfig() (*rest.Config, string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, "", fmt.Errorf("loading in-cluster config: %w", err)
	}
	return config, currentNamespace(), nil
}

// NewInClusterClient builds a Kubernetes client from the pod's in-cluster service account
// and returns it along with the pod's own namespace.
func NewInClusterClient() (kubernetes.Interface, string, error) {
	config, namespace, err := NewInClusterConfig()
	if err != nil {
		return nil, "", err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", fmt.Errorf("creating kubernetes client: %w", err)
	}
	return client, namespace, nil
}

func currentNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	b, err := os.ReadFile(namespaceFile)
	if err == nil {
		if ns := strings.TrimSpace(string(b)); ns != "" {
			return ns
		}
	}
	return "default"
}

// ReadRoleMappings fetches the rolemap ConfigMap and parses the role mappings out of the
// given key. A missing ConfigMap or key is treated as an empty rolemap, not an error, so a
// server in an environment without the cronjob (e.g. rdev) still serves cleanly.
func ReadRoleMappings(ctx context.Context, client kubernetes.Interface, namespace, name, key string) (okta.OIDCRoleMappings, error) {
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return okta.OIDCRoleMappings{}, nil
		}
		return nil, fmt.Errorf("getting configmap %s/%s: %w", namespace, name, err)
	}

	raw, ok := cm.Data[key]
	if !ok {
		return okta.OIDCRoleMappings{}, nil
	}

	mappings := okta.OIDCRoleMappings{}
	err = yaml.Unmarshal([]byte(raw), &mappings)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling rolemap from configmap %s/%s key %q: %w", namespace, name, key, err)
	}
	return mappings, nil
}

// NewMappingsProvider returns an awsaccess.MappingsProvider that reads the rolemap
// ConfigMap fresh on each call, grouped by client ID. Both the config server and the portal
// use it so they read the rolemap the same way.
func NewMappingsProvider(client kubernetes.Interface, namespace, name, key string) awsaccess.MappingsProvider {
	return func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error) {
		mappings, err := ReadRoleMappings(ctx, client, namespace, name, key)
		if err != nil {
			return nil, err
		}
		return mappings.ByClientID(), nil
	}
}

// WriteData upserts the rolemap ConfigMap, setting data[key] to the given bytes. The
// ConfigMap is created if it does not exist.
func WriteData(ctx context.Context, client kubernetes.Interface, namespace, name, key string, data []byte) error {
	configMaps := client.CoreV1().ConfigMaps(namespace)

	existing, err := configMaps.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("getting configmap %s/%s: %w", namespace, name, err)
		}
		_, err = configMaps.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Data:       map[string]string{key: string(data)},
		}, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating configmap %s/%s: %w", namespace, name, err)
		}
		return nil
	}

	if existing.Data == nil {
		existing.Data = map[string]string{}
	}
	existing.Data[key] = string(data)
	_, err = configMaps.Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating configmap %s/%s: %w", namespace, name, err)
	}
	return nil
}
