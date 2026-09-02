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

// Config wires the portal's dependencies. BasePath is the URL prefix the portal is served
// under (for example "/portal" when the gateway routes a sub-path to it); empty means root.
type Config struct {
	Apps             okta.AppLister
	MappingsProvider awsaccess.MappingsProvider
	Store            agentstore.AgentStore
	Identity         *IdentityResolver
	BasePath         string
	// AgentRuntime offers the option of running an agent's threads as pods. It is off unless
	// the operator can actually run them, so the form does not promise what it cannot deliver.
	AgentRuntime bool
	// AgentTailscale offers the Tailscale page. Off unless the operator is configured for
	// tailnet enrollment, so the nav item is not shown in environments without tailscale.
	AgentTailscale bool
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
	tmpl, err := template.New("").ParseFS(templatesFS, "templates/*.html")
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
	mux.HandleFunc("POST /agents/{name}/delete", s.handleDelete)
	mux.HandleFunc("GET /agents/{name}/threads", s.handleThreadsView)
	mux.HandleFunc("POST /agents/{name}/threads", s.handleSpawnThread)
	mux.HandleFunc("POST /agents/{name}/threads/{thread}/suspend", s.handleToggleSuspend)
	mux.HandleFunc("POST /agents/{name}/threads/{thread}/delete", s.handleDeleteThread)

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
	// Nav is the active sidebar item (general, aws, tailscale, runtime, threads).
	Nav string
	// RuntimeOffered mirrors Config.AgentRuntime.
	RuntimeOffered bool
	// TailscaleOffered mirrors Config.AgentTailscale.
	TailscaleOffered bool
	Runtime          runtimeForm
	TailscaleForm    tailscaleForm
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
	s.render(w, "form", pageData{Title: "Register agent", User: user})
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
		s.render(w, "form", pageData{Title: "Register agent", User: user, Error: msg})
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
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "saving agent", err)
		return
	}

	s.redirect(w, r, "/agents/"+name+"/aws")
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
	s.redirect(w, r, "/agents/"+agent.Name+"/aws")
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
	s.render(w, "agent_tailscale", pageData{
		Title:         "Tailscale — " + agent.Name,
		User:          user,
		Agent:         agent,
		Nav:           "tailscale",
		TailscaleForm: tf,
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
	s.redirect(w, r, "/agents/"+agent.Name+"/tailscale")
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
	s.render(w, "agent_runtime", pageData{
		Title:   "Runtime — " + agent.Name,
		User:    user,
		Agent:   agent,
		Nav:     "runtime",
		Runtime: runtimeFromAgent(agent, s.limits()),
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
		})
	}
	runtime, threads, err := s.parseAgentRuntime(r, agent)
	if err != nil {
		renderErr(err.Error())
		return
	}
	agent.Spec.Runtime = runtime
	agent.Spec.Threads = threads
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name+"/runtime")
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

func (s *Server) handleThreadsView(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	s.render(w, "threads", pageData{
		Title: "Threads — " + agent.Name,
		User:  user,
		Agent: agent,
		Nav:   "threads",
	})
}

func (s *Server) handleSpawnThread(w http.ResponseWriter, r *http.Request) {
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
		s.render(w, "threads", pageData{
			Title: "Threads — " + agent.Name,
			User:  user, Agent: agent, Nav: "threads", Error: msg,
		})
	}

	name := strings.ToLower(strings.TrimSpace(r.FormValue("new-thread")))
	if name == "" {
		renderErr("Thread name is required.")
		return
	}
	if len(name) > threadNameMaxLength {
		renderErr(fmt.Sprintf("Thread names are limited to %d characters.", threadNameMaxLength))
		return
	}
	if !threadNameRe.MatchString(name) {
		renderErr(fmt.Sprintf("Thread name %q must use only lowercase letters, numbers, and dashes.", name))
		return
	}
	for _, t := range agent.Spec.Threads {
		if t.Name == name {
			renderErr(fmt.Sprintf("A thread named %q already exists.", name))
			return
		}
	}
	maxThreads := s.limits().defaults().MaxThreads
	if len(agent.Spec.Threads)+1 > maxThreads {
		renderErr(fmt.Sprintf("An agent is limited to %d threads.", maxThreads))
		return
	}

	agent.Spec.Threads = append(agent.Spec.Threads, agentsv1.AgentThread{Name: name})
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "spawning thread", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name+"/threads")
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
	threadName := r.PathValue("thread")
	for i, t := range agent.Spec.Threads {
		if t.Name == threadName {
			agent.Spec.Threads[i].Suspended = !agent.Spec.Threads[i].Suspended
			err := s.cfg.Store.Upsert(ctx, agent)
			if err != nil {
				s.fail(w, "toggling thread suspend", err)
				return
			}
			s.redirect(w, r, "/agents/"+agent.Name+"/threads")
			return
		}
	}
	http.Error(w, "thread not found", http.StatusNotFound)
}

func (s *Server) handleDeleteThread(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}
	threadName := r.PathValue("thread")
	threads := agent.Spec.Threads[:0]
	for _, t := range agent.Spec.Threads {
		if t.Name != threadName {
			threads = append(threads, t)
		}
	}
	agent.Spec.Threads = threads
	err := s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "deleting thread", err)
		return
	}
	s.redirect(w, r, "/agents/"+agent.Name+"/threads")
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
	if l.MaxThreads == 0 {
		l.MaxThreads = d.MaxThreads
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
func (s *Server) parseAgentRuntime(r *http.Request, current *agentsv1.Agent) (*agentsv1.AgentRuntime, []agentsv1.AgentThread, error) {
	if !s.cfg.AgentRuntime {
		if current == nil {
			return nil, nil, nil
		}
		return current.Spec.Runtime, current.Spec.Threads, nil
	}
	return parseRuntime(r, current, s.limits())
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	data.BasePath = s.basePath
	data.Namespace = s.cfg.Namespace
	data.RuntimeOffered = s.cfg.AgentRuntime
	data.TailscaleOffered = s.cfg.AgentTailscale
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
