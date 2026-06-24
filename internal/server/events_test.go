package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHubEmitsOnTaskFileWrite(t *testing.T) {
	root := t.TempDir()
	tasksDir := filepath.Join(root, ".cairn", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	hub := NewHub(10 * time.Millisecond)
	ch, cancel, err := hub.Subscribe(root)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	if err := os.WriteFile(filepath.Join(tasksDir, "PROJ-003.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-ch:
		if e.Type != evtTaskChanged || e.ID != "PROJ-003" {
			t.Fatalf("got %+v, want task-changed PROJ-003", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no event after task-file write")
	}
}

func TestHubEmitsOnSessionFileWrite(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cairn", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}

	hub := NewHub(10 * time.Millisecond)
	ch, cancel, err := hub.Subscribe(root)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	path := filepath.Join(root, ".cairn", "sessions", "ses_123.yaml")
	if err := os.WriteFile(path, []byte("id: ses_123\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-ch:
		if e.Type != evtSessionChanged || e.Session != "ses_123" {
			t.Fatalf("got %+v, want session-changed ses_123", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no event after session-file write")
	}
}

func TestHubRefCountedTeardown(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cairn", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	hub := NewHub(10 * time.Millisecond)

	_, c1, err := hub.Subscribe(root)
	if err != nil {
		t.Fatal(err)
	}
	_, c2, err := hub.Subscribe(root)
	if err != nil {
		t.Fatal(err)
	}
	if n := hub.activeRoots(); n != 1 {
		t.Fatalf("two subs on one root should share one watcher, got %d", n)
	}

	c1()
	if n := hub.activeRoots(); n != 1 {
		t.Fatalf("watcher stopped while a subscriber remained, got %d", n)
	}

	c2()
	deadline := time.Now().Add(time.Second)
	for hub.activeRoots() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("watcher not torn down after last unsubscribe, roots=%d", hub.activeRoots())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// A single atomic save (temp file + rename + chmod) is a burst of raw fs events for one
// task; the coalescer must collapse it to one task-changed event.
func TestCoalesceCollapsesBurstToOneTaskEvent(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/tasks/.tmp-987654" // temp file: ignored
	in <- "/x/.cairn/tasks/PROJ-003.md"
	in <- "/x/.cairn/tasks/PROJ-003.md"

	select {
	case e := <-got:
		if e.Type != evtTaskChanged || e.ID != "PROJ-003" {
			t.Fatalf("got %+v, want task-changed PROJ-003", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event emitted")
	}

	select {
	case e := <-got:
		t.Fatalf("unexpected second event from one burst: %+v", e)
	case <-time.After(50 * time.Millisecond):
	}
	close(in)
}

// Distinct task files changing in one window is a list-level change (membership/multiple),
// so the board refetches the whole list.
func TestCoalesceMultipleTasksIsListLevel(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/tasks/PROJ-001.md"
	in <- "/x/.cairn/tasks/PROJ-002.md"

	select {
	case e := <-got:
		if e.Type != evtTasksChanged || e.ID != "" {
			t.Fatalf("got %+v, want tasks-changed", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event emitted")
	}
	close(in)
}

// One task touched alongside its own session/live writes in the same window must still
// target that task, so the open task detail refreshes live (regression: a coincident
// session write used to downgrade this to a list-only refresh, leaving the detail stale).
func TestCoalesceTaskWithSessionTargetsTask(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/tasks/PROJ-048.md"
	in <- "/x/.cairn/sessions/ses_123.yaml"
	in <- "/x/.cairn/live/ses_123.json"

	select {
	case e := <-got:
		if e.Type != evtTaskChanged || e.ID != "PROJ-048" {
			t.Fatalf("got %+v, want task-changed PROJ-048", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event emitted")
	}
	close(in)
}

// A config change affects the board's states, so it is list-level too.
func TestCoalesceConfigIsListLevel(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/config.yaml"

	select {
	case e := <-got:
		if e.Type != evtTasksChanged {
			t.Fatalf("got %+v, want tasks-changed", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event emitted")
	}
	close(in)
}

func TestCoalesceSessionChange(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/sessions/ses_123.yaml"
	in <- "/x/.cairn/live/ses_123.json"

	select {
	case e := <-got:
		if e.Type != evtSessionChanged || e.Session != "ses_123" {
			t.Fatalf("got %+v, want session-changed ses_123", e)
		}
	case <-time.After(time.Second):
		t.Fatal("no event emitted")
	}
	close(in)
}

// Only ignored paths (temp files) must not arm the debounce timer or emit anything.
func TestCoalesceIgnoredPathsEmitNothing(t *testing.T) {
	in := make(chan string, 8)
	got := make(chan Event, 8)
	go coalesce(in, 10*time.Millisecond, func(e Event) { got <- e })

	in <- "/x/.cairn/tasks/.tmp-1"
	in <- "/x/.cairn/tasks/notes.txt" // non-task file

	select {
	case e := <-got:
		t.Fatalf("expected no event, got %+v", e)
	case <-time.After(50 * time.Millisecond):
	}
	close(in)
}
