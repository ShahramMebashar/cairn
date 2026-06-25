package server

import (
	"net/http"

	"cairn/internal/connect"
)

// Integration endpoints let the UI wire AI agents to cairn with one click: the server
// detects installed agents and writes their MCP config files itself (it already runs
// locally with the user's permissions), so it works in both the desktop app and a browser.
//
//	GET    /api/connect              ?path           -> { agents: [AgentStatus] }
//	POST   /api/connect/{agent}      { path?, actor? } -> { connected, path }
//	DELETE /api/connect/{agent}      ?path           -> { connected:false, path }
//	GET    /api/connect/{agent}/manual ?path&actor    -> { path, lang, config }

func (s *Server) handleListIntegrations(w http.ResponseWriter, r *http.Request) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	agents, err := connect.Detect(root)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"agents": agents})
}

type connectReq struct {
	Path  string `json:"path"`
	Actor string `json:"actor"`
}

func (s *Server) handleConnectAgent(w http.ResponseWriter, r *http.Request) {
	var req connectReq
	decode(r, &req)
	root, err := s.resolveRoot(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	// Empty actor ⇒ connect package defaults to agent:<id>. We deliberately do NOT fall
	// back to the human request actor: an agent's writes must be stamped as that agent.
	path, err := connect.Connect(r.PathValue("agent"), root, sanitizeActor(req.Actor))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"connected": true, "path": path})
}

func (s *Server) handleDisconnectAgent(w http.ResponseWriter, r *http.Request) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	path, err := connect.Disconnect(r.PathValue("agent"), root)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"connected": false, "path": path})
}

func (s *Server) handleAgentManual(w http.ResponseWriter, r *http.Request) {
	root, err := s.resolveRoot(r.URL.Query().Get("path"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	guide, err := connect.ManualGuide(r.PathValue("agent"), root, sanitizeActor(r.URL.Query().Get("actor")))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, guide)
}
