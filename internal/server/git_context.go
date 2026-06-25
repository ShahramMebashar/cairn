package server

import (
	"net/http"
	"sync"

	"cairn/internal/gitctx"
	"cairn/internal/mcp"
	"cairn/internal/session"
)

type sessionGitContextDTO struct {
	Session mcp.SessionView `json:"session"`
	Context gitctx.Context  `json:"context"`
}

func (s *Server) handleTaskGitContext(w http.ResponseWriter, r *http.Request) {
	svc, root, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	taskID := r.PathValue("id")
	views, err := svc.ListSessions(taskID, "", "", "")
	if err != nil {
		writeErr(w, err)
		return
	}
	latestHead := latestRunHead(root, taskID)
	// Each session triggers several git subprocesses; fan out so a task with many
	// sessions does not serialize their (timeout-bounded) latencies.
	out := make([]sessionGitContextDTO, len(views))
	var wg sync.WaitGroup
	for i, view := range views {
		wg.Add(1)
		go func(i int, view mcp.SessionView) {
			defer wg.Done()
			repo := root
			if view.Live != nil && view.Live.Worktree != "" {
				repo = view.Live.Worktree
			}
			out[i] = sessionGitContextDTO{
				Session: view,
				Context: gitctx.Session(
					r.Context(),
					repo,
					view.HeadStarted,
					view.HeadFinished,
					view.Branch,
					latestHead,
					view.Status == session.StatusActive,
				),
			}
		}(i, view)
	}
	wg.Wait()
	writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

func (s *Server) handleSessionGitContext(w http.ResponseWriter, r *http.Request) {
	svc, root, ok := s.svcFor(w, r)
	if !ok {
		return
	}
	view, err := svc.GetSession(r.PathValue("session"))
	if err != nil {
		writeErr(w, err)
		return
	}
	repo := root
	if view.Live != nil && view.Live.Worktree != "" {
		repo = view.Live.Worktree
	}
	ctx := gitctx.Session(
		r.Context(),
		repo,
		view.HeadStarted,
		view.HeadFinished,
		view.Branch,
		latestRunHead(root, view.TaskID),
		view.Status == session.StatusActive,
	)
	writeJSON(w, http.StatusOK, map[string]any{"session": view, "context": ctx})
}
