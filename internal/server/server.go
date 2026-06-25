// Package server is the HTTP front-end for the web UI. It mirrors the MCP verbs over
// HTTP by reusing mcp.Service, so the web and agent front-ends share one rule-set and
// can't drift (SPEC §0).
//
// Every endpoint accepts a `?path=` (the project folder to operate on), falling back to
// the server's launch `--repo`. Arbitrary path access is intentional: cairn web is a
// local, single-user tool. Paths must resolve to an existing directory.
//
// Endpoints:
//
//	GET  /api/status                       -> { initialized, root, prefix?, suggestedPrefix, states, closed, initial }
//	POST /api/init        { path?, prefix? }
//	GET  /api/tasks       ?status&assignee&ready
//	POST /api/tasks       { title, body?, deps?, checks? }
//	GET    /api/tasks/{id}
//	DELETE /api/tasks/{id}                            -> { id, deleted } ; refused if it has children/dependents
//	GET    /api/tasks/{id}/runs            -> { runs: [{ file, at, cmd, cwd, exit, timedout, duration, output }] }
//	POST   /api/tasks/{id}/update    { title?, body?, checks?, priority?, labels?, parent? }
//	POST   /api/tasks/{id}/transition  { to }
//	POST   /api/tasks/{id}/claim
//	POST   /api/tasks/{id}/run_checks  { only? }
//	POST   /api/tasks/{id}/attest      { index, pass? }   -> attest a manual check (pass defaults true)
//	POST   /api/tasks/{id}/note        { text }
//	PATCH  /api/tasks/{id}/notes/{note}  { text }         -> edit a note (?index= for a legacy note, {note}="-")
//	DELETE /api/tasks/{id}/notes/{note}                   -> delete a note (?index= for a legacy note, {note}="-")
//	GET    /api/events      ?path                  -> text/event-stream of { type, id? } change signals
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/mcp"
	"cairn/internal/repo"
	"cairn/internal/session"
	"cairn/internal/store"
	"cairn/internal/task"
	"cairn/web"
)

// Server serves the web API. defaultRoot is used when a request omits `path`; actor is
// stamped on every write as provenance.
type Server struct {
	defaultRoot string
	actor       string
	hub         *Hub
}

// New returns a Server. actor defaults to human:web.
func New(defaultRoot, actor string) *Server {
	if actor == "" {
		actor = "human:web"
	}
	return &Server{defaultRoot: defaultRoot, actor: actor, hub: NewHub(0)}
}

// Handler builds the HTTP routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/identity", s.handleIdentity)
	mux.HandleFunc("POST /api/init", s.handleInit)
	mux.HandleFunc("GET /api/tasks", s.handleList)
	mux.HandleFunc("POST /api/tasks", s.handleCreate)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGet)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDelete)
	mux.HandleFunc("GET /api/tasks/{id}/runs", s.handleRuns)
	mux.HandleFunc("GET /api/tasks/{id}/sessions", s.handleListTaskSessions)
	mux.HandleFunc("POST /api/tasks/{id}/sessions/begin", s.handleBeginSession)
	mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	mux.HandleFunc("GET /api/sessions/{session}", s.handleGetSession)
	mux.HandleFunc("POST /api/sessions/{session}/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("POST /api/sessions/{session}/finish", s.handleFinishSession)
	mux.HandleFunc("POST /api/sessions/{session}/cancel", s.handleCancelSession)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("POST /api/tasks/{id}/update", s.handleUpdate)
	mux.HandleFunc("POST /api/tasks/{id}/reorder", s.handleReorder)
	mux.HandleFunc("POST /api/tasks/{id}/transition", s.handleTransition)
	mux.HandleFunc("POST /api/tasks/{id}/claim", s.handleClaim)
	mux.HandleFunc("POST /api/tasks/{id}/run_checks", s.handleRunChecks)
	mux.HandleFunc("POST /api/tasks/{id}/attest", s.handleAttest)
	mux.HandleFunc("POST /api/tasks/{id}/note", s.handleNote)
	mux.HandleFunc("PATCH /api/tasks/{id}/notes/{note}", s.handleEditNote)
	mux.HandleFunc("DELETE /api/tasks/{id}/notes/{note}", s.handleDeleteNote)

	// Integrations: detect agents and write their MCP config (one-click connect).
	mux.HandleFunc("GET /api/connect", s.handleListIntegrations)
	mux.HandleFunc("POST /api/connect/{agent}", s.handleConnectAgent)
	mux.HandleFunc("DELETE /api/connect/{agent}", s.handleDisconnectAgent)
	mux.HandleFunc("GET /api/connect/{agent}/manual", s.handleAgentManual)

	// Readiness probe for the Tauri shell: no ?path needed, returns once the
	// server is accepting requests.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// MCP over Streamable HTTP, so an agent can connect to the running app by URL
	// (the same rule-set as `cairn serve` and the /api front-end). Each connection
	// binds its own repo + actor via the query: /mcp?repo=/abs/path&actor=agent:x.
	mux.Handle("/mcp", s.mcpHandler())
	mux.Handle("/mcp/", s.mcpHandler())

	// Embedded UI last: the catch-all serves the SPA, falling back to index.html.
	mux.Handle("/", spaHandler(web.FS()))
	return mux
}

// mcpHandler serves MCP over Streamable HTTP. It validates ?repo and ?actor up
// front (clear 400s) and builds a per-connection mcp.Server bound to that repo
// and identity, reusing the same Service rule-set as the stdio server.
func (s *Server) mcpHandler() http.Handler {
	h := mcpsdk.NewStreamableHTTPHandler(func(r *http.Request) *mcpsdk.Server {
		root, err := s.resolveRoot(r.URL.Query().Get("repo"))
		if err != nil {
			return nil
		}
		actor := r.URL.Query().Get("actor")
		if actor == "" {
			return nil
		}
		// Auto-init so a freshly opened project just works (mirrors `cairn serve`).
		if err := repo.Init(root, ""); err != nil {
			return nil
		}
		client := r.URL.Query().Get("client")
		return mcp.NewServer(mcp.NewServiceWithClient(store.New(root), actor, client, nil))
	}, nil)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("actor") == "" {
			http.Error(w, "missing ?actor= (e.g. agent:claude-1)", http.StatusBadRequest)
			return
		}
		if _, err := s.resolveRoot(r.URL.Query().Get("repo")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// spaHandler serves embedded static assets, falling back to index.html so
// client-side routes resolve. Tolerates a missing index.html (placeholder dist)
// by letting the file server return its own 404.
func spaHandler(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(root, strings.TrimPrefix(r.URL.Path, "/")); err != nil && r.URL.Path != "/" {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

// Run serves on addr until the process exits.
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

// --- request helpers ---

// resolveRoot turns a raw path (query/body) into an absolute, existing directory.
func (s *Server) resolveRoot(raw string) (string, error) {
	if raw == "" {
		raw = s.defaultRoot
	}
	raw = expandHome(raw)
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("no such folder: %s", abs)
	}
	return abs, nil
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func (s *Server) service(root, actor string) *mcp.Service {
	return mcp.NewService(store.New(root), actor, nil)
}

// actorFor resolves who is making this request: the client-asserted identity from the
// X-Cairn-Actor header (URL-encoded; falls back to ?actor=), sanitized, else the server
// default. Trust model is local-dev (like a git author) — no auth.
func (s *Server) actorFor(r *http.Request) string {
	raw := r.Header.Get("X-Cairn-Actor")
	if dec, err := url.QueryUnescape(raw); err == nil {
		raw = dec
	}
	if raw == "" {
		raw = r.URL.Query().Get("actor")
	}
	if a := sanitizeActor(raw); a != "" {
		return a
	}
	return s.actor
}

// sanitizeActor keeps an actor string safe to store in YAML/provenance: single line, trimmed,
// bounded length. Returns "" when nothing usable remains.
func sanitizeActor(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range raw {
		if r == '\n' || r == '\r' || r == '\t' || r < 0x20 {
			continue // drop control chars / newlines
		}
		b.WriteRune(r)
		if b.Len() >= 64 {
			break
		}
	}
	return strings.TrimSpace(b.String())
}

// --- handlers ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, s.status(root))
}

type initReq struct {
	Path   string `json:"path"`
	Prefix string `json:"prefix"`
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	var req initReq
	decode(r, &req)
	root, err := s.resolveRoot(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	if err := repo.Init(root, req.Prefix); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.status(root))
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	q := r.URL.Query()
	var ready *bool
	if v := q.Get("ready"); v != "" {
		b, _ := strconv.ParseBool(v)
		ready = &b
	}
	views, err := s.service(root, s.actor).ListWithExecution(q.Get("status"), q.Get("assignee"), ready, q.Get("execution"))
	if err != nil {
		writeErr(w, err)
		return
	}
	tasks := make([]taskDTO, 0, len(views))
	for _, v := range views {
		dto := dtoFromTask(v.Task, v.Ready)
		dto.UpdatedAt = v.UpdatedAt
		dto.ExecutionState = v.ExecutionState
		dto.SessionID = v.SessionID
		tasks = append(tasks, dto)
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	doc, err := svc.Get(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type createReq struct {
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Deps     []string   `json:"deps"`
	Checks   []checkDTO `json:"checks"`
	Labels   []string   `json:"labels"`
	Priority string     `json:"priority"`
	Parent   string     `json:"parent"`
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req createReq
	decode(r, &req)
	checks := make([]task.Check, 0, len(req.Checks))
	for _, c := range req.Checks {
		checks = append(checks, task.Check{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Cwd: c.Cwd, Timeout: c.Timeout, Result: "pending"})
	}
	doc, err := svc.Create(store.Draft{
		Title: req.Title, Body: req.Body, Deps: req.Deps, Checks: checks,
		Labels: req.Labels, Priority: req.Priority, Parent: req.Parent,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type updateReq struct {
	Priority *string     `json:"priority"`
	Labels   *[]string   `json:"labels"`
	Parent   *string     `json:"parent"`
	Title    *string     `json:"title"`
	Body     *string     `json:"body"`
	Checks   *[]checkDTO `json:"checks"`
}

type reorderReq struct {
	Rank float64 `json:"rank"`
}

func (s *Server) handleReorder(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req reorderReq
	decode(r, &req)
	doc, err := svc.Reorder(r.PathValue("id"), req.Rank)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req updateReq
	decode(r, &req)
	f := mcp.UpdateFields{Priority: req.Priority, Labels: req.Labels, Parent: req.Parent, Title: req.Title, Body: req.Body}
	if req.Checks != nil {
		checks := make([]task.Check, 0, len(*req.Checks))
		for _, c := range *req.Checks {
			result := c.Result
			if result == "" {
				result = "pending"
			}
			checks = append(checks, task.Check{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Cwd: c.Cwd, Timeout: c.Timeout, Result: result})
		}
		f.Checks = &checks
	}
	doc, err := svc.Update(r.PathValue("id"), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if err := svc.Delete(id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "deleted": true})
}

type editNoteReq struct {
	Text string `json:"text"`
}

// noteRef resolves how a note sub-resource request addresses its target: by stable id
// (the {note} path segment) or, for a legacy note, by 0-based index (?index=, with {note}
// set to the "-" sentinel).
func noteRef(r *http.Request) (noteID string, index int) {
	noteID = r.PathValue("note")
	index = -1
	if v := r.URL.Query().Get("index"); v != "" {
		index, _ = strconv.Atoi(v)
	}
	if noteID == "-" {
		noteID = ""
	}
	return noteID, index
}

func (s *Server) handleEditNote(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req editNoteReq
	decode(r, &req)
	noteID, index := noteRef(r)
	doc, err := svc.EditNote(r.PathValue("id"), noteID, index, req.Text)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	noteID, index := noteRef(r)
	doc, err := svc.DeleteNote(r.PathValue("id"), noteID, index)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type transitionReq struct {
	To string `json:"to"`
}

func (s *Server) handleTransition(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req transitionReq
	decode(r, &req)
	doc, err := svc.Transition(r.PathValue("id"), req.To)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

func (s *Server) handleClaim(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	doc, err := svc.Claim(r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type runChecksReq struct {
	Only []int `json:"only"`
}

func (s *Server) handleRunChecks(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req runChecksReq
	decode(r, &req)
	doc, err := svc.RunChecks(r.PathValue("id"), req.Only)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type attestReq struct {
	Index int   `json:"index"`
	Pass  *bool `json:"pass"`
}

func (s *Server) handleAttest(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req attestReq
	decode(r, &req)
	doc, err := svc.Attest(r.PathValue("id"), req.Index, req.Pass == nil || *req.Pass)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

type noteReq struct {
	Text string `json:"text"`
}

func (s *Server) handleNote(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req noteReq
	decode(r, &req)
	doc, err := svc.Note(r.PathValue("id"), req.Text)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dtoFromDoc(svc, doc))
}

// svcFor resolves the root and returns a Service, writing an error response on failure.
func (s *Server) svcFor(w http.ResponseWriter, r *http.Request) (*mcp.Service, string, bool) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return nil, "", false
	}
	return s.service(root, s.actorFor(r)), root, true
}

func (s *Server) status(root string) statusResp {
	resp := statusResp{
		Initialized:     repo.IsInitialized(root),
		Root:            root,
		SuggestedPrefix: repo.DerivePrefix(root),
		Actor:           s.actor,
		SuggestedActor:  repo.DeriveActor(),
	}
	if resp.Initialized {
		if cfg, err := store.New(root).Config(); err == nil {
			resp.Prefix = cfg.Prefix
			resp.States = cfg.States
			resp.Closed = cfg.Closed
			resp.Initial = cfg.Initial
			resp.Review = cfg.Review()
		}
	}
	return resp
}

// --- DTOs ---

type statusResp struct {
	Initialized     bool     `json:"initialized"`
	Root            string   `json:"root"`
	Prefix          string   `json:"prefix,omitempty"`
	SuggestedPrefix string   `json:"suggestedPrefix"`
	States          []string `json:"states,omitempty"`
	Closed          []string `json:"closed,omitempty"`
	Initial         string   `json:"initial,omitempty"`
	Review          string   `json:"review,omitempty"`
	Actor           string   `json:"actor"`
	SuggestedActor  string   `json:"suggestedActor"`
}

type taskDTO struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	Assignee       string     `json:"assignee,omitempty"`
	Deps           []string   `json:"deps,omitempty"`
	Ready          bool       `json:"ready"`
	UpdatedAt      string     `json:"updatedAt,omitempty"`
	Rank           float64    `json:"rank,omitempty"`
	Labels         []string   `json:"labels,omitempty"`
	Priority       string     `json:"priority,omitempty"`
	Parent         string     `json:"parent,omitempty"`
	ActiveAttempt  string     `json:"activeAttempt,omitempty"`
	ExecutionState string     `json:"executionState,omitempty"`
	SessionID      string     `json:"sessionId,omitempty"`
	Checks         []checkDTO `json:"checks,omitempty"`
	Provenance     []provDTO  `json:"provenance,omitempty"`
	Body           string     `json:"body,omitempty"`
}

type checkDTO struct {
	Desc    string `json:"desc"`
	Cmd     string `json:"cmd,omitempty"`
	Type    string `json:"type,omitempty"`
	Result  string `json:"result,omitempty"`
	Cwd     string `json:"cwd,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type provDTO struct {
	ID       string `json:"id,omitempty"`
	Who      string `json:"who"`
	At       string `json:"at"`
	Did      string `json:"did"`
	Text     string `json:"text,omitempty"`
	EditedAt string `json:"editedAt,omitempty"`
}

func dtoFromTask(t task.Task, ready bool) taskDTO {
	d := taskDTO{ID: t.ID, Title: t.Title, Status: t.Status, Assignee: t.Assignee, Deps: t.Deps,
		Ready: ready, Rank: t.Rank, Labels: t.Labels, Priority: t.Priority, Parent: t.Parent,
		ActiveAttempt: t.ActiveAttempt}
	for _, c := range t.Checks {
		d.Checks = append(d.Checks, checkDTO{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Result: c.Result, Cwd: c.Cwd, Timeout: c.Timeout})
	}
	return d
}

func dtoFromDoc(svc *mcp.Service, doc *store.Doc) taskDTO {
	d := dtoFromTask(doc.Task, svc.ReadyOf(doc.Task))
	d.ExecutionState, d.SessionID = svc.ExecutionOf(doc.Task)
	d.Body = doc.Body
	if n := len(doc.Provenance); n > 0 {
		d.UpdatedAt = doc.Provenance[n-1].At
	}
	for _, p := range doc.Provenance {
		d.Provenance = append(d.Provenance, provDTO{ID: p.ID, Who: p.Who, At: p.At, Did: p.Did, Text: p.Text, EditedAt: p.EditedAt})
	}
	return d
}

// --- responses ---

func decode(r *http.Request, v any) {
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(v)
	}
}

func errBody(err error) map[string]string { return map[string]string{"error": err.Error()} }

// writeErr maps domain errors to HTTP status codes so the UI can react (and show the
// gate reason for a refused transition).
func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, store.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, store.ErrNoteNotFound):
		code = http.StatusNotFound
	case errors.Is(err, store.ErrSessionNotFound):
		code = http.StatusNotFound
	case errors.Is(err, mcp.ErrAlreadyClaimed),
		errors.Is(err, store.ErrConflict),
		errors.Is(err, store.ErrSessionConflict),
		errors.Is(err, store.ErrLiveSession),
		errors.Is(err, session.ErrTerminal):
		code = http.StatusConflict
	case errors.Is(err, task.ErrDepsNotClosed),
		errors.Is(err, task.ErrChecksNotPassed),
		errors.Is(err, task.ErrUnknownState),
		errors.Is(err, task.ErrParentMissing),
		errors.Is(err, task.ErrParentCycle),
		errors.Is(err, task.ErrDanglingDep),
		errors.Is(err, task.ErrCycle),
		errors.Is(err, task.ErrInvalidPriority),
		errors.Is(err, task.ErrHasChildren),
		errors.Is(err, task.ErrHasDependents),
		errors.Is(err, store.ErrNotEditable),
		errors.Is(err, mcp.ErrEmptyTitle),
		errors.Is(err, mcp.ErrNotManual),
		errors.Is(err, mcp.ErrIdentityMismatch),
		errors.Is(err, mcp.ErrClientMismatch),
		errors.Is(err, mcp.ErrIdempotencyRequired),
		errors.Is(err, mcp.ErrSessionActor),
		errors.Is(err, mcp.ErrTaskClosed),
		errors.Is(err, session.ErrSummaryRequired),
		errors.Is(err, session.ErrReasonRequired):
		code = http.StatusUnprocessableEntity
	}
	writeJSON(w, code, errBody(err))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
