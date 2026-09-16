package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
)

func TestRuntimeObjectCacheScopesJobsToNamespace(t *testing.T) {
	const namespace = "agent-system"

	objectCaches := runtimeObjectCache(namespace)

	for object, objectCache := range objectCaches {
		if _, ok := object.(*batchv1.Job); !ok {
			continue
		}
		require.Contains(t, objectCache.Namespaces, namespace)
		return
	}

	t.Fatal("job cache not configured")
}
