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
	// Limits caps the sizing an owner may ask for. Unset fields fall back to defaults.
	Limits AgentLimits
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
	mux.HandleFunc("GET /agents/{name}", s.handleView)
	mux.HandleFunc("POST /agents/{name}", s.handleUpdate)
	mux.HandleFunc("POST /agents/{name}/delete", s.handleDelete)

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
	User         *identity.User
	Agents       []agentsv1.Agent
	Agent        *agentsv1.Agent
	Entitlements *Entitlements
	Checked      map[string]bool
	Action       string
	Error        string
	// RuntimeOffered mirrors Config.AgentRuntime, so the form hides the section entirely in an
	// environment where agents do not run in the cluster.
	RuntimeOffered bool
	Runtime        runtimeForm
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
	ent, err := s.entitlements(r.Context(), user.Sub)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}
	s.render(w, "form", pageData{
		Title:        "Register agent",
		User:         user,
		Entitlements: ent,
		Checked:      map[string]bool{},
		Action:       "/agents",
		Runtime:      runtimeFromAgent(nil, s.cfg.Limits),
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

	ent, err := s.entitlements(ctx, user.Sub)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	renderErr := func(msg string) {
		s.render(w, "form", pageData{
			Title: "Register agent", User: user, Entitlements: ent,
			Checked: checkedFromForm(r), Action: "/agents", Error: msg,
			Runtime: runtimeFromForm(r, s.cfg.Limits),
		})
	}

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

	grants, err := parseGrants(r, ent)
	if err != nil {
		renderErr(err.Error())
		return
	}

	runtime, threads, err := s.parseAgentRuntime(r, nil)
	if err != nil {
		renderErr(err.Error())
		return
	}

	agent := &agentsv1.Agent{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: agentsv1.AgentSpec{
			DisplayName: name,
			Owner:       user.Sub,
			OwnerEmail:  user.Email,
			Grants:      grants,
			Runtime:     runtime,
			Threads:     threads,
		},
	}
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "saving agent", err)
		return
	}

	s.redirect(w, r, "/")
}

func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	agent, ok := s.ownedAgent(w, r, user)
	if !ok {
		return
	}

	// Bound the choices by the owner's access, not the (possibly admin) editor's.
	ent, err := s.entitlements(ctx, agent.Spec.Owner)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}

	s.render(w, "form", pageData{
		Title:        "Edit " + agent.Name,
		User:         user,
		Agent:        agent,
		Entitlements: ent,
		Checked:      checkedFromAgent(agent),
		Action:       "/agents/" + agent.Name,
		Runtime:      runtimeFromAgent(agent, s.cfg.Limits),
	})
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
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
		s.render(w, "form", pageData{
			Title: "Edit " + agent.Name, User: user, Agent: agent, Entitlements: ent,
			Checked: checkedFromForm(r), Action: "/agents/" + agent.Name, Error: msg,
			Runtime: runtimeFromForm(r, s.cfg.Limits),
		})
	}

	grants, err := parseGrants(r, ent)
	if err != nil {
		renderErr(err.Error())
		return
	}

	runtime, threads, err := s.parseAgentRuntime(r, agent)
	if err != nil {
		renderErr(err.Error())
		return
	}

	agent.Spec.Grants = grants
	agent.Spec.Runtime = runtime
	agent.Spec.Threads = threads
	err = s.cfg.Store.Upsert(ctx, agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}

	s.redirect(w, r, "/")
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
	return parseRuntime(r, current, s.cfg.Limits)
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	data.BasePath = s.basePath
	data.RuntimeOffered = s.cfg.AgentRuntime
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
