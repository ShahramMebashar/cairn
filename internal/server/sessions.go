package server

import (
	"net/http"

	"cairn/internal/mcp"
	"cairn/internal/repo"
	"cairn/internal/session"
)

type beginSessionReq struct {
	ExpectedActor  string `json:"expectedActor"`
	Client         string `json:"client"`
	Model          string `json:"model"`
	Worktree       string `json:"worktree"`
	Branch         string `json:"branch"`
	Head           string `json:"head"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type usageReq struct {
	InputTokens  int64 `json:"inputTokens"`
	OutputTokens int64 `json:"outputTokens"`
	CachedTokens int64 `json:"cachedTokens"`
	ToolCalls    int64 `json:"toolCalls"`
}

func (u usageReq) sessionUsage() session.Usage {
	return session.Usage{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens, CachedTokens: u.CachedTokens, ToolCalls: u.ToolCalls}
}

type heartbeatReq struct {
	Progress string   `json:"progress"`
	Usage    usageReq `json:"usage"`
}

type finishSessionReq struct {
	Summary string   `json:"summary"`
	Head    string   `json:"head"`
	Usage   usageReq `json:"usage"`
}

type cancelSessionReq struct {
	Reason string `json:"reason"`
}

func (s *Server) handleIdentity(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, svc.Identity())
}

func (s *Server) handleBeginSession(w http.ResponseWriter, r *http.Request) {
	svc, root, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	if err := repo.EnsureSessionDirs(root); err != nil {
		writeErr(w, err)
		return
	}
	var req beginSessionReq
	decode(r, &req)
	view, err := svc.BeginSession(r.Context(), mcp.BeginSessionInput{
		TaskID: r.PathValue("id"), ExpectedActor: req.ExpectedActor, Client: req.Client,
		Model: req.Model, Worktree: req.Worktree, Branch: req.Branch, Head: req.Head,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleListTaskSessions(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	views, err := svc.ListSessions(r.PathValue("id"), "", "", "")
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": views})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	views, err := svc.ListSessions(q.Get("task"), q.Get("actor"), q.Get("status"), q.Get("health"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": views})
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	view, err := svc.GetSession(r.PathValue("session"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req heartbeatReq
	decode(r, &req)
	view, err := svc.Heartbeat(r.Context(), mcp.HeartbeatInput{
		SessionID: r.PathValue("session"), Progress: req.Progress, Usage: req.Usage.sessionUsage(),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleFinishSession(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req finishSessionReq
	decode(r, &req)
	view, err := svc.FinishSession(r.Context(), mcp.FinishSessionInput{
		SessionID: r.PathValue("session"), Summary: req.Summary, Head: req.Head, Usage: req.Usage.sessionUsage(),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleCancelSession(w http.ResponseWriter, r *http.Request) {
	svc, _, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	var req cancelSessionReq
	decode(r, &req)
	view, err := svc.CancelSession(r.Context(), mcp.CancelSessionInput{
		SessionID: r.PathValue("session"), Reason: req.Reason,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}
