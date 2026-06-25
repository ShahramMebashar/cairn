package connect

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// AgentStatus is the per-agent view returned to the UI.
type AgentStatus struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Mode       Mode   `json:"mode"`
	Installed  bool   `json:"installed"`
	Connected  bool   `json:"connected"`
	TargetPath string `json:"targetPath,omitempty"`
	DocsURL    string `json:"docsURL,omitempty"`
}

// Guide is a manual-setup snippet for one agent: the exact file content to write and the
// path to write it to.
type Guide struct {
	Path   string `json:"path,omitempty"`
	Lang   string `json:"lang"`
	Config string `json:"config"`
}

// Detect returns the status of every known agent for repo — whether it looks installed and
// whether its config already points at cairn. Installed agents sort first.
func Detect(repo string) ([]AgentStatus, error) {
	s, err := defaultSys()
	if err != nil {
		return nil, err
	}
	return detectWith(s, repo), nil
}

func detectWith(s sys, repo string) []AgentStatus {
	reg := registry()
	out := make([]AgentStatus, 0, len(reg))
	for _, a := range reg {
		st := AgentStatus{ID: a.id, Name: a.name, Mode: a.mode, DocsURL: a.docsURL}
		if a.detect != nil {
			st.Installed = a.detect(s)
		}
		if a.target != nil {
			st.TargetPath = a.target(s, repo)
			if a.format != nil {
				if b, err := os.ReadFile(st.TargetPath); err == nil {
					st.Connected = a.format.connected(b)
				}
			}
		}
		out = append(out, st)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Installed && !out[j].Installed
	})
	return out
}

// Connect merges the cairn MCP entry into agentID's config for repo, stamping actor. It
// writes atomically, backs up any existing file to <path>.bak, then verifies the entry is
// present by re-reading. Returns the path written.
func Connect(agentID, repo, actor string) (string, error) {
	s, err := defaultSys()
	if err != nil {
		return "", err
	}
	bin, err := selfPath()
	if err != nil {
		return "", err
	}
	return connectWith(s, bin, agentID, repo, actor)
}

func connectWith(s sys, bin, agentID, repo, actor string) (string, error) {
	a, ok := find(agentID)
	if !ok {
		return "", fmt.Errorf("unknown agent %q", agentID)
	}
	if a.mode != ModeAuto || a.format == nil || a.target == nil {
		return "", fmt.Errorf("agent %q has no auto-connect; use the manual guide", agentID)
	}
	cfg := newServerConfig(bin, repo, resolveActor(agentID, actor))
	path := a.target(s, repo)
	existing, _ := os.ReadFile(path) // empty if absent
	next, err := a.format.upsert(existing, cfg)
	if err != nil {
		return "", err
	}
	if err := writeFileAtomic(path, existing, next); err != nil {
		return "", err
	}
	if got, err := os.ReadFile(path); err != nil || !a.format.connected(got) {
		return "", fmt.Errorf("wrote %s but could not verify the cairn entry", path)
	}
	return path, nil
}

// Disconnect removes only cairn's entry from agentID's config for repo, leaving the file and
// any other MCP servers intact (backing up to <path>.bak first). It's a no-op if the file is
// absent or already has no cairn entry. Returns the config path.
func Disconnect(agentID, repo string) (string, error) {
	s, err := defaultSys()
	if err != nil {
		return "", err
	}
	a, ok := find(agentID)
	if !ok {
		return "", fmt.Errorf("unknown agent %q", agentID)
	}
	if a.format == nil || a.target == nil {
		return "", fmt.Errorf("agent %q has no config to disconnect", agentID)
	}
	path := a.target(s, repo)
	existing, err := os.ReadFile(path)
	if err != nil || !a.format.connected(existing) {
		return path, nil // nothing to remove
	}
	next, err := a.format.remove(existing)
	if err != nil {
		return "", err
	}
	if err := writeFileAtomic(path, existing, next); err != nil {
		return "", err
	}
	if got, err := os.ReadFile(path); err != nil || a.format.connected(got) {
		return "", fmt.Errorf("removed cairn from %s but it's still present", path)
	}
	return path, nil
}

// ManualGuide renders the standalone config text auto-connect would write for agentID, for
// display when the user wants to wire it up by hand.
func ManualGuide(agentID, repo, actor string) (Guide, error) {
	s, err := defaultSys()
	if err != nil {
		return Guide{}, err
	}
	a, ok := find(agentID)
	if !ok || a.format == nil {
		return Guide{}, fmt.Errorf("no guide for agent %q", agentID)
	}
	bin, err := selfPath()
	if err != nil {
		bin = serverName // fall back to bare name in the snippet
	}
	b, err := a.format.upsert(nil, newServerConfig(bin, repo, resolveActor(agentID, actor)))
	if err != nil {
		return Guide{}, err
	}
	g := Guide{Lang: a.format.lang(), Config: string(b)}
	if a.target != nil {
		g.Path = a.target(s, repo)
	}
	return g, nil
}

func newServerConfig(bin, repo, actor string) serverConfig {
	return serverConfig{Name: serverName, Bin: bin, Args: []string{"serve", "--actor", actor, "--repo", repo}}
}

// resolveActor gives each agent its own identity by default (e.g. agent:cursor) so its task
// writes are attributed to it in provenance — never to the human operator. A caller-supplied
// actor (to run multiple instances, e.g. agent:cursor-2) overrides it.
func resolveActor(agentID, actor string) string {
	if actor == "" {
		return "agent:" + agentID
	}
	return actor
}

// selfPath returns the absolute, symlink-resolved path of the running cairn binary. In the
// desktop app this is the Tauri sidecar path, which is exactly what agents must launch.
func selfPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		return resolved, nil
	}
	return exe, nil
}

// writeFileAtomic writes data to path via a temp file + rename, creating parent dirs and
// backing up any existing file to <path>.bak first.
func writeFileAtomic(path string, existing, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if len(existing) > 0 {
		if err := os.WriteFile(path+".bak", existing, 0o644); err != nil {
			return err
		}
	}
	tmp, err := os.CreateTemp(dir, ".cairn-connect-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
