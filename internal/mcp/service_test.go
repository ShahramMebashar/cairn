package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/store"
	"cairn/internal/task"
)

func service(t *testing.T, actor string) *Service {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cairn", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "prefix: PROJ\ncounter: 0\nstates: [backlog, in_progress, in_review, done, canceled]\nclosed: [done, canceled]\ninitial: backlog\ncheck_timeout_default: 30\n"
	if err := os.WriteFile(filepath.Join(root, ".cairn", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 6, 21, 9, 0, 0, 0, time.UTC)
	return NewService(store.New(root), actor, func() time.Time { return at })
}

func TestCreateAndGet(t *testing.T) {
	svc := service(t, "agent:claude-1")
	d, err := svc.Create(store.Draft{Title: "first", Body: "body\n"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.Task.ID != "PROJ-001" || d.Task.Status != "backlog" {
		t.Fatalf("bad created task: %+v", d.Task)
	}
	got, err := svc.Get("PROJ-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Provenance[0].Who != "agent:claude-1" || got.Provenance[0].Did != "created" {
		t.Fatalf("bad provenance: %+v", got.Provenance)
	}
}

func TestCreateRejectsMissingDep(t *testing.T) {
	svc := service(t, "agent:a")
	if _, err := svc.Create(store.Draft{Title: "x", Deps: []string{"GHOST"}}); err == nil {
		t.Fatal("expected error creating task with missing dep")
	}
}

func TestClaim(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x"})

	if _, err := svc.Claim(d.Task.ID); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	// Same actor re-claiming is a no-op/ok.
	if _, err := svc.Claim(d.Task.ID); err != nil {
		t.Fatalf("re-claim by same actor should be ok: %v", err)
	}
	// Different actor is refused.
	other := NewService(svc.store, "agent:b", svc.now)
	if _, err := other.Claim(d.Task.ID); !errors.Is(err, ErrAlreadyClaimed) {
		t.Fatalf("other claim = %v, want ErrAlreadyClaimed", err)
	}
}

func TestTransitionDepsGate(t *testing.T) {
	svc := service(t, "agent:a")
	dep, _ := svc.Create(store.Draft{Title: "dep"})
	blocked, _ := svc.Create(store.Draft{Title: "blocked", Deps: []string{dep.Task.ID}})

	if _, err := svc.Transition(blocked.Task.ID, "in_progress"); !errors.Is(err, task.ErrDepsNotClosed) {
		t.Fatalf("transition = %v, want deps gate", err)
	}
	// Close the dep, then the blocked task may start.
	if _, err := svc.Transition(dep.Task.ID, "done"); err != nil {
		t.Fatalf("close dep: %v", err)
	}
	if _, err := svc.Transition(blocked.Task.ID, "in_progress"); err != nil {
		t.Fatalf("transition after dep closed: %v", err)
	}
}

func TestTransitionAutoRunsChecksAndCloses(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x", Checks: []task.Check{{Desc: "ok", Cmd: "exit 0"}}})

	got, err := svc.Transition(d.Task.ID, "done")
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if got.Task.Status != "done" {
		t.Fatalf("status = %q, want done", got.Task.Status)
	}
	if got.Task.Checks[0].Result != "pass" {
		t.Fatalf("check result = %q, want pass (auto-run)", got.Task.Checks[0].Result)
	}
}

func TestTransitionRefusesOnFailingCheck(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x", Checks: []task.Check{{Desc: "bad", Cmd: "exit 1"}}})

	if _, err := svc.Transition(d.Task.ID, "done"); !errors.Is(err, task.ErrChecksNotPassed) {
		t.Fatalf("transition = %v, want checks gate refusal", err)
	}
	got, _ := svc.Get(d.Task.ID)
	if got.Task.Status == "done" {
		t.Fatal("task closed despite failing check")
	}
	if got.Task.Checks[0].Result != "fail" {
		t.Fatalf("check result = %q, want fail (recorded)", got.Task.Checks[0].Result)
	}
}

func TestTransitionRefusesOnPendingManualCheck(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x", Checks: []task.Check{{Desc: "review", Type: "manual"}}})

	if _, err := svc.Transition(d.Task.ID, "done"); !errors.Is(err, task.ErrChecksNotPassed) {
		t.Fatalf("transition = %v, want refusal on pending manual check", err)
	}
}

func TestRunChecks(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x", Checks: []task.Check{{Desc: "a", Cmd: "exit 0"}, {Desc: "b", Cmd: "exit 1"}}})

	got, err := svc.RunChecks(d.Task.ID, nil)
	if err != nil {
		t.Fatalf("RunChecks: %v", err)
	}
	if got.Task.Checks[0].Result != "pass" || got.Task.Checks[1].Result != "fail" {
		t.Fatalf("results = %q,%q want pass,fail", got.Task.Checks[0].Result, got.Task.Checks[1].Result)
	}
}

func TestNoteAppendsProvenance(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x"})
	if _, err := svc.Note(d.Task.ID, "looked into it"); err != nil {
		t.Fatalf("Note: %v", err)
	}
	got, _ := svc.Get(d.Task.ID)
	last := got.Provenance[len(got.Provenance)-1]
	if last.Did != "note" || last.Text != "looked into it" {
		t.Fatalf("bad note provenance: %+v", last)
	}
}

func TestListFilters(t *testing.T) {
	svc := service(t, "agent:a")
	dep, _ := svc.Create(store.Draft{Title: "dep"})
	svc.Create(store.Draft{Title: "blocked", Deps: []string{dep.Task.ID}})
	free, _ := svc.Create(store.Draft{Title: "free"})

	// ready=true should include the depless tasks but exclude the blocked one.
	ready := true
	views, err := svc.List("", "", &ready)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	ids := map[string]bool{}
	for _, v := range views {
		if !v.Ready {
			t.Fatalf("ready filter returned not-ready task %s", v.ID)
		}
		ids[v.ID] = true
	}
	if !ids[dep.Task.ID] || !ids[free.Task.ID] {
		t.Fatalf("ready set missing depless tasks: %v", ids)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ready tasks, got %d (%v)", len(ids), ids)
	}
}

func TestAttestManualCheck(t *testing.T) {
	svc := service(t, "agent:a")
	d, err := svc.Create(store.Draft{Title: "x", Checks: []task.Check{
		{Desc: "human review", Type: "manual", Result: "pending"},
		{Desc: "build", Cmd: "exit 0", Result: "pending"},
	}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	id := d.Task.ID

	got, err := svc.Attest(id, 0, true)
	if err != nil {
		t.Fatalf("Attest: %v", err)
	}
	if got.Task.Checks[0].Result != "pass" {
		t.Fatalf("manual check = %q, want pass", got.Task.Checks[0].Result)
	}
	last := got.Provenance[len(got.Provenance)-1]
	if last.Did != "attested" || last.Who != "agent:a" {
		t.Fatalf("provenance = %+v, want attested by agent:a", last)
	}

	if _, err := svc.Attest(id, 1, true); !errors.Is(err, ErrNotManual) {
		t.Fatalf("attesting a cmd check = %v, want ErrNotManual", err)
	}
	if _, err := svc.Attest(id, 9, true); err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestAttestUnblocksClose(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x", Checks: []task.Check{{Desc: "review", Type: "manual", Result: "pending"}}})
	id := d.Task.ID

	if _, err := svc.Transition(id, "done"); err == nil {
		t.Fatal("expected refusal closing with a pending manual check")
	}
	if _, err := svc.Attest(id, 0, true); err != nil {
		t.Fatalf("Attest: %v", err)
	}
	if _, err := svc.Transition(id, "done"); err != nil {
		t.Fatalf("close after attest should pass: %v", err)
	}
}

func TestUpdateFields(t *testing.T) {
	svc := service(t, "agent:a")
	a, _ := svc.Create(store.Draft{Title: "epic"})
	b, _ := svc.Create(store.Draft{Title: "child"})

	// set priority + labels + parent
	got, err := svc.Update(b.Task.ID, UpdateFields{
		Priority: ptr("high"),
		Labels:   ptrSlice([]string{"backend", "db"}),
		Parent:   ptr(a.Task.ID),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Task.Priority != "high" || got.Task.Parent != a.Task.ID || len(got.Task.Labels) != 2 {
		t.Fatalf("update not applied: %+v", got.Task)
	}

	// invalid priority rejected
	if _, err := svc.Update(b.Task.ID, UpdateFields{Priority: ptr("ASAP")}); !errors.Is(err, task.ErrInvalidPriority) {
		t.Fatalf("invalid priority = %v, want ErrInvalidPriority", err)
	}
	// self-parent rejected as a cycle
	if _, err := svc.Update(b.Task.ID, UpdateFields{Parent: ptr(b.Task.ID)}); !errors.Is(err, task.ErrParentCycle) {
		t.Fatalf("self parent = %v, want ErrParentCycle", err)
	}
	// no-op update appends no provenance
	before := len(got.Provenance)
	noop, _ := svc.Update(b.Task.ID, UpdateFields{})
	if len(noop.Provenance) != before {
		t.Fatalf("no-op update changed provenance: %d -> %d", before, len(noop.Provenance))
	}
}

func ptr(s string) *string          { return &s }
func ptrSlice(s []string) *[]string { return &s }

func TestReorderAddsNoProvenance(t *testing.T) {
	svc := service(t, "agent:a")
	d, _ := svc.Create(store.Draft{Title: "x"})
	before := len(d.Provenance)

	got, err := svc.Reorder(d.Task.ID, 4096)
	if err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	if got.Task.Rank != 4096 {
		t.Fatalf("rank = %v, want 4096", got.Task.Rank)
	}
	if len(got.Provenance) != before {
		t.Fatalf("reorder appended provenance: %d -> %d", before, len(got.Provenance))
	}
}
