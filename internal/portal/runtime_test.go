package portal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
)

func form(t *testing.T, values url.Values) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/agents", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, r.ParseForm())
	return r
}

func TestParseRuntimeDisabled(t *testing.T) {
	runtime, err := parseRuntime(form(t, url.Values{}), nil, AgentLimits{}, true)
	require.NoError(t, err)
	require.Nil(t, runtime)
}

func TestParseRuntimeUsesDefaults(t *testing.T) {
	runtime, err := parseRuntime(form(t, url.Values{"runtime": {"on"}}), nil, AgentLimits{}, true)
	require.NoError(t, err)
	require.NotNil(t, runtime)
	require.Equal(t, defaultCPU, runtime.Resources.Requests.Cpu().String())
	require.Equal(t, defaultCPU, runtime.Resources.Limits.Cpu().String())
	require.Equal(t, defaultStorageSize, runtime.StorageSize.String())
}

func TestParseRuntimeReadsSizingAndSuspension(t *testing.T) {
	runtime, err := parseRuntime(form(t, url.Values{
		"runtime":      {"on"},
		"cpu":          {"2"},
		"memory":       {"4Gi"},
		"storage-size": {"100Gi"},
		"suspended":    {"on"},
	}), nil, AgentLimits{}, true)
	require.NoError(t, err)

	require.Equal(t, "2", runtime.Resources.Requests.Cpu().String())
	require.Equal(t, "4Gi", runtime.Resources.Requests.Memory().String())
	require.Equal(t, "100Gi", runtime.StorageSize.String())
	require.True(t, runtime.Suspended)
}

func TestParseRuntimeRejectsOversizedRequests(t *testing.T) {
	_, err := parseRuntime(form(t, url.Values{"runtime": {"on"}, "cpu": {"64"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "CPU is limited to 4")

	_, err = parseRuntime(form(t, url.Values{"runtime": {"on"}, "memory": {"512Gi"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "Memory is limited to 16Gi")

	_, err = parseRuntime(form(t, url.Values{"runtime": {"on"}, "storage-size": {"1Ti"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "Storage size is limited to 500Gi")
}

func TestRuntimeFromAgentShowsStoredState(t *testing.T) {
	agent := &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{Suspended: true}
	agent.Status.Runtime = &agentsv1.RuntimeStatus{State: agentsv1.RuntimeStateSuspended}

	runtime := runtimeFromAgent(agent, AgentLimits{}.defaults())
	require.True(t, runtime.Enabled)
	require.True(t, runtime.Suspended)
	require.Equal(t, "Suspended", runtime.State)
	require.Equal(t, defaultCPU, runtime.CPU)
}

func TestFormHidesRuntimeWhenNotOffered(t *testing.T) {
	withRuntime, err := NewServer(Config{AgentRuntime: true})
	require.NoError(t, err)
	withoutRuntime, err := NewServer(Config{})
	require.NoError(t, err)

	data := pageData{
		Title:        "Edit",
		User:         &identity.User{Sub: "s"},
		Agent:        &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}},
		Entitlements: &Entitlements{},
		Runtime:      runtimeFromAgent(nil, AgentLimits{}.defaults()),
		Action:       "/agents/bot",
	}

	shown := httptest.NewRecorder()
	withRuntime.render(shown, "agent_runtime", data)
	require.Contains(t, shown.Body.String(), "Run this agent in the cluster")

	hidden := httptest.NewRecorder()
	withoutRuntime.render(hidden, "agent_runtime", data)
	require.NotContains(t, hidden.Body.String(), "Run this agent in the cluster")
}

func TestParseAgentRuntimeLeavesStoredRuntimeAloneWhenNotOffered(t *testing.T) {
	s, err := NewServer(Config{})
	require.NoError(t, err)

	current := &agentsv1.Agent{Spec: agentsv1.AgentSpec{Runtime: &agentsv1.AgentRuntime{Suspended: true}}}
	runtime, err := s.parseAgentRuntime(form(t, url.Values{}), current, false)
	require.NoError(t, err)
	require.Equal(t, current.Spec.Runtime, runtime)
}

func TestToggleSuspendScalesTheAgentAsAWhole(t *testing.T) {
	store := newMemStore()
	server := fullServer(t, store)
	postCreate(t, server, "bot", "a@example.com")

	request := httptest.NewRequest(http.MethodPost, "/agents/bot/suspend", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusSeeOther, response.Code)

	agent, err := store.Get(request.Context(), "bot")
	require.NoError(t, err)
	require.True(t, agent.Spec.Runtime.Suspended)
}
