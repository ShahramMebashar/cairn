package connect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestMcpServersJSONUpsertPreservesAndIsIdempotent(t *testing.T) {
	existing := []byte(`{
  "mcpServers": { "other": { "command": "x" } },
  "extra": true
}`)
	cfg := serverConfig{Name: "cairn", Bin: "/abs/cairn", Args: []string{"serve", "--repo", "/r"}}

	out, err := (mcpServersJSON{}).upsert(existing, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	if root["extra"] != true {
		t.Error("top-level key not preserved")
	}
	servers := root["mcpServers"].(map[string]any)
	if _, ok := servers["other"]; !ok {
		t.Error("sibling server not preserved")
	}
	cairn := servers["cairn"].(map[string]any)
	if cairn["command"] != "/abs/cairn" {
		t.Errorf("command = %v", cairn["command"])
	}
	if !(mcpServersJSON{}).connected(out) {
		t.Error("connected should be true after upsert")
	}

	// Idempotent: a second upsert yields identical bytes.
	out2, err := (mcpServersJSON{}).upsert(out, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(out2) {
		t.Error("upsert not idempotent")
	}
}

func TestMcpServersJSONUpsertEmpty(t *testing.T) {
	out, err := (mcpServersJSON{}).upsert(nil, serverConfig{Name: "cairn", Bin: "/c", Args: []string{"serve"}})
	if err != nil {
		t.Fatal(err)
	}
	if !(mcpServersJSON{}).connected(out) {
		t.Errorf("fresh file should be connected:\n%s", out)
	}
}

func TestOpenCodeJSONUpsert(t *testing.T) {
	cfg := serverConfig{Name: "cairn", Bin: "/abs/cairn", Args: []string{"serve", "--repo", "/r"}}
	out, err := (openCodeJSON{}).upsert([]byte(`{"mcp":{"keep":{"type":"local"}}}`), cfg)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	if root["$schema"] != "https://opencode.ai/config.json" {
		t.Errorf("schema = %v", root["$schema"])
	}
	mcp := root["mcp"].(map[string]any)
	if _, ok := mcp["keep"]; !ok {
		t.Error("sibling not preserved")
	}
	cairn := mcp["cairn"].(map[string]any)
	if cairn["type"] != "local" || cairn["enabled"] != true {
		t.Errorf("cairn entry = %v", cairn)
	}
	cmd := cairn["command"].([]any)
	if len(cmd) != 4 || cmd[0] != "/abs/cairn" {
		t.Errorf("command argv = %v", cmd)
	}
	if !(openCodeJSON{}).connected(out) {
		t.Error("connected should be true")
	}
}

func TestCodexTOMLUpsertPreserves(t *testing.T) {
	existing := []byte("model = \"o1\"\n\n[mcp_servers.other]\ncommand = \"x\"\n")
	cfg := serverConfig{Name: "cairn", Bin: "/abs/cairn", Args: []string{"serve", "--repo", "/r"}}
	out, err := (codexTOML{}).upsert(existing, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := toml.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	if root["model"] != "o1" {
		t.Error("top-level key not preserved")
	}
	servers := root["mcp_servers"].(map[string]any)
	if _, ok := servers["other"]; !ok {
		t.Error("sibling table not preserved")
	}
	cairn := servers["cairn"].(map[string]any)
	if cairn["command"] != "/abs/cairn" {
		t.Errorf("command = %v", cairn["command"])
	}
	if !(codexTOML{}).connected(out) {
		t.Error("connected should be true")
	}
}

// stubSys builds a sys rooted at a temp HOME with the given binaries "on PATH".
func stubSys(home string, onPath ...string) sys {
	set := map[string]bool{}
	for _, b := range onPath {
		set[b] = true
	}
	return sys{home: home, lookPath: func(b string) (string, error) {
		if set[b] {
			return "/usr/bin/" + b, nil
		}
		return "", os.ErrNotExist
	}}
}

func TestConnectWritesBackupAndVerifies(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	s := stubSys(home)

	// First connect: creates .cursor/mcp.json, no backup (nothing pre-existed).
	path, err := connectWith(s, "/abs/cairn", "cursor", repo, "agent:test")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(repo, ".cursor", "mcp.json"); path != want {
		t.Errorf("path = %s want %s", path, want)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Error("backup should not exist on first write")
	}
	b, _ := os.ReadFile(path)
	if !(mcpServersJSON{}).connected(b) {
		t.Errorf("config not connected:\n%s", b)
	}
	var root map[string]any
	json.Unmarshal(b, &root)
	args := root["mcpServers"].(map[string]any)["cairn"].(map[string]any)["args"].([]any)
	if args[len(args)-1] != repo {
		t.Errorf("--repo arg = %v want %s", args[len(args)-1], repo)
	}

	// Second connect: overwrites, so a .bak of the prior content is kept.
	if _, err := connectWith(s, "/abs/cairn", "cursor", repo, "agent:test2"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Errorf("backup missing after overwrite: %v", err)
	}
}

func TestConnectAutoAgentsWriteExpectedPaths(t *testing.T) {
	cases := map[string]string{
		"kilo": ".kilocode/mcp.json", // project-scoped, standard mcpServers
		"pi":   ".mcp.json",          // Pi's preferred project config (shared with Claude)
	}
	for agent, rel := range cases {
		repo := t.TempDir()
		path, err := connectWith(stubSys(t.TempDir()), "/abs/cairn", agent, repo, "agent:test")
		if err != nil {
			t.Fatalf("%s: %v", agent, err)
		}
		if want := filepath.Join(repo, rel); path != want {
			t.Errorf("%s path = %s want %s", agent, path, want)
		}
		b, _ := os.ReadFile(path)
		if !(mcpServersJSON{}).connected(b) {
			t.Errorf("%s config not connected:\n%s", agent, b)
		}
	}
}

func TestConnectDefaultsToPerAgentIdentity(t *testing.T) {
	repo := t.TempDir()
	// Empty actor → the agent gets its own identity (agent:cursor), not a human/default name.
	path, err := connectWith(stubSys(t.TempDir()), "/abs/cairn", "cursor", repo, "")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	var root map[string]any
	json.Unmarshal(b, &root)
	args := root["mcpServers"].(map[string]any)["cairn"].(map[string]any)["args"].([]any)
	// args = [serve --actor agent:cursor --repo <repo>]
	if args[2] != "agent:cursor" {
		t.Errorf("default actor = %v, want agent:cursor", args[2])
	}
}

func TestDisconnectRemovesOnlyCairnAndKeepsSiblings(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// A config with cairn plus another server.
	seed := []byte(`{"mcpServers":{"cairn":{"command":"/c"},"other":{"command":"x"}}}`)
	if err := os.WriteFile(path, seed, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Disconnect("cursor", repo)
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Errorf("path = %s want %s", got, path)
	}
	b, _ := os.ReadFile(path)
	if (mcpServersJSON{}).connected(b) {
		t.Errorf("cairn entry should be gone:\n%s", b)
	}
	var root map[string]any
	json.Unmarshal(b, &root)
	if _, ok := root["mcpServers"].(map[string]any)["other"]; !ok {
		t.Errorf("sibling server should be preserved:\n%s", b)
	}
	// Idempotent: disconnecting again is a no-op (not an error).
	if _, err := Disconnect("cursor", repo); err != nil {
		t.Errorf("second disconnect errored: %v", err)
	}
}

func TestDisconnectMissingFileIsNoop(t *testing.T) {
	if _, err := Disconnect("cursor", t.TempDir()); err != nil {
		t.Errorf("disconnect with no config should be a no-op, got %v", err)
	}
}

func TestConnectRejectsManualAgent(t *testing.T) {
	if _, err := connectWith(stubSys(t.TempDir()), "/c", "antigravity", t.TempDir(), ""); err == nil {
		t.Error("expected error connecting a manual-only agent")
	}
}

func TestDetectMarksInstalledAndSortsFirst(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	// Cursor "installed" via ~/.cursor; others not.
	if err := os.MkdirAll(filepath.Join(home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	out := detectWith(stubSys(home), repo)
	if len(out) == 0 {
		t.Fatal("no agents")
	}
	if !out[0].Installed {
		t.Errorf("first agent should be installed, got %+v", out[0])
	}
	byID := map[string]AgentStatus{}
	for _, a := range out {
		byID[a.ID] = a
	}
	if !byID["cursor"].Installed {
		t.Error("cursor should be detected as installed")
	}
	if byID["claude"].Installed {
		t.Error("claude should not be installed in this stub")
	}
}

func TestManualGuideRendersSnippet(t *testing.T) {
	g, err := ManualGuide("codex", "/my/repo", "agent:x")
	if err != nil {
		t.Fatal(err)
	}
	if g.Lang != "toml" {
		t.Errorf("lang = %s", g.Lang)
	}
	if !(codexTOML{}).connected([]byte(g.Config)) {
		t.Errorf("guide snippet missing cairn entry:\n%s", g.Config)
	}
}
