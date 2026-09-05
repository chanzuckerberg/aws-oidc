// Package portal implements the agent-registry control plane UI and API: a person logs in,
// sees the AWS access they already have, registers agents, and grants each agent a subset
// of that access.
//
// Agents are stored as Agent custom resources (agents.czi.team/v1), one per agent, through
// the shared agentstore. The portal writes the desired grants to a CR's spec; the operator
// reconciles them. There is no database and no ConfigMap registry.
package portal

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/agentdefaults"
	"github.com/chanzuckerberg/aws-oidc/internal/agentstore"
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

//go:embed templates/*.html
var templatesFS embed.FS

var agentNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// repoRe matches an "owner/repo" reference. It mirrors the Repository pattern on the CRD so
// the portal rejects the same shapes the API server would.
var repoRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// Config wires the portal's dependencies. BasePath is the URL prefix the portal is served
// under (for example "/portal" when the gateway routes a sub-path to it); empty means root.
type Config struct {
	Apps             okta.AppLister
	MappingsProvider awsaccess.MappingsProvider
	Store            agentstore.AgentStore
	Identity         *IdentityResolver
	BasePath         string
	// AgentRuntime offers the option of running an agent's workspaces as pods. It is off unless
	// the operator can actually run them, so the form does not promise what it cannot deliver.
	AgentRuntime bool
	// AgentTailscale offers the Tailscale page. Off unless the operator is configured for
	// tailnet enrollment, so the nav item is not shown in environments without tailscale.
	AgentTailscale bool
	// Repositories powers the Repositories page's type-ahead and validates saved entries
	// against the repositories the fleet's GitHub App can reach. Nil where the portal has no
	// GitHub credentials, which hides the page.
	Repositories repoSuggester
	// Limits caps the sizing an owner may ask for. Unset fields fall back to defaults.
	Limits AgentLimits
	// Namespace is the Kubernetes namespace the operator and agent pods run in. It is shown
	// in the connect widget so users can copy-paste kubectl exec commands.
	Namespace string
	// DefaultsLoader reads live defaults from the agent-defaults ConfigMap. When set its
	// values pre-fill AgentLimits fields that are unset, so the form reflects the ConfigMap
	// without a restart.
	DefaultsLoader *agentdefaults.Loader
}

// Server is the agent-registry portal.
type Server struct {
	cfg      Config
	tmpl     *template.Template
	basePath string
	entCache *EntitlementsCache
}

// NewServer parses templates and returns a portal server.
func NewServer(cfg Config) (*Server, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"localPart": func(email string) string {
			local, _, _ := strings.Cut(email, "@")
			return local
		},
		"add": func(a, b int) int { return a + b },
	}).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	cfg.Limits = cfg.Limits.defaults()
	s := &Server{cfg: cfg, tmpl: tmpl, basePath: strings.TrimRight(cfg.BasePath, "/")}
	s.entCache = newEntitlementsCache(0, s.fetchEntitlements)
	return s, nil
}

// Handler returns the HTTP handler for the portal.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthy)
	mux.HandleFunc("GET /{$}", s.handleList)
	mux.HandleFunc("GET /agents/new", s.handleNew)
	mux.HandleFunc("POST /agents", s.handleCreate)
	mux.HandleFunc("GET /agents/{name}", s.handleGeneral)
	mux.HandleFunc("POST /agents/{name}", s.handleUpdateGeneral)
	mux.HandleFunc("GET /agents/{name}/aws", s.handleAWS)
	mux.HandleFunc("POST /agents/{name}/aws", s.handleUpdateAWS)
	mux.HandleFunc("GET /agents/{name}/tailscale", s.handleTailscale)
	mux.HandleFunc("POST /agents/{name}/tailscale", s.handleUpdateTailscale)
	mux.HandleFunc("GET /agents/{name}/runtime", s.handleRuntime)
	mux.HandleFunc("POST /agents/{name}/runtime", s.handleUpdateRuntime)
	mux.HandleFunc("GET /agents/{name}/repositories", s.handleRepositories)
	mux.HandleFunc("POST /agents/{name}/repositories", s.handleUpdateRepositories)
	mux.HandleFunc("GET /agents/{name}/repositories/search", s.handleRepositorySearch)
	mux.HandleFunc("POST /agents/{name}/delete", s.handleDelete)
	mux.HandleFunc("GET /agents/{name}/workspaces", s.handleWorkspacesView)
	mux.HandleFunc("POST /agents/{name}/workspaces", s.handleSpawnWorkspace)
	mux.HandleFunc("POST /agents/{name}/workspaces/{workspace}/suspend", s.handleToggleSuspend)
	mux.HandleFunc("POST /agents/{name}/workspaces/{workspace}/delete", s.handleDeleteWorkspace)
	mux.HandleFunc("GET /agents/{name}/connection", s.handleConnection)

	handler := http.Handler(mux)
	if s.basePath != "" {
		root := http.NewServeMux()
		root.HandleFunc("GET /health", healthy)
		root.Handle(s.basePath+"/", http.StripPrefix(s.basePath, mux))
		handler = root
	}

	recovery := handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(recoveryLogger{slog.Default()}),
	)
	return logRequests(recovery(handler))
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || strings.HasSuffix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		slog.Info("portal request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(started),
			"remote", r.RemoteAddr,
			"forwarded_for", r.Header.Get("X-Forwarded-For"),
			"has_authorization", r.Header.Get("Authorization") != "",
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func healthy(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, s.basePath+path, http.StatusSeeOther)
}

type pageData struct {
	Title        string
	BasePath     string
	Namespace    string
	User         *identity.User
	Agents       []agentsv1.Agent
	Agent        *agentsv1.Agent
	Entitlements *Entitlements
	Checked      map[string]bool
	Action       string
	Error        string
	// Nav is the active sidebar item (general, aws, tailscale, runtime, workspaces).
	Nav string
	// RuntimeOffered mirrors Config.AgentRuntime.
	RuntimeOffered bool
	// TailscaleOffered mirrors Config.AgentTailscale.
	TailscaleOffered bool
	// RepositoriesOffered is true when the portal has GitHub credentials to power the
	// Repositories page.
	RepositoriesOffered bool
	Runtime             runtimeForm
	TailscaleForm       tailscaleForm
	// Repositories is the agent's current (or just-submitted) "owner/repo" list, shown as
	// chips on the Repositories page.
	Repositories []string
	// Onboarding drives the post-create walkthrough. Its zero value renders the page as a
	// standalone edit screen.
	Onboarding onboarding
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var (
		agents []agentsv1.Agent
		err    error
	)
	if user.Admin {
		agents, err = s.cfg.Store.List(ctx)
	} else {
		agents, err = s.cfg.Store.ListByOwner(ctx, user.Sub)
	}
	if err != nil {
		s.fail(w, "listing agents", err)
		return
	}

	s.render(w, "list", pageData{Title: "Your agents", User: user, Agents: agents})
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	s.entCache.Warm(user.Sub)
	s.render(w, "form", pageData{
		Title:      "Register agent",
		User:       user,
		Onboarding: onboarding{Steps: s.onboardingSteps("")},
	})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	renderErr := func(msg string) {
		s.render(w, "form", pageData{
			Title:      "Register agent",
			User:       user,
			Error:      msg,
			Onboarding: onboarding{Steps: s.onboardingSteps("")},
		})
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if !agentNameRe.MatchString(name) {
		renderErr("Name must be non-empty and use only letters, numbers, dashes, or underscores.")
		return
	}

	existing, err := s.cfg.Store.Get(ctx, name)
	if err != nil {
		s.fail(w, "checking existing agent", err)
		return
	}
	if existing != nil {
		renderErr(fmt.Sprintf("An agent named %q already exists.", name))
		return
	}

	displayName := strings.TrimSpace(r.FormValue("display-name"))
	if displayName == "" {
		displayName = name
	}

	agent := &agentsv1.Agent{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: agentsv1.AgentSpec{
			DisplayName: displayName,
			Owner:       user.Sub,
			OwnerEmail:  user.Email,
		},
	}
	if s.cfg.AgentRuntime {
		agent.Spec.Runtime, agent.Spec.Workspaces = defaultRuntime(s.limits())
	}
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "saving agent", err)
		return
	}

	// The walkthrough ends on AWS access, whose Okta lookup is the slowest thing the portal
	// does. Starting it here means it has the earlier steps to finish in.
	s.entCache.Warm(user.Sub)

	s.redirect(w, r, s.onboardingStart(name))
}

func (s *Server) handleGeneral(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.render(w, "agent_general", pageData{
		Title: "Edit " + agent.Name,
		User:  user,
		Agent: agent,
		Nav:   "general",
	})
}

func (s *Server) handleUpdateGeneral(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	displayName := strings.TrimSpace(r.FormValue("display-name"))
	if displayName == "" {
		displayName = agent.Name
	}
	agent.Spec.DisplayName = displayName
	err = s.cfg.Store.Upsert(r.Context(), agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name)
}

func (s *Server) handleAWS(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	ent, err := s.entitlements(r.Context(), agent.Spec.Owner)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}
	s.render(w, "agent_aws", pageData{
		Title:        "AWS access — " + agent.Name,
		User:         user,
		Agent:        agent,
		Nav:          "aws",
		Entitlements: ent,
		Checked:      checkedFromAgent(agent),
		Onboarding:   s.onboardingFor(r, agent.Name, "aws"),
	})
}

func (s *Server) handleUpdateAWS(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ent, err := s.entitlements(ctx, agent.Spec.Owner)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}
	renderErr := func(msg string) {
		s.render(w, "agent_aws", pageData{
			Title: "AWS access — " + agent.Name, User: user, Agent: agent, Nav: "aws",
			Entitlements: ent, Checked: checkedFromForm(r), Error: msg,
			Onboarding: s.onboardingFor(r, agent.Name, "aws"),
		})
	}
	grants, err := parseGrants(r, ent)
	if err != nil {
		renderErr(err.Error())
		return
	}
	agent.Spec.Grants = grants
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirectAfterSave(w, r, agent.Name, "aws")
}

func (s *Server) handleTailscale(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.AgentTailscale {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	tf := tailscaleFormFromAgent(agent)
	if tf.SSHUser == "" {
		if derived, err := deriveTailscaleUser(agent.Spec.OwnerEmail); err == nil {
			tf.SSHUser = derived
		}
	}
	s.entCache.Warm(user.Sub)
	s.render(w, "agent_tailscale", pageData{
		Title:         "Tailscale — " + agent.Name,
		User:          user,
		Agent:         agent,
		Nav:           "tailscale",
		TailscaleForm: tf,
		Onboarding:    s.onboardingFor(r, agent.Name, "tailscale"),
	})
}

func (s *Server) handleUpdateTailscale(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.AgentTailscale {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	renderErr := func(msg string) {
		s.render(w, "agent_tailscale", pageData{
			Title: "Tailscale — " + agent.Name, User: user, Agent: agent, Nav: "tailscale",
			TailscaleForm: tailscaleFormFromAgent(agent), Error: msg,
			Onboarding: s.onboardingFor(r, agent.Name, "tailscale"),
		})
	}
	if r.FormValue("tailscale") == "on" {
		var sshUser string
		if user.Admin && r.FormValue("ssh-user") != "" {
			sshUser = strings.TrimSpace(r.FormValue("ssh-user"))
			if sshUser == "root" {
				renderErr("root is not allowed as the run-as user")
				return
			}
			if !validSSHUser(sshUser) {
				renderErr("run-as user must start with a letter or underscore and contain only lowercase letters, digits, underscores and hyphens")
				return
			}
		} else {
			var err error
			sshUser, err = deriveTailscaleUser(agent.Spec.OwnerEmail)
			if err != nil {
				renderErr(err.Error())
				return
			}
		}
		agent.Spec.Tailscale = &agentsv1.TailscaleAccess{SSHUser: sshUser}
	} else {
		agent.Spec.Tailscale = nil
	}
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirectAfterSave(w, r, agent.Name, "tailscale")
}

func (s *Server) handleRuntime(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.AgentRuntime {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.entCache.Warm(user.Sub)
	s.render(w, "agent_runtime", pageData{
		Title:      "Runtime — " + agent.Name,
		User:       user,
		Agent:      agent,
		Nav:        "runtime",
		Runtime:    runtimeFromAgent(agent, s.limits()),
		Onboarding: s.onboardingFor(r, agent.Name, "runtime"),
	})
}

func (s *Server) handleUpdateRuntime(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.AgentRuntime {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	renderErr := func(msg string) {
		s.render(w, "agent_runtime", pageData{
			Title: "Runtime — " + agent.Name, User: user, Agent: agent, Nav: "runtime",
			Runtime: runtimeFromForm(r, s.limits()), Error: msg,
			Onboarding: s.onboardingFor(r, agent.Name, "runtime"),
		})
	}
	runtime, workspaces, err := s.parseAgentRuntime(r, agent, user.Admin)
	if err != nil {
		renderErr(err.Error())
		return
	}
	agent.Spec.Runtime = runtime
	agent.Spec.Workspaces = workspaces
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirectAfterSave(w, r, agent.Name, "runtime")
}

func (s *Server) handleRepositories(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Repositories == nil {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.entCache.Warm(user.Sub)
	s.render(w, "agent_repositories", pageData{
		Title:        "Repositories — " + agent.Name,
		User:         user,
		Agent:        agent,
		Nav:          "repositories",
		Repositories: repositoriesFromAgent(agent),
		Onboarding:   s.onboardingFor(r, agent.Name, "repositories"),
	})
}

func (s *Server) handleUpdateRepositories(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Repositories == nil {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	repos := normalizeRepos(r.Form["repository"])
	renderErr := func(msg string) {
		s.render(w, "agent_repositories", pageData{
			Title: "Repositories — " + agent.Name, User: user, Agent: agent, Nav: "repositories",
			Repositories: repos, Error: msg,
			Onboarding: s.onboardingFor(r, agent.Name, "repositories"),
		})
	}
	for _, repo := range repos {
		if !repoRe.MatchString(repo) {
			renderErr(fmt.Sprintf("%q is not a valid owner/repo.", repo))
			return
		}
		reachable, err := s.cfg.Repositories.Reachable(ctx, repo)
		if err != nil {
			s.fail(w, "checking repository access", err)
			return
		}
		if !reachable {
			renderErr(fmt.Sprintf("The agent's GitHub App cannot reach %q. Pick a repository from the suggestions.", repo))
			return
		}
	}
	agent.Spec.Repositories = toRepositories(repos)
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirectAfterSave(w, r, agent.Name, "repositories")
}

// handleRepositorySearch answers the type-ahead with a JSON array of "owner/repo" strings the
// agent's GitHub App can reach and that match the q query.
func (s *Server) handleRepositorySearch(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Repositories == nil {
		http.NotFound(w, r)
		return
	}
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	if _, ok := s.ownedAgent(w, r, user); !ok {
		return
	}
	repos, err := s.cfg.Repositories.Suggest(r.Context(), r.URL.Query().Get("q"), 20)
	if err != nil {
		slog.Warn("repository suggestions failed", "error", err)
		http.Error(w, "suggestions unavailable", http.StatusBadGateway)
		return
	}
	if repos == nil {
		repos = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		slog.Error("encoding repository suggestions", "error", err)
	}
}

func repositoriesFromAgent(agent *agentsv1.Agent) []string {
	repos := make([]string, 0, len(agent.Spec.Repositories))
	for _, repo := range agent.Spec.Repositories {
		repos = append(repos, string(repo))
	}
	return repos
}

// normalizeRepos trims each value, drops blanks, and removes case-insensitive duplicates while
// keeping the submitted order.
func normalizeRepos(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	return out
}

func toRepositories(values []string) []agentsv1.Repository {
	if len(values) == 0 {
		return nil
	}
	repos := make([]agentsv1.Repository, len(values))
	for i, v := range values {
		repos[i] = agentsv1.Repository(v)
	}
	return repos
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}

	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}

	err := s.cfg.Store.Delete(r.Context(), agent.Name)
	if err != nil {
		s.fail(w, "deleting agent", err)
		return
	}
	s.redirect(w, r, "/")
}

func (s *Server) handleWorkspacesView(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.render(w, "workspaces", pageData{
		Title: "Workspaces — " + agent.Name,
		User:  user,
		Agent: agent,
		Nav:   "workspaces",
	})
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.render(w, "connection", pageData{
		Title: "Connection — " + agent.Name,
		User:  user,
		Agent: agent,
		Nav:   "connection",
	})
}

func (s *Server) handleSpawnWorkspace(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	renderErr := func(msg string) {
		s.render(w, "workspaces", pageData{
			Title: "Workspaces — " + agent.Name,
			User:  user, Agent: agent, Nav: "workspaces", Error: msg,
		})
	}

	name := strings.ToLower(strings.TrimSpace(r.FormValue("new-workspace")))
	if name == "" {
		renderErr("Workspace name is required.")
		return
	}
	if len(name) > workspaceNameMaxLength {
		renderErr(fmt.Sprintf("Workspace names are limited to %d characters.", workspaceNameMaxLength))
		return
	}
	if !workspaceNameRe.MatchString(name) {
		renderErr(fmt.Sprintf("Workspace name %q must use only lowercase letters, numbers, and dashes.", name))
		return
	}
	for _, t := range agent.Spec.Workspaces {
		if t.Name == name {
			renderErr(fmt.Sprintf("A workspace named %q already exists.", name))
			return
		}
	}
	maxWorkspaces := s.limits().defaults().MaxWorkspaces
	if len(agent.Spec.Workspaces)+1 > maxWorkspaces {
		renderErr(fmt.Sprintf("An agent is limited to %d workspaces.", maxWorkspaces))
		return
	}

	agent.Spec.Workspaces = append(agent.Spec.Workspaces, agentsv1.AgentWorkspace{Name: name})
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "spawning workspace", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name+"/workspaces")
}

func (s *Server) handleToggleSuspend(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	workspaceName := r.PathValue("workspace")
	for i, t := range agent.Spec.Workspaces {
		if t.Name == workspaceName {
			agent.Spec.Workspaces[i].Suspended = !agent.Spec.Workspaces[i].Suspended
			err := s.cfg.Store.Upsert(ctx, agent)
			if err != nil {
				s.fail(w, "toggling workspace suspend", err)
				return
			}
			s.redirect(w, r, "/agents/"+agent.Name+"/workspaces")
			return
		}
	}
	http.Error(w, "workspace not found", http.StatusNotFound)
}

func (s *Server) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	workspaceName := r.PathValue("workspace")
	workspaces := agent.Spec.Workspaces[:0]
	for _, t := range agent.Spec.Workspaces {
		if t.Name != workspaceName {
			workspaces = append(workspaces, t)
		}
	}
	agent.Spec.Workspaces = workspaces
	err := s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "deleting workspace", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name+"/workspaces")
}

// ownedAgent loads the agent named in the path and enforces that the current user may act
// on it (owner or admin). It writes the response and returns ok=false on any failure.
func (s *Server) ownedAgent(w http.ResponseWriter, r *http.Request, user *identity.User) (*agentsv1.Agent, bool) {
	name := r.PathValue("name")
	agent, err := s.cfg.Store.Get(r.Context(), name)
	if err != nil {
		s.fail(w, "loading agent", err)
		return nil, false
	}
	if agent == nil {
		slog.Warn("portal answered not found", "agent", name, "sub", user.Sub)
		http.Error(w, "agent not found", http.StatusNotFound)
		return nil, false
	}
	if agent.Spec.Owner != user.Sub && !user.Admin {
		slog.Warn("portal answered forbidden",
			"agent", name,
			"sub", user.Sub,
			"owner", agent.Spec.Owner,
			"admin", user.Admin,
		)
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, false
	}
	return agent, true
}

func (s *Server) entitlements(ctx context.Context, sub string) (*Entitlements, error) {
	return s.entCache.Get(ctx, sub)
}

func (s *Server) fetchEntitlements(ctx context.Context, sub string) (*Entitlements, error) {
	mappings, err := s.cfg.MappingsProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading rolemap: %w", err)
	}
	return ResolveEntitlements(ctx, sub, s.cfg.Apps, mappings)
}

// limits returns AgentLimits for the current request. ConfigMap loader values fill any zero
// fields in the static Config.Limits, so a ConfigMap update propagates without a restart.
func (s *Server) limits() AgentLimits {
	l := s.cfg.Limits
	if s.cfg.DefaultsLoader == nil {
		return l
	}
	d, err := s.cfg.DefaultsLoader.Load()
	if err != nil {
		slog.Warn("loading agent defaults in portal", "error", err)
		return l
	}
	if l.MaxCPU == "" {
		l.MaxCPU = d.MaxCPU
	}
	if l.MaxMemory == "" {
		l.MaxMemory = d.MaxMemory
	}
	if l.MaxWorkspace == "" {
		l.MaxWorkspace = d.MaxWorkspace
	}
	if l.MaxWorkspaces == 0 {
		l.MaxWorkspaces = d.MaxWorkspaces
	}
	if l.DefaultImage == "" {
		l.DefaultImage = d.Image
	}
	if l.DefaultStorageClass == "" {
		l.DefaultStorageClass = d.StorageClass
	}
	return l
}

func (s *Server) user(w http.ResponseWriter, r *http.Request) (*identity.User, bool) {
	user, err := s.cfg.Identity.Resolve(r.Context(), r)
	if err != nil {
		slog.Warn("portal answered unauthorized",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"error", err,
		)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}

// parseAgentRuntime reads the runtime section of a submission, or leaves the agent's runtime
// alone in an environment where the portal does not offer it. Without that guard a form posted
// against a portal with the runtime disabled would silently clear an existing runtime.
func (s *Server) parseAgentRuntime(r *http.Request, current *agentsv1.Agent, isAdmin bool) (*agentsv1.AgentRuntime, []agentsv1.AgentWorkspace, error) {
	if !s.cfg.AgentRuntime {
		if current == nil {
			return nil, nil, nil
		}
		return current.Spec.Runtime, current.Spec.Workspaces, nil
	}
	return parseRuntime(r, current, s.limits(), isAdmin)
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	data.BasePath = s.basePath
	data.Namespace = s.cfg.Namespace
	data.RuntimeOffered = s.cfg.AgentRuntime
	data.TailscaleOffered = s.cfg.AgentTailscale
	data.RepositoriesOffered = s.cfg.Repositories != nil
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		slog.Error("rendering template", "template", name, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (s *Server) fail(w http.ResponseWriter, what string, err error) {
	slog.Error(what, "error", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

// parseGrants reads the selected grants from the form and validates each one is within the
// entitlements, enforcing that an agent gets only a subset of the owner's access.
func parseGrants(r *http.Request, ent *Entitlements) ([]agentsv1.Grant, error) {
	selected := r.Form["grant"]
	grants := make([]agentsv1.Grant, 0, len(selected))
	for _, raw := range selected {
		accountID, roleARN, ok := strings.Cut(raw, "|")
		if !ok {
			return nil, fmt.Errorf("malformed selection %q", raw)
		}
		grant, allowed := ent.Allows(accountID, roleARN)
		if !allowed {
			return nil, fmt.Errorf("you do not have access to %s in account %s", roleARN, accountID)
		}
		awsGrant := grant
		grants = append(grants, agentsv1.Grant{AWS: &awsGrant})
	}
	return grants, nil
}

func checkedFromForm(r *http.Request) map[string]bool {
	checked := map[string]bool{}
	for _, raw := range r.Form["grant"] {
		checked[raw] = true
	}
	return checked
}

func checkedFromAgent(agent *agentsv1.Agent) map[string]bool {
	checked := map[string]bool{}
	for _, g := range agent.Spec.Grants {
		if key := g.Key(); key != "" {
			checked[key] = true
		}
	}
	return checked
}

type recoveryLogger struct {
	logger *slog.Logger
}

func (l recoveryLogger) Println(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}
