package portal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
)

// memStore is an in-memory AgentStore for exercising the handlers.
type memStore struct {
	mu     sync.Mutex
	agents map[string]*agentsv1.Agent
}

func newMemStore() *memStore {
	return &memStore{agents: map[string]*agentsv1.Agent{}}
}

func (m *memStore) Get(_ context.Context, name string) (*agentsv1.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	agent, ok := m.agents[name]
	if !ok {
		return nil, nil
	}
	return agent.DeepCopy(), nil
}

func (m *memStore) Upsert(_ context.Context, agent *agentsv1.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.Name] = agent.DeepCopy()
	return nil
}

func (m *memStore) Delete(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.agents, name)
	return nil
}

func (m *memStore) List(_ context.Context) ([]agentsv1.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	agents := make([]agentsv1.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, *agent.DeepCopy())
	}
	return agents, nil
}

func (m *memStore) ListByOwner(ctx context.Context, owner string) ([]agentsv1.Agent, error) {
	all, err := m.List(ctx)
	if err != nil {
		return nil, err
	}
	owned := make([]agentsv1.Agent, 0, len(all))
	for _, agent := range all {
		if agent.Spec.Owner == owner {
			owned = append(owned, agent)
		}
	}
	return owned, nil
}

// stubSuggester stands in for the GitHub-backed repository lookup so the Repositories step is
// offered without any GitHub credentials.
type stubSuggester struct{}

func (stubSuggester) Suggest(context.Context, string, int) ([]string, error) { return nil, nil }
func (stubSuggester) Reachable(context.Context, string) (bool, error)        { return true, nil }

// fullServer is a portal offering every optional page, which is the shape the walkthrough is
// designed around.
func fullServer(t *testing.T, store *memStore) *Server {
	t.Helper()
	s, err := NewServer(Config{
		Store:          store,
		AgentRuntime:   true,
		AgentTailscale: true,
		Repositories:   stubSuggester{},
		Identity:       &IdentityResolver{devSub: "s", devEmail: "a@example.com"},
	})
	require.NoError(t, err)
	s.entCache = newEntitlementsCache(0, func(context.Context, string) (*Entitlements, error) {
		return &Entitlements{allowed: map[string]agentsv1.AWSGrant{}}, nil
	})
	return s
}

// The walkthrough runs runtime, then Tailscale, then repositories, and ends on AWS access so the
// Okta lookup behind it has the earlier steps to finish in.
func TestOnboardingStepsAreOrderedWithAWSLast(t *testing.T) {
	s := fullServer(t, newMemStore())

	steps := s.onboardingSteps("bot")
	labels := make([]string, len(steps))
	for i, step := range steps {
		labels[i] = step.Nav
	}
	require.Equal(t, []string{"runtime", "tailscale", "repositories", "aws"}, labels)
	require.Equal(t, "/agents/bot/runtime?onboarding=1", s.onboardingStart("bot"))
}

// A portal without the optional pages still walks its owner to AWS access rather than dropping
// them on a step it does not serve.
func TestOnboardingStepsSkipPagesNotOffered(t *testing.T) {
	s, err := NewServer(Config{})
	require.NoError(t, err)

	steps := s.onboardingSteps("bot")
	require.Len(t, steps, 1)
	require.Equal(t, "aws", steps[0].Nav)
	require.Equal(t, "/agents/bot/aws?onboarding=1", s.onboardingStart("bot"))
	require.Equal(t, "/agents/bot", s.onboardingDone("bot"))
}

func TestOnboardingForMarksPositionAndNext(t *testing.T) {
	s := fullServer(t, newMemStore())
	r := httptest.NewRequest(http.MethodGet, "/agents/bot/tailscale?onboarding=1", nil)

	o := s.onboardingFor(r, "bot", "tailscale")
	require.True(t, o.Active)
	require.Equal(t, 2, o.Number())
	require.Equal(t, 4, o.Total())
	require.False(t, o.Last())
	require.Equal(t, "/agents/bot/repositories?onboarding=1", o.NextURL())
	require.Equal(t, "/agents/bot/runtime?onboarding=1", o.PrevURL())

	// The first step has nowhere to go back to, so the template renders no Back button.
	first := s.onboardingFor(httptest.NewRequest(http.MethodGet, "/agents/bot/runtime?onboarding=1", nil), "bot", "runtime")
	require.Equal(t, "", first.Prev)
	require.Equal(t, "", first.PrevURL())

	last := s.onboardingFor(httptest.NewRequest(http.MethodGet, "/agents/bot/aws?onboarding=1", nil), "bot", "aws")
	require.True(t, last.Last())
	require.Equal(t, "", last.Next)
	require.Equal(t, "/agents/bot/repositories?onboarding=1", last.PrevURL())
	require.Equal(t, "/agents/bot/connection", last.Done)
}

// Going back is both a Back button and a click on any step already completed, so a person who
// notices a mistake two steps later can reach it.
func TestCompletedStepsLinkBack(t *testing.T) {
	s := fullServer(t, newMemStore())
	data := pageData{
		Title:               "Repositories",
		User:                &identity.User{Sub: "s"},
		Agent:               &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}},
		Nav:                 "repositories",
		RepositoriesOffered: true,
		Onboarding: s.onboardingFor(
			httptest.NewRequest(http.MethodGet, "/agents/bot/repositories?onboarding=1", nil), "bot", "repositories"),
	}

	rec := httptest.NewRecorder()
	s.render(rec, "agent_repositories", data)
	body := rec.Body.String()

	require.Contains(t, body, `href="/agents/bot/tailscale?onboarding=1">Back</a>`)
	require.Contains(t, body, `href="/agents/bot/runtime?onboarding=1">Runtime</a>`)
	require.Contains(t, body, `href="/agents/bot/tailscale?onboarding=1">Tailscale</a>`)
	// The step being edited and the ones still ahead are not links.
	require.NotContains(t, body, `href="/agents/bot/repositories?onboarding=1">Repositories</a>`)
	require.NotContains(t, body, `href="/agents/bot/aws?onboarding=1">AWS access</a>`)
}

// Editing a page on its own must not pick up the walkthrough chrome.
func TestOnboardingInactiveWithoutTheMarker(t *testing.T) {
	s := fullServer(t, newMemStore())
	r := httptest.NewRequest(http.MethodGet, "/agents/bot/runtime", nil)
	require.False(t, s.onboardingFor(r, "bot", "runtime").Active)
}

// Saving a step during the walkthrough moves to the next one; saving it outside the walkthrough
// stays put, which is what an edit should do.
func TestRedirectAfterSave(t *testing.T) {
	s := fullServer(t, newMemStore())

	cases := []struct {
		name   string
		target string
		nav    string
		want   string
	}{
		{"first step continues", "/agents/bot/runtime?onboarding=1", "runtime", "/agents/bot/tailscale?onboarding=1"},
		{"last step finishes", "/agents/bot/aws?onboarding=1", "aws", "/agents/bot/connection"},
		{"plain edit stays", "/agents/bot/runtime", "runtime", "/agents/bot/runtime"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.redirectAfterSave(rec, httptest.NewRequest(http.MethodPost, tc.target, nil), "bot", tc.nav)
			require.Equal(t, http.StatusSeeOther, rec.Code)
			require.Equal(t, tc.want, rec.Header().Get("Location"))
		})
	}
}

// Creating an agent hands it a runtime, a workspace and tailnet enrollment, so a person who
// accepts every default still ends up with a pod they can SSH into.
func TestCreateTurnsOnTheRuntimeAndTailscale(t *testing.T) {
	store := newMemStore()
	s := fullServer(t, store)

	rec := postCreate(t, s, "bot", "a@example.com")
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/agents/bot/runtime?onboarding=1", rec.Header().Get("Location"))

	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.NotNil(t, agent.Spec.Runtime)
	require.Equal(t, defaultCPU, agent.Spec.Runtime.Resources.Requests.Cpu().String())
	require.Equal(t, defaultMemory, agent.Spec.Runtime.Resources.Requests.Memory().String())
	require.Equal(t, defaultWorkspaceSize, agent.Spec.Runtime.WorkspaceSize.String())
	require.Equal(t, []agentsv1.AgentWorkspace{{Name: firstWorkspaceName}}, agent.Spec.Workspaces)

	require.NotNil(t, agent.Spec.Tailscale)
	require.Equal(t, "a", agent.Spec.Tailscale.SSHUser)
}

// An owner whose email yields no usable SSH user still gets an agent. Enrollment is the thing
// that is left off, not the agent.
func TestCreateSkipsTailscaleWhenNoSSHUserCanBeDerived(t *testing.T) {
	store := newMemStore()
	s := fullServer(t, store)
	s.cfg.Identity = &IdentityResolver{devSub: "s"}

	rec := postCreate(t, s, "bot", "")
	require.Equal(t, http.StatusSeeOther, rec.Code)

	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Nil(t, agent.Spec.Tailscale)
	require.NotNil(t, agent.Spec.Runtime)
}

// A portal offering neither must not promise either, so the agent is created with no runtime
// and no tailnet enrollment.
func TestCreateLeavesTheRuntimeAndTailscaleOffWhenNotOffered(t *testing.T) {
	store := newMemStore()
	s, err := NewServer(Config{
		Store:    store,
		Identity: &IdentityResolver{devSub: "s", devEmail: "a@example.com"},
	})
	require.NoError(t, err)
	s.entCache = newEntitlementsCache(0, func(context.Context, string) (*Entitlements, error) {
		return &Entitlements{}, nil
	})

	rec := postCreate(t, s, "bot", "a@example.com")
	require.Equal(t, "/agents/bot/aws?onboarding=1", rec.Header().Get("Location"))

	agent, err := store.Get(context.Background(), "bot")
	require.NoError(t, err)
	require.Nil(t, agent.Spec.Runtime)
	require.Nil(t, agent.Spec.Workspaces)
	require.Nil(t, agent.Spec.Tailscale)
}

func postCreate(t *testing.T, s *Server, name, email string) *httptest.ResponseRecorder {
	t.Helper()
	s.cfg.Identity = &IdentityResolver{devSub: "s", devEmail: email}
	body := url.Values{"name": {name}}.Encode()
	r := httptest.NewRequest(http.MethodPost, "/agents", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.handleCreate(rec, r)
	return rec
}

// The walkthrough is a chain of redirects through the real handler, so this drives it the way a
// browser does: create an agent, then follow each step to the one after it.
func TestWalkthroughDrivesEveryStepInOrder(t *testing.T) {
	store := newMemStore()
	s := fullServer(t, store)
	s.entCache = newEntitlementsCache(0, func(context.Context, string) (*Entitlements, error) {
		return grantable("111111111111", "core-platform-nonprod", "arn:aws:iam::111111111111:role/poweruser", "poweruser"), nil
	})
	handler := s.Handler()

	post := func(path string, values url.Values) string {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		require.Equal(t, http.StatusSeeOther, rec.Code, path)
		return rec.Header().Get("Location")
	}
	page := func(path string) string {
		t.Helper()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, rec.Code, path)
		return rec.Body.String()
	}

	at := post("/agents", url.Values{"name": {"infra-worker"}, "display-name": {"Infra Worker"}})
	require.Equal(t, "/agents/infra-worker/runtime?onboarding=1", at)
	require.Contains(t, page(at), "step 1 of 4")

	at = post(at, url.Values{"onboarding": {"1"}, "runtime": {"on"}, "cpu": {"2"}, "memory": {"8Gi"}})
	require.Equal(t, "/agents/infra-worker/tailscale?onboarding=1", at)
	require.Contains(t, page(at), "step 2 of 4")

	at = post(at, url.Values{"onboarding": {"1"}, "tailscale": {"on"}})
	require.Equal(t, "/agents/infra-worker/repositories?onboarding=1", at)
	require.Contains(t, page(at), "step 3 of 4")

	at = post(at, url.Values{"onboarding": {"1"}, "repository": {"chanzuckerberg/aws-oidc"}})
	require.Equal(t, "/agents/infra-worker/aws?onboarding=1", at)
	require.Contains(t, page(at), "Finish setup")

	at = post(at, url.Values{
		"onboarding": {"1"},
		"grant":      {"111111111111|arn:aws:iam::111111111111:role/poweruser"},
	})
	require.Equal(t, "/agents/infra-worker/connection", at)

	agent, err := store.Get(context.Background(), "infra-worker")
	require.NoError(t, err)
	require.Equal(t, "2", agent.Spec.Runtime.Resources.Requests.Cpu().String())
	require.Equal(t, "a", agent.Spec.Tailscale.SSHUser)
	require.Equal(t, []agentsv1.Repository{"chanzuckerberg/aws-oidc"}, agent.Spec.Repositories)
	require.Len(t, agent.Spec.Grants, 1)
	require.Equal(t, []agentsv1.AgentWorkspace{{Name: firstWorkspaceName}}, agent.Spec.Workspaces)
}

// Saving a step outside the walkthrough returns to that step rather than moving on, so an edit
// does not drag its owner through the rest of setup.
func TestSavingAStepOutsideTheWalkthroughStaysPut(t *testing.T) {
	store := newMemStore()
	s := fullServer(t, store)
	handler := s.Handler()

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/agents", strings.NewReader("name=bot"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(rec, r)

	rec = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/agents/bot/tailscale", strings.NewReader("tailscale=on"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(rec, r)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/agents/bot/tailscale", rec.Header().Get("Location"))
}

// grantable builds entitlements holding one account and role, indexed the way the AWS page
// validates a submission.
func grantable(accountID, alias, roleARN, roleName string) *Entitlements {
	return &Entitlements{
		Accounts: []awsaccess.Account{{
			ID: accountID, Alias: alias,
			Roles: []awsaccess.Role{{RoleARN: roleARN, RoleName: roleName}},
		}},
		allowed: map[string]agentsv1.AWSGrant{
			accountID + "|" + roleARN: {
				AccountID: accountID, AccountAlias: alias, RoleARN: roleARN, RoleName: roleName,
			},
		},
	}
}

// Warm runs the Okta lookup off the request path, so the AWS step finds its entitlements ready
// rather than waiting on Okta itself.
func TestWarmFillsTheCacheInTheBackground(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	cache := newEntitlementsCache(time.Minute, func(context.Context, string) (*Entitlements, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		close(started)
		<-release
		return &Entitlements{Accounts: []awsaccess.Account{{ID: "111"}}}, nil
	})

	// Warm returning while the fetch is still blocked is the whole point. Were it synchronous
	// this would deadlock, because nothing closes release until after it returns.
	cache.Warm("s")
	<-started
	close(release)

	ent, err := cache.Get(context.Background(), "s")
	require.NoError(t, err)
	require.Len(t, ent.Accounts, 1)

	// A fresh entry means neither a second Warm nor a second Get goes back to Okta.
	cache.Warm("s")
	_, err = cache.Get(context.Background(), "s")
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, calls)
}

// The walkthrough renders its step list and a continue button; a plain edit renders neither.
func TestStepChromeRendersOnlyDuringOnboarding(t *testing.T) {
	s := fullServer(t, newMemStore())
	data := pageData{
		Title:          "Runtime",
		User:           &identity.User{Sub: "s"},
		Agent:          &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}},
		Nav:            "runtime",
		RuntimeOffered: true,
		Runtime:        runtimeFromAgent(nil, AgentLimits{}.defaults()),
	}

	plain := httptest.NewRecorder()
	s.render(plain, "agent_runtime", data)
	require.NotContains(t, plain.Body.String(), "Save and continue")
	require.Contains(t, plain.Body.String(), ">Save<")

	data.Onboarding = s.onboardingFor(
		httptest.NewRequest(http.MethodGet, "/agents/bot/runtime?onboarding=1", nil), "bot", "runtime")
	wizard := httptest.NewRecorder()
	s.render(wizard, "agent_runtime", data)
	body := wizard.Body.String()
	require.Contains(t, body, "step 1 of 4")
	require.Contains(t, body, "Save and continue")
	require.Contains(t, body, `name="onboarding" value="1"`)
	require.Contains(t, body, "/agents/bot/tailscale?onboarding=1")
	require.Contains(t, body, "AWS access")
}
