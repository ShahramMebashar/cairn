package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"cairn/internal/session"
	"cairn/internal/store"
)

func TestSessionIdentityMismatchWritesNothing(t *testing.T) {
	svc := service(t, "agent:codex")
	taskDoc, err := svc.Create(store.Draft{Title: "observable work"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.BeginSession(context.Background(), BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:claude", Client: "codex", IdempotencyKey: "begin-1",
	})
	if !errors.Is(err, ErrIdentityMismatch) {
		t.Fatalf("BeginSession error = %v, want ErrIdentityMismatch", err)
	}
	sessions, err := svc.store.ListSessions()
	if err != nil || len(sessions) != 0 {
		t.Fatalf("sessions after mismatch = %d, %v", len(sessions), err)
	}
	reloaded, _ := svc.Get(taskDoc.Task.ID)
	if reloaded.Task.Status != "backlog" || reloaded.Task.Assignee != "" {
		t.Fatalf("task mutated after mismatch: %+v", reloaded.Task)
	}
}

func TestBeginSessionIsIdempotentAndClaimsTask(t *testing.T) {
	svc := NewServiceWithClient(service(t, "agent:codex").store, "agent:codex", "codex", nil)
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	in := BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", Client: "codex", Model: "gpt-5",
		Worktree: svc.store.Root(), Branch: "codex/sessions", Head: "abc123", IdempotencyKey: "begin-1",
	}

	first, err := svc.BeginSession(context.Background(), in)
	if err != nil {
		t.Fatalf("BeginSession: %v", err)
	}
	second, err := svc.BeginSession(context.Background(), in)
	if err != nil {
		t.Fatalf("retry BeginSession: %v", err)
	}
	if first.ID != second.ID || first.AttemptID != second.AttemptID {
		t.Fatalf("retry changed identity: first=%+v second=%+v", first.Session, second.Session)
	}

	reloaded, _ := svc.Get(taskDoc.Task.ID)
	if reloaded.Task.Status != "in_progress" || reloaded.Task.Assignee != "agent:codex" || reloaded.Task.ActiveAttempt != first.AttemptID {
		t.Fatalf("task after begin = %+v", reloaded.Task)
	}
	begins := 0
	for _, p := range reloaded.Provenance {
		if p.Did == "began session "+first.ID {
			begins++
		}
	}
	if begins != 1 {
		t.Fatalf("begin provenance count = %d, want 1", begins)
	}

	in.IdempotencyKey = "begin-2"
	if _, err := svc.BeginSession(context.Background(), in); !errors.Is(err, store.ErrLiveSession) {
		t.Fatalf("second live begin error = %v, want ErrLiveSession", err)
	}
}

func TestHeartbeatAndFinishSession(t *testing.T) {
	at := time.Date(2026, 6, 22, 9, 0, 0, 0, time.UTC)
	rootSvc := service(t, "agent:codex")
	svc := NewServiceWithClient(rootSvc.store, "agent:codex", "codex", func() time.Time { return at })
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	started, err := svc.BeginSession(context.Background(), BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", Client: "codex", IdempotencyKey: "begin-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	at = at.Add(time.Minute)
	heartbeat, err := svc.Heartbeat(context.Background(), HeartbeatInput{
		SessionID: started.ID, Progress: "running tests", Usage: session.Usage{InputTokens: 100, OutputTokens: 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	if heartbeat.Live == nil || heartbeat.Live.Progress != "running tests" || heartbeat.Live.Usage.InputTokens != 100 {
		t.Fatalf("heartbeat = %+v", heartbeat)
	}

	at = at.Add(time.Minute)
	finished, err := svc.FinishSession(context.Background(), FinishSessionInput{
		SessionID: started.ID, Summary: "Implemented observable sessions", Head: "def456",
		Usage: session.Usage{InputTokens: 90, OutputTokens: 30},
	})
	if err != nil {
		t.Fatal(err)
	}
	if finished.Status != session.StatusFinished || finished.Health != session.HealthFinished || finished.Usage.InputTokens != 100 || finished.Usage.OutputTokens != 30 {
		t.Fatalf("finished = %+v", finished)
	}
	if live, _ := svc.store.ReadLive(started.ID); live != nil {
		t.Fatalf("live state retained after finish: %+v", live)
	}
	reloaded, _ := svc.Get(taskDoc.Task.ID)
	if reloaded.Task.Status != "in_review" {
		t.Fatalf("task status = %q, want in_review", reloaded.Task.Status)
	}
	views, err := svc.ListWithExecution("", "", nil, ExecutionAwaitingReview)
	if err != nil || len(views) != 1 || views[0].SessionID != started.ID {
		t.Fatalf("awaiting-review list = %+v, %v", views, err)
	}

	// A retry completes any later task step without mutating the terminal session again.
	if retry, err := svc.FinishSession(context.Background(), FinishSessionInput{SessionID: started.ID}); err != nil || retry.ID != started.ID {
		t.Fatalf("finish retry = %+v, %v", retry, err)
	}
}

func TestCancelSessionReleasesAssignee(t *testing.T) {
	svc := service(t, "agent:codex")
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	started, _ := svc.BeginSession(context.Background(), BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", IdempotencyKey: "begin-1",
	})
	canceled, err := svc.CancelSession(context.Background(), CancelSessionInput{SessionID: started.ID, Reason: "superseded"})
	if err != nil {
		t.Fatal(err)
	}
	if canceled.Status != session.StatusCanceled {
		t.Fatalf("status = %q", canceled.Status)
	}
	reloaded, _ := svc.Get(taskDoc.Task.ID)
	if reloaded.Task.Assignee != "" || reloaded.Task.Status != "in_progress" || reloaded.Task.ActiveAttempt != "" {
		t.Fatalf("task after cancel = %+v", reloaded.Task)
	}
	if state, _ := svc.ExecutionOf(reloaded.Task); state != "" {
		t.Fatalf("execution after cancel = %q, want none", state)
	}
}

// A previously-successful begin must remain idempotent even after a human closes the task:
// retrying with the same idempotency key returns the original session, not an error.
func TestBeginAfterCloseIsIdempotent(t *testing.T) {
	svc := service(t, "agent:codex")
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	in := BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", IdempotencyKey: "begin-1",
	}
	first, err := svc.BeginSession(context.Background(), in)
	if err != nil {
		t.Fatalf("BeginSession: %v", err)
	}
	if _, err := svc.Transition(taskDoc.Task.ID, "done"); err != nil {
		t.Fatalf("Transition to done: %v", err)
	}
	second, err := svc.BeginSession(context.Background(), in)
	if err != nil {
		t.Fatalf("retry begin after close = %v, want nil", err)
	}
	if second.ID != first.ID || second.AttemptID != first.AttemptID {
		t.Fatalf("retry after close changed identity: first=%+v second=%+v", first.Session, second.Session)
	}
}

// finish retains the task's active_attempt so it derives awaiting_review, while cancel
// clears it; this guards the asymmetry between the two terminal verbs.
func TestFinishRetainsActiveAttempt(t *testing.T) {
	svc := service(t, "agent:codex")
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	started, _ := svc.BeginSession(context.Background(), BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", IdempotencyKey: "begin-1",
	})
	if _, err := svc.FinishSession(context.Background(), FinishSessionInput{SessionID: started.ID, Summary: "done work"}); err != nil {
		t.Fatalf("FinishSession: %v", err)
	}
	reloaded, _ := svc.Get(taskDoc.Task.ID)
	if reloaded.Task.ActiveAttempt != started.AttemptID {
		t.Fatalf("active_attempt after finish = %q, want %q", reloaded.Task.ActiveAttempt, started.AttemptID)
	}
	if state, _ := svc.ExecutionOf(reloaded.Task); state != ExecutionAwaitingReview {
		t.Fatalf("execution after finish = %q, want awaiting_review", state)
	}
}

func TestSessionHealthBecomesStalled(t *testing.T) {
	at := time.Date(2026, 6, 22, 9, 0, 0, 0, time.UTC)
	rootSvc := service(t, "agent:codex")
	svc := NewService(rootSvc.store, "agent:codex", func() time.Time { return at })
	taskDoc, _ := svc.Create(store.Draft{Title: "observable work"})
	started, _ := svc.BeginSession(context.Background(), BeginSessionInput{
		TaskID: taskDoc.Task.ID, ExpectedActor: "agent:codex", IdempotencyKey: "begin-1",
	})
	at = at.Add(4 * time.Minute)
	view, err := svc.GetSession(started.ID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Health != session.HealthStalled {
		t.Fatalf("health = %q, want stalled", view.Health)
	}
	views, err := svc.ListWithExecution("", "", nil, ExecutionStalled)
	if err != nil || len(views) != 1 {
		t.Fatalf("stalled list = %+v, %v", views, err)
	}
}
