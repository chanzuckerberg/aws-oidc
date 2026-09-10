package portal

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const memoryImportTestNamespace = "agents-test"

func TestClaudePageOffersLocalImportWithUndo(t *testing.T) {
	store := newMemStore()
	store.agents["bot"] = memoryImportAgent()
	server := memoryImportServer(t, store)

	request := httptest.NewRequest(http.MethodGet, "/agents/bot/claude", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "Import from this computer")
	require.Contains(t, response.Body.String(), "Undo import")
	require.Contains(t, response.Body.String(), "Do not put credentials or secrets")
}

func TestClaudeConfigRejectsOversizedImportedContent(t *testing.T) {
	store := newMemStore()
	store.agents["bot"] = memoryImportAgent()
	server := memoryImportServer(t, store)

	form := "claude-md=" + strings.Repeat("x", maxClaudeConfigBytes+1) + "&settings-json=%7B%7D"
	request := httptest.NewRequest(http.MethodPost, "/agents/bot/claude", strings.NewReader(form))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "no larger than 256 KiB")
	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Nil(t, agent.Spec.Claude)
}

func TestImportRepositoryMemoryStagesFilesAndReplacesRevision(t *testing.T) {
	store := newMemStore()
	store.agents["bot"] = memoryImportAgent()
	kubeClient := kubernetesfake.NewSimpleClientset()
	server := memoryImportServerWithClient(t, store, kubeClient)

	response := importMemory(t, server, "chanzuckerberg/aws-oidc", map[string]string{
		"MEMORY.md":  "# Index\n",
		"testing.md": "# Tests\n",
	})
	require.Equal(t, http.StatusSeeOther, response.Code)

	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Len(t, agent.Spec.Claude.MemoryImports, 1)
	first := agent.Spec.Claude.MemoryImports[0]
	require.Equal(t, agentsv1.Repository("chanzuckerberg/aws-oidc"), first.Repository)
	require.Len(t, first.Revision, 64)

	configMap, err := kubeClient.CoreV1().ConfigMaps(memoryImportTestNamespace).Get(
		context.Background(),
		agent.MemoryImportConfigMapName(first.Repository, first.Revision),
		metav1.GetOptions{},
	)
	require.NoError(t, err)
	require.Equal(t, "# Index\n", configMap.Data["MEMORY.md"])
	require.Equal(t, "# Tests\n", configMap.Data["testing.md"])
	require.Equal(t, agent.Name, configMap.Labels["agents.czi.team/agent"])

	response = importMemory(t, server, "chanzuckerberg/aws-oidc", map[string]string{
		"MEMORY.md": "# Index\n",
	})
	require.Equal(t, http.StatusSeeOther, response.Code)
	agent, err = store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Len(t, agent.Spec.Claude.MemoryImports, 1)
	require.NotEqual(t, first.Revision, agent.Spec.Claude.MemoryImports[0].Revision)
}

func TestImportRepositoryMemoryRejectsInvalidUploads(t *testing.T) {
	tooManyFiles := map[string]string{"MEMORY.md": "# Index\n"}
	for i := 0; i < maxMemoryImportFiles; i++ {
		tooManyFiles[fmt.Sprintf("topic-%d.md", i)] = "x"
	}
	oversizedTotal := map[string]string{"MEMORY.md": strings.Repeat("x", 60*1024)}
	for i := 0; i < 8; i++ {
		oversizedTotal[fmt.Sprintf("topic-%d.md", i)] = strings.Repeat("x", 60*1024)
	}
	tests := map[string]struct {
		repository string
		files      map[string]string
	}{
		"unconfigured repository": {
			repository: "chanzuckerberg/not-configured",
			files:      map[string]string{"MEMORY.md": "# Index\n"},
		},
		"missing index": {
			repository: "chanzuckerberg/aws-oidc",
			files:      map[string]string{"topic.md": "# Topic\n"},
		},
		"nested file": {
			repository: "chanzuckerberg/aws-oidc",
			files:      map[string]string{"nested/MEMORY.md": "# Index\n"},
		},
		"non markdown": {
			repository: "chanzuckerberg/aws-oidc",
			files:      map[string]string{"MEMORY.md": "# Index\n", "secret.txt": "no"},
		},
		"oversized file": {
			repository: "chanzuckerberg/aws-oidc",
			files:      map[string]string{"MEMORY.md": strings.Repeat("x", maxMemoryImportFileBytes+1)},
		},
		"too many files": {
			repository: "chanzuckerberg/aws-oidc",
			files:      tooManyFiles,
		},
		"oversized total": {
			repository: "chanzuckerberg/aws-oidc",
			files:      oversizedTotal,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			store := newMemStore()
			store.agents["bot"] = memoryImportAgent()
			server := memoryImportServer(t, store)

			response := importMemory(t, server, test.repository, test.files)
			require.Equal(t, http.StatusBadRequest, response.Code)
			agent, err := store.Get(context.Background(), "bot")
			require.NoError(t, err)
			require.Empty(t, agent.Spec.Claude)
		})
	}
}

func TestImportRepositoryMemoryRejectsUnreachableRepository(t *testing.T) {
	store := newMemStore()
	store.agents["bot"] = memoryImportAgent()
	kubeClient := kubernetesfake.NewSimpleClientset()
	server, err := NewServer(Config{
		Store:            store,
		Identity:         &IdentityResolver{devSub: "s", devEmail: "a@example.com"},
		Repositories:     unreachableSuggester{},
		MemoryConfigMaps: kubeClient.CoreV1().ConfigMaps(memoryImportTestNamespace),
		Namespace:        memoryImportTestNamespace,
		AgentRuntime:     true,
	})
	require.NoError(t, err)

	response := importMemory(t, server, "chanzuckerberg/aws-oidc", map[string]string{
		"MEMORY.md": "# Index\n",
	})
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestImportRepositoryMemoryEnforcesCountAndTotalLimits(t *testing.T) {
	tests := map[string]map[string]string{
		"file count": {"MEMORY.md": "# Index\n"},
		"total size": {"MEMORY.md": strings.Repeat("x", 60*1024)},
	}
	for index := 0; index < maxMemoryImportFiles; index++ {
		tests["file count"][fmt.Sprintf("topic-%02d.md", index)] = "x"
	}
	for index := 0; index < 8; index++ {
		tests["total size"][fmt.Sprintf("topic-%02d.md", index)] = strings.Repeat("x", 60*1024)
	}

	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			store := newMemStore()
			store.agents["bot"] = memoryImportAgent()
			server := memoryImportServer(t, store)

			response := importMemory(t, server, "chanzuckerberg/aws-oidc", files)
			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestImportRepositoryMemoryRejectsAnotherOwnersAgent(t *testing.T) {
	store := newMemStore()
	agent := memoryImportAgent()
	agent.Spec.Owner = "someone-else"
	store.agents["bot"] = agent
	kubeClient := kubernetesfake.NewSimpleClientset()
	server, err := NewServer(Config{
		Store: store,
		Identity: &IdentityResolver{verifyIDToken: func(context.Context, string) (string, string, []string, error) {
			return "s", "a@example.com", nil, nil
		}},
		Repositories:     stubSuggester{},
		MemoryConfigMaps: kubeClient.CoreV1().ConfigMaps(memoryImportTestNamespace),
		Namespace:        memoryImportTestNamespace,
		AgentRuntime:     true,
	})
	require.NoError(t, err)

	response := importMemoryWithIDToken(t, server, "chanzuckerberg/aws-oidc", map[string]string{
		"MEMORY.md": "# Index\n",
	})
	require.Equal(t, http.StatusForbidden, response.Code)
}

func TestClaudeAndRepositoryUpdatesPreserveOnlyConfiguredMemoryImports(t *testing.T) {
	store := newMemStore()
	agent := memoryImportAgent()
	agent.Spec.Claude = &agentsv1.ClaudeConfig{MemoryImports: []agentsv1.ProjectMemoryImport{{
		Repository: "chanzuckerberg/aws-oidc",
		Revision:   strings.Repeat("a", 64),
	}}}
	store.agents["bot"] = agent
	server := memoryImportServer(t, store)

	form := strings.NewReader("claude-md=%23+Rules&settings-json=%7B%7D")
	request := httptest.NewRequest(http.MethodPost, "/agents/bot/claude", form)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusSeeOther, response.Code)

	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Len(t, agent.Spec.Claude.MemoryImports, 1)

	request = httptest.NewRequest(http.MethodPost, "/agents/bot/repositories", nil)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusSeeOther, response.Code)
	agent, err = store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Empty(t, agent.Spec.Claude.MemoryImports)
}

type unreachableSuggester struct{}

func (unreachableSuggester) Suggest(context.Context, string, int) ([]string, error) {
	return nil, nil
}

func (unreachableSuggester) Reachable(context.Context, string) (bool, error) {
	return false, nil
}

func memoryImportAgent() *agentsv1.Agent {
	return &agentsv1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bot",
			Namespace: memoryImportTestNamespace,
			UID:       types.UID("1234"),
		},
		Spec: agentsv1.AgentSpec{
			Owner:        "s",
			OwnerEmail:   "a@example.com",
			Repositories: []agentsv1.Repository{"chanzuckerberg/aws-oidc"},
			Runtime:      &agentsv1.AgentRuntime{},
		},
	}
}

func memoryImportServer(t *testing.T, store *memStore) *Server {
	t.Helper()
	return memoryImportServerWithClient(t, store, kubernetesfake.NewSimpleClientset())
}

func memoryImportServerWithClient(t *testing.T, store *memStore, kubeClient *kubernetesfake.Clientset) *Server {
	t.Helper()
	server, err := NewServer(Config{
		Store:            store,
		Identity:         &IdentityResolver{devSub: "s", devEmail: "a@example.com"},
		Repositories:     stubSuggester{},
		MemoryConfigMaps: kubeClient.CoreV1().ConfigMaps(memoryImportTestNamespace),
		Namespace:        memoryImportTestNamespace,
		AgentRuntime:     true,
		AgentTailscale:   true,
	})
	require.NoError(t, err)
	return server
}

func importMemory(t *testing.T, server *Server, repository string, files map[string]string) *httptest.ResponseRecorder {
	return importMemoryRequest(t, server, repository, files, false)
}

func importMemoryWithIDToken(t *testing.T, server *Server, repository string, files map[string]string) *httptest.ResponseRecorder {
	return importMemoryRequest(t, server, repository, files, true)
}

func importMemoryRequest(t *testing.T, server *Server, repository string, files map[string]string, withIDToken bool) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("repository", repository))
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		require.NoError(t, writer.WriteField("memory-path", "memory/"+name))
		part, err := writer.CreateFormFile("memory-files", name)
		require.NoError(t, err)
		_, err = part.Write([]byte(files[name]))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/agents/bot/repositories/memory", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if withIDToken {
		request.Header.Set("X-ID-Token", "test-token")
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	return response
}
