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
//	GET  /api/tasks/{id}
//	GET  /api/tasks/{id}/runs              -> { runs: [{ file, at, cmd, cwd, exit, timedout, duration, output }] }
//	POST /api/tasks/{id}/transition  { to }
//	POST /api/tasks/{id}/claim
//	POST /api/tasks/{id}/run_checks  { only? }
//	POST /api/tasks/{id}/attest      { index, pass? }   -> attest a manual check (pass defaults true)
//	POST /api/tasks/{id}/note        { text }
//	GET  /api/events      ?path                  -> text/event-stream of { type, id? } change signals
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cairn/internal/mcp"
	"cairn/internal/repo"
	"cairn/internal/store"
	"cairn/internal/task"
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
	mux.HandleFunc("POST /api/init", s.handleInit)
	mux.HandleFunc("GET /api/tasks", s.handleList)
	mux.HandleFunc("POST /api/tasks", s.handleCreate)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGet)
	mux.HandleFunc("GET /api/tasks/{id}/runs", s.handleRuns)
	mux.HandleFunc("GET /api/events", s.handleEvents)
	mux.HandleFunc("POST /api/tasks/{id}/update", s.handleUpdate)
	mux.HandleFunc("POST /api/tasks/{id}/reorder", s.handleReorder)
	mux.HandleFunc("POST /api/tasks/{id}/transition", s.handleTransition)
	mux.HandleFunc("POST /api/tasks/{id}/claim", s.handleClaim)
	mux.HandleFunc("POST /api/tasks/{id}/run_checks", s.handleRunChecks)
	mux.HandleFunc("POST /api/tasks/{id}/attest", s.handleAttest)
	mux.HandleFunc("POST /api/tasks/{id}/note", s.handleNote)
	return mux
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

func (s *Server) service(root string) *mcp.Service {
	return mcp.NewService(store.New(root), s.actor, nil)
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
	views, err := s.service(root).List(q.Get("status"), q.Get("assignee"), ready)
	if err != nil {
		writeErr(w, err)
		return
	}
	tasks := make([]taskDTO, 0, len(views))
	for _, v := range views {
		dto := dtoFromTask(v.Task, v.Ready)
		dto.UpdatedAt = v.UpdatedAt
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
	Priority *string   `json:"priority"`
	Labels   *[]string `json:"labels"`
	Parent   *string   `json:"parent"`
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
	doc, err := svc.Update(r.PathValue("id"), mcp.UpdateFields{Priority: req.Priority, Labels: req.Labels, Parent: req.Parent})
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
	return s.service(root), root, true
}

func (s *Server) status(root string) statusResp {
	resp := statusResp{
		Initialized:     repo.IsInitialized(root),
		Root:            root,
		SuggestedPrefix: repo.DerivePrefix(root),
		Actor:           s.actor,
	}
	if resp.Initialized {
		if cfg, err := store.New(root).Config(); err == nil {
			resp.Prefix = cfg.Prefix
			resp.States = cfg.States
			resp.Closed = cfg.Closed
			resp.Initial = cfg.Initial
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
	Actor           string   `json:"actor"`
}

type taskDTO struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	Assignee   string     `json:"assignee,omitempty"`
	Deps       []string   `json:"deps,omitempty"`
	Ready      bool       `json:"ready"`
	UpdatedAt  string     `json:"updatedAt,omitempty"`
	Rank       float64    `json:"rank,omitempty"`
	Labels     []string   `json:"labels,omitempty"`
	Priority   string     `json:"priority,omitempty"`
	Parent     string     `json:"parent,omitempty"`
	Checks     []checkDTO `json:"checks,omitempty"`
	Provenance []provDTO  `json:"provenance,omitempty"`
	Body       string     `json:"body,omitempty"`
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
	Who  string `json:"who"`
	At   string `json:"at"`
	Did  string `json:"did"`
	Text string `json:"text,omitempty"`
}

func dtoFromTask(t task.Task, ready bool) taskDTO {
	d := taskDTO{ID: t.ID, Title: t.Title, Status: t.Status, Assignee: t.Assignee, Deps: t.Deps,
		Ready: ready, Rank: t.Rank, Labels: t.Labels, Priority: t.Priority, Parent: t.Parent}
	for _, c := range t.Checks {
		d.Checks = append(d.Checks, checkDTO{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Result: c.Result, Cwd: c.Cwd, Timeout: c.Timeout})
	}
	return d
}

func dtoFromDoc(svc *mcp.Service, doc *store.Doc) taskDTO {
	d := dtoFromTask(doc.Task, svc.ReadyOf(doc.Task))
	d.Body = doc.Body
	if n := len(doc.Provenance); n > 0 {
		d.UpdatedAt = doc.Provenance[n-1].At
	}
	for _, p := range doc.Provenance {
		d.Provenance = append(d.Provenance, provDTO{Who: p.Who, At: p.At, Did: p.Did, Text: p.Text})
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
	case errors.Is(err, mcp.ErrAlreadyClaimed),
		errors.Is(err, store.ErrConflict):
		code = http.StatusConflict
	case errors.Is(err, task.ErrDepsNotClosed),
		errors.Is(err, task.ErrChecksNotPassed),
		errors.Is(err, task.ErrUnknownState),
		errors.Is(err, task.ErrParentMissing),
		errors.Is(err, task.ErrParentCycle),
		errors.Is(err, task.ErrDanglingDep),
		errors.Is(err, task.ErrCycle),
		errors.Is(err, task.ErrInvalidPriority),
		errors.Is(err, mcp.ErrNotManual):
		code = http.StatusUnprocessableEntity
	}
	writeJSON(w, code, errBody(err))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
