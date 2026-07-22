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

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

//go:embed templates/*.html
var templatesFS embed.FS

var agentNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Config wires the portal's dependencies.
type Config struct {
	Apps             okta.AppLister
	MappingsProvider MappingsProvider
	Store            AgentStore
	Identity         *IdentityResolver
}

// Server is the agent-registry portal.
type Server struct {
	cfg  Config
	tmpl *template.Template
}

// NewServer parses templates and returns a portal server.
func NewServer(cfg Config) (*Server, error) {
	tmpl, err := template.New("").ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	return &Server{cfg: cfg, tmpl: tmpl}, nil
}

// Handler returns the HTTP handler for the portal.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /{$}", s.handleList)
	mux.HandleFunc("GET /agents/new", s.handleNew)
	mux.HandleFunc("POST /agents", s.handleCreate)
	mux.HandleFunc("GET /agents/{name}", s.handleView)
	mux.HandleFunc("POST /agents/{name}", s.handleUpdate)
	mux.HandleFunc("POST /agents/{name}/delete", s.handleDelete)

	recovery := handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(recoveryLogger{slog.Default()}),
	)
	return recovery(mux)
}

type pageData struct {
	Title        string
	User         *User
	Agents       []Agent
	Agent        *Agent
	Entitlements *Entitlements
	Checked      map[string]bool
	Action       string
	Error        string
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	user, ok := s.user(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var (
		agents []Agent
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

	now := time.Now().UTC()
	err = s.cfg.Store.Upsert(ctx, Agent{
		Name:       name,
		Owner:      user.Sub,
		OwnerEmail: user.Email,
		Grants:     grants,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		s.fail(w, "saving agent", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	ent, err := s.entitlements(ctx, agent.Owner)
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

	ent, err := s.entitlements(ctx, agent.Owner)
	if err != nil {
		s.fail(w, "resolving entitlements", err)
		return
	}

	grants, err := parseGrants(r, ent)
	if err != nil {
		s.render(w, "form", pageData{
			Title: "Edit " + agent.Name, User: user, Agent: agent, Entitlements: ent,
			Checked: checkedFromForm(r), Action: "/agents/" + agent.Name, Error: err.Error(),
		})
		return
	}

	agent.Grants = grants
	agent.UpdatedAt = time.Now().UTC()
	err = s.cfg.Store.Upsert(ctx, *agent)
	if err != nil {
		s.fail(w, "updating agent", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ownedAgent loads the agent named in the path and enforces that the current user may act
// on it (owner or admin). It writes the response and returns ok=false on any failure.
func (s *Server) ownedAgent(w http.ResponseWriter, r *http.Request, user *User) (*Agent, bool) {
	name := r.PathValue("name")
	agent, err := s.cfg.Store.Get(r.Context(), name)
	if err != nil {
		s.fail(w, "loading agent", err)
		return nil, false
	}
	if agent == nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return nil, false
	}
	if agent.Owner != user.Sub && !user.Admin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, false
	}
	return agent, true
}

func (s *Server) entitlements(ctx context.Context, sub string) (*Entitlements, error) {
	mappings, err := s.cfg.MappingsProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading rolemap: %w", err)
	}
	return ResolveEntitlements(ctx, sub, s.cfg.Apps, mappings)
}

func (s *Server) user(w http.ResponseWriter, r *http.Request) (*User, bool) {
	user, err := s.cfg.Identity.Resolve(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
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
func parseGrants(r *http.Request, ent *Entitlements) ([]Grant, error) {
	selected := r.Form["grant"]
	grants := make([]Grant, 0, len(selected))
	for _, raw := range selected {
		accountID, roleARN, ok := strings.Cut(raw, "|")
		if !ok {
			return nil, fmt.Errorf("malformed selection %q", raw)
		}
		grant, allowed := ent.Allows(accountID, roleARN)
		if !allowed {
			return nil, fmt.Errorf("you do not have access to %s in account %s", roleARN, accountID)
		}
		grants = append(grants, grant)
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

func checkedFromAgent(agent *Agent) map[string]bool {
	checked := map[string]bool{}
	for _, g := range agent.Grants {
		checked[g.Key()] = true
	}
	return checked
}

type recoveryLogger struct {
	logger *slog.Logger
}

func (l recoveryLogger) Println(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}
