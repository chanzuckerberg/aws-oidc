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

// form builds the parsed submission the handlers hand to parseRuntime.
func form(t *testing.T, values url.Values) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/agents", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, r.ParseForm())
	return r
}

func TestParseRuntimeDisabled(t *testing.T) {
	runtime, workspaces, err := parseRuntime(form(t, url.Values{}), nil, AgentLimits{}, true)
	require.NoError(t, err)
	require.Nil(t, runtime)
	require.Nil(t, workspaces)
}

// Turning the runtime on without naming a workspace gives the agent one, so enabling it is a
// single click rather than two steps.
func TestParseRuntimeGivesAFirstWorkspace(t *testing.T) {
	runtime, workspaces, err := parseRuntime(form(t, url.Values{"runtime": {"on"}}), nil, AgentLimits{}, true)
	require.NoError(t, err)
	require.NotNil(t, runtime)
	require.Len(t, workspaces, 1)
	require.Equal(t, firstWorkspaceName, workspaces[0].Name)

	// Sizing lands on both requests and limits, so a workspace cannot burst past what its owner
	// asked for.
	require.Equal(t, defaultCPU, runtime.Resources.Requests.Cpu().String())
	require.Equal(t, defaultCPU, runtime.Resources.Limits.Cpu().String())
}

func TestParseRuntimeReadsSizing(t *testing.T) {
	runtime, _, err := parseRuntime(form(t, url.Values{
		"runtime": {"on"},
		"cpu":     {"2"},
		"memory":  {"4Gi"},
	}), nil, AgentLimits{}, true)
	require.NoError(t, err)

	require.Equal(t, "2", runtime.Resources.Requests.Cpu().String())
	require.Equal(t, "4Gi", runtime.Resources.Requests.Memory().String())
}

func TestParseRuntimeRejectsOversizedRequests(t *testing.T) {
	_, _, err := parseRuntime(form(t, url.Values{"runtime": {"on"}, "cpu": {"64"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "CPU is limited to 4")

	_, _, err = parseRuntime(form(t, url.Values{"runtime": {"on"}, "memory": {"512Gi"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "Memory is limited to 16Gi")

	_, _, err = parseRuntime(form(t, url.Values{"runtime": {"on"}, "cpu": {"a lot"}}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "not a valid quantity")
}

func TestParseWorkspacesAddsSuspendsAndRemoves(t *testing.T) {
	values := url.Values{
		"runtime":          {"on"},
		"workspace":        {"main", "review", "stale"},
		"suspended":        {"review"},
		"remove-workspace": {"stale"},
		"new-workspace":    {"Docs"},
	}

	_, workspaces, err := parseRuntime(form(t, values), nil, AgentLimits{}, true)
	require.NoError(t, err)

	require.Equal(t, []agentsv1.AgentWorkspace{
		{Name: "main"},
		{Name: "review", Suspended: true},
		// A name typed with capitals is lowercased rather than rejected, since the CRD only
		// accepts a DNS label.
		{Name: "docs"},
	}, workspaces)
}

func TestParseWorkspacesRejectsBadAndDuplicateNames(t *testing.T) {
	_, _, err := parseRuntime(form(t, url.Values{
		"runtime":       {"on"},
		"new-workspace": {"my_workspace"},
	}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "lowercase letters, numbers, and dashes")

	_, _, err = parseRuntime(form(t, url.Values{
		"runtime":       {"on"},
		"workspace":     {"main"},
		"new-workspace": {"main"},
	}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "already a workspace named")

	_, _, err = parseRuntime(form(t, url.Values{
		"runtime":       {"on"},
		"new-workspace": {strings.Repeat("a", 25)},
	}), nil, AgentLimits{}, true)
	require.ErrorContains(t, err, "limited to 24 characters")
}

func TestParseWorkspacesEnforcesTheLimit(t *testing.T) {
	_, _, err := parseRuntime(form(t, url.Values{
		"runtime":   {"on"},
		"workspace": {"a", "b", "c"},
	}), nil, AgentLimits{MaxWorkspaces: 2}, true)
	require.ErrorContains(t, err, "limited to 2 workspaces")
}

// Removing the last workspace means the agent runs nothing, which is expressed by suspending or
// disabling the runtime, not by an empty workspace list.
func TestParseWorkspacesRemovingTheLastWorkspaceKeepsOne(t *testing.T) {
	_, workspaces, err := parseRuntime(form(t, url.Values{
		"runtime":          {"on"},
		"workspace":        {"main"},
		"remove-workspace": {"main"},
	}), nil, AgentLimits{}, true)
	require.NoError(t, err)
	require.Len(t, workspaces, 1)
	require.Equal(t, firstWorkspaceName, workspaces[0].Name)
}

func TestRuntimeFromAgentShowsStoredSizingAndState(t *testing.T) {
	agent := &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}}
	agent.Spec.Runtime = &agentsv1.AgentRuntime{}
	agent.Spec.Workspaces = []agentsv1.AgentWorkspace{{Name: "main"}, {Name: "review", Suspended: true}}
	agent.Status.Workspaces = []agentsv1.WorkspaceStatus{
		{Name: "main", State: agentsv1.WorkspaceStateRunning},
		{Name: "review", State: agentsv1.WorkspaceStateSuspended},
	}

	form := runtimeFromAgent(agent, AgentLimits{}.defaults())
	require.True(t, form.Enabled)
	// Sizing the agent never set falls back to the default rather than rendering blank.
	require.Equal(t, defaultCPU, form.CPU)
	require.Equal(t, []workspaceForm{
		{Name: "main", State: "Running"},
		{Name: "review", Suspended: true, State: "Suspended"},
	}, form.Workspaces)
}

// The runtime section is only rendered where the operator can actually run workspaces.
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
	require.Contains(t, shown.Body.String(), "Run this agent as pods")

	hidden := httptest.NewRecorder()
	withoutRuntime.render(hidden, "agent_runtime", data)
	require.NotContains(t, hidden.Body.String(), "Run this agent as pods")
}

// A portal that does not offer the runtime must not clear the runtime of an agent that has one.
func TestParseAgentRuntimeLeavesStoredRuntimeAloneWhenNotOffered(t *testing.T) {
	s, err := NewServer(Config{})
	require.NoError(t, err)

	current := &agentsv1.Agent{Spec: agentsv1.AgentSpec{
		Runtime:    &agentsv1.AgentRuntime{},
		Workspaces: []agentsv1.AgentWorkspace{{Name: "main"}},
	}}

	runtime, workspaces, err := s.parseAgentRuntime(form(t, url.Values{}), current, false)
	require.NoError(t, err)
	require.Equal(t, current.Spec.Runtime, runtime)
	require.Equal(t, current.Spec.Workspaces, workspaces)
}
