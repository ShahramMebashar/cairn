package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/store"
)

// TestServerEndToEnd drives the full MCP path: a real client calls the registered tools
// over an in-memory transport, exercising schema marshaling and the create -> claim ->
// run_checks -> transition lifecycle (SPEC §11.7 dogfood, in miniature).
func TestServerEndToEnd(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cairn", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "prefix: PROJ\ncounter: 0\nstates: [backlog, in_progress, done, canceled]\nclosed: [done, canceled]\ninitial: backlog\ncheck_timeout_default: 30\n"
	if err := os.WriteFile(filepath.Join(root, ".cairn", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	at := time.Date(2026, 6, 21, 9, 0, 0, 0, time.UTC)
	svc := NewService(store.New(root), "agent:claude-1", func() time.Time { return at })
	srv := NewServer(svc)

	ctx := context.Background()
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer ss.Close()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	call := func(name string, args map[string]any) taskOut {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatalf("%s call: %v", name, err)
		}
		if res.IsError {
			t.Fatalf("%s returned tool error: %+v", name, res.Content)
		}
		var out taskOut
		b, _ := json.Marshal(res.StructuredContent)
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("%s decode: %v", name, err)
		}
		return out
	}

	created := call("create", map[string]any{
		"title":  "ship it",
		"checks": []map[string]any{{"desc": "tests", "cmd": "exit 0"}},
	})
	if created.ID != "PROJ-001" {
		t.Fatalf("create id = %q", created.ID)
	}

	claimed := call("claim", map[string]any{"id": "PROJ-001"})
	if claimed.Assignee != "agent:claude-1" {
		t.Fatalf("assignee = %q", claimed.Assignee)
	}

	done := call("transition", map[string]any{"id": "PROJ-001", "to": "done"})
	if done.Status != "done" {
		t.Fatalf("status = %q, want done", done.Status)
	}
	if done.Checks[0].Result != "pass" {
		t.Fatalf("check result = %q, want pass", done.Checks[0].Result)
	}
}
