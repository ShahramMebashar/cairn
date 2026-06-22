package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/task"
)

const minimalTask = `---
id: PROJ-001
title: Fix the thing
status: backlog
---

Body prose that must survive byte-for-byte.
`

const fullTask = `---
id: PROJ-002
title: Add idempotency keys
status: in_progress
priority: high
deps: [PROJ-001]
checks:
  - desc: tests pass
    cmd: go test ./...
    result: pending
provenance:
  - {who: human:shah, at: 2026-06-21T10:00:00Z, did: created}
---

Full task body.
`

// repo writes config + the given task files into a temp .cairn tree and returns root.
func repo(t *testing.T, tasks map[string]string) string {
	t.Helper()
	root := t.TempDir()
	tdir := filepath.Join(root, ".cairn", "tasks")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "prefix: PROJ\ncounter: 2\nstates: [backlog, in_progress, in_review, done, canceled]\nclosed: [done, canceled]\ninitial: backlog\ncheck_timeout_default: 120\n"
	if err := os.WriteFile(filepath.Join(root, ".cairn", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	for id, body := range tasks {
		if err := os.WriteFile(filepath.Join(tdir, id+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestGetParses(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-002": fullTask}))
	d, err := s.Get("PROJ-002")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if d.Task.ID != "PROJ-002" || d.Task.Status != "in_progress" || d.Task.Title != "Add idempotency keys" {
		t.Fatalf("bad task: %+v", d.Task)
	}
	if len(d.Task.Deps) != 1 || d.Task.Deps[0] != "PROJ-001" {
		t.Fatalf("bad deps: %+v", d.Task.Deps)
	}
	if len(d.Task.Checks) != 1 || d.Task.Checks[0].Cmd != "go test ./..." {
		t.Fatalf("bad checks: %+v", d.Task.Checks)
	}
	if len(d.Provenance) != 1 || d.Provenance[0].Who != "human:shah" {
		t.Fatalf("bad provenance: %+v", d.Provenance)
	}
}

func TestListValidatesDangling(t *testing.T) {
	s := New(repo(t, map[string]string{
		"PROJ-009": "---\nid: PROJ-009\ntitle: x\nstatus: backlog\ndeps: [GHOST]\n---\n",
	}))
	if _, err := s.List(); !errors.Is(err, task.ErrDanglingDep) {
		t.Fatalf("List = %v, want dangling", err)
	}
}

func TestListValidatesCycle(t *testing.T) {
	s := New(repo(t, map[string]string{
		"A": "---\nid: A\ntitle: x\nstatus: backlog\ndeps: [B]\n---\n",
		"B": "---\nid: B\ntitle: x\nstatus: backlog\ndeps: [A]\n---\n",
	}))
	if _, err := s.List(); !errors.Is(err, task.ErrCycle) {
		t.Fatalf("List = %v, want cycle", err)
	}
}

func TestSetStatusPreservesBodyAndUnknownKeys(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-002": fullTask}))
	d, err := s.Get("PROJ-002")
	if err != nil {
		t.Fatal(err)
	}
	origBody := d.Body
	if err := d.SetStatus("done"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, _ := os.ReadFile(filepath.Join(s.tasksDir(), "PROJ-002.md"))
	got := string(raw)
	if !strings.Contains(got, "status: done") {
		t.Fatalf("status not updated:\n%s", got)
	}
	if !strings.Contains(got, "priority: high") {
		t.Fatalf("unknown key 'priority' dropped:\n%s", got)
	}
	// Body after the frontmatter is byte-for-byte preserved.
	if !strings.HasSuffix(got, origBody) {
		t.Fatalf("body changed.\norig: %q\ngot tail differs:\n%s", origBody, got)
	}

	reloaded, err := s.Get("PROJ-002")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Task.Status != "done" {
		t.Fatalf("reload status = %q", reloaded.Task.Status)
	}
}

func TestAppendProvenance(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-002": fullTask}))
	d, _ := s.Get("PROJ-002")
	at := time.Date(2026, 6, 21, 11, 0, 0, 0, time.UTC)
	if err := d.AppendProvenance("agent:claude-1", "claimed", "", at); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := s.Get("PROJ-002")
	if len(reloaded.Provenance) != 2 {
		t.Fatalf("want 2 provenance entries, got %d", len(reloaded.Provenance))
	}
	last := reloaded.Provenance[1]
	if last.Who != "agent:claude-1" || last.Did != "claimed" {
		t.Fatalf("bad appended entry: %+v", last)
	}
}

func TestSetCheckResult(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-002": fullTask}))
	d, _ := s.Get("PROJ-002")
	if err := d.SetCheckResult(0, "pass"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d); err != nil {
		t.Fatal(err)
	}
	reloaded, _ := s.Get("PROJ-002")
	if reloaded.Task.Checks[0].Result != "pass" {
		t.Fatalf("result = %q, want pass", reloaded.Task.Checks[0].Result)
	}
	if err := d.SetCheckResult(9, "pass"); err == nil {
		t.Fatal("expected out-of-range error")
	}
}

func TestCreateMintsIDAndIncrementsCounter(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	at := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	d, err := s.Create(Draft{Title: "New work", Body: "the body\n", Deps: []string{"PROJ-001"}}, "agent:claude-1", at)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.Task.ID != "PROJ-003" {
		t.Fatalf("id = %q, want PROJ-003", d.Task.ID)
	}
	if d.Task.Status != "backlog" {
		t.Fatalf("status = %q, want initial backlog", d.Task.Status)
	}
	if len(d.Provenance) != 1 || d.Provenance[0].Did != "created" {
		t.Fatalf("missing created provenance: %+v", d.Provenance)
	}
	// File exists and counter persisted.
	if _, err := os.Stat(filepath.Join(s.tasksDir(), "PROJ-003.md")); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	again, err := s.Create(Draft{Title: "More"}, "agent:claude-1", at)
	if err != nil {
		t.Fatal(err)
	}
	if again.Task.ID != "PROJ-004" {
		t.Fatalf("second id = %q, want PROJ-004 (counter not persisted)", again.Task.ID)
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-002": fullTask}))
	d, _ := s.Get("PROJ-002")
	_ = d.SetStatus("done")
	if err := s.Save(d); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(s.tasksDir())
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestSaveConflictsOnStaleDoc(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))

	d1, err := s.Get("PROJ-001")
	if err != nil {
		t.Fatal(err)
	}
	d2, err := s.Get("PROJ-001")
	if err != nil {
		t.Fatal(err)
	}

	// d2 writes first; the file on disk changes underneath d1.
	if err := d2.AppendProvenance("agent:a", "note", "first", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d2); err != nil {
		t.Fatalf("first save: %v", err)
	}

	// d1 is now stale: saving it must conflict, not silently clobber d2's write.
	if err := d1.AppendProvenance("agent:b", "note", "second", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d1); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale save = %v, want ErrConflict", err)
	}
}

func TestSaveSucceedsAfterReread(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	d, _ := s.Get("PROJ-001")
	if err := d.AppendProvenance("agent:a", "note", "x", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d); err != nil {
		t.Fatalf("save: %v", err)
	}
	// A fresh read reflects the write and can save again (version moved forward).
	d2, _ := s.Get("PROJ-001")
	if err := d2.AppendProvenance("agent:a", "note", "y", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(d2); err != nil {
		t.Fatalf("second save after reread: %v", err)
	}
}

func TestCreateReadsAndClearsOrgFields(t *testing.T) {
	s := New(repo(t, map[string]string{}))
	at := time.Date(2026, 6, 22, 9, 0, 0, 0, time.UTC)
	d, err := s.Create(Draft{Title: "x", Labels: []string{"backend", "db"}, Priority: "high", Parent: "PROJ-001"}, "a", at)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(d.Task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Task.Priority != "high" || got.Task.Parent != "PROJ-001" || len(got.Task.Labels) != 2 {
		t.Fatalf("org fields not round-tripped: %+v", got.Task)
	}
	// clearing removes the keys
	_ = got.SetPriority("")
	_ = got.SetLabels(nil)
	_ = got.SetParent("")
	if err := s.Save(got); err != nil {
		t.Fatal(err)
	}
	again, _ := s.Get(d.Task.ID)
	if again.Task.Priority != "" || again.Task.Parent != "" || len(again.Task.Labels) != 0 {
		t.Fatalf("org fields not cleared: %+v", again.Task)
	}
}

func TestRankRoundTripAndClear(t *testing.T) {
	s := New(repo(t, map[string]string{}))
	at := time.Date(2026, 6, 22, 9, 0, 0, 0, time.UTC)
	d, err := s.Create(Draft{Title: "x", Rank: 1500.5}, "a", at)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(d.Task.ID)
	if got.Task.Rank != 1500.5 {
		t.Fatalf("rank = %v, want 1500.5", got.Task.Rank)
	}
	_ = got.SetRank(0)
	if err := s.Save(got); err != nil {
		t.Fatal(err)
	}
	again, _ := s.Get(d.Task.ID)
	if again.Task.Rank != 0 {
		t.Fatalf("rank not cleared: %v", again.Task.Rank)
	}
}
