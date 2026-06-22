package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cairn/internal/session"
)

func storedSession() session.Session {
	return session.Session{
		ID:             "ses_001",
		TaskID:         "PROJ-001",
		AttemptID:      "att_001",
		Actor:          "agent:codex",
		Client:         "codex",
		Model:          "gpt-5",
		Status:         session.StatusActive,
		IdempotencyKey: "begin-001",
		StartedAt:      time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC),
		Branch:         "codex/sessions",
		HeadStarted:    "abc123",
	}
}

func TestSessionRoundTripAndLiveState(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	value := storedSession()
	live := &session.Live{
		SessionID:   value.ID,
		HeartbeatAt: value.StartedAt,
		Progress:    "starting",
		Worktree:    s.Root(),
		Usage:       session.Usage{InputTokens: 12},
	}

	created, err := s.CreateSession(context.Background(), value.Actor, value, live)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if created.Session.ID != value.ID || created.Session.AttemptID != value.AttemptID {
		t.Fatalf("created session = %+v", created.Session)
	}

	got, err := s.GetSession(value.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got.Session.Actor != value.Actor || got.Session.Status != session.StatusActive {
		t.Fatalf("session = %+v", got.Session)
	}
	gotLive, err := s.ReadLive(value.ID)
	if err != nil {
		t.Fatalf("ReadLive: %v", err)
	}
	if gotLive == nil || gotLive.Progress != "starting" || gotLive.Usage.InputTokens != 12 {
		t.Fatalf("live = %+v", gotLive)
	}

	found, err := s.FindSessionByIdempotency(value.TaskID, value.IdempotencyKey)
	if err != nil || found == nil || found.Session.ID != value.ID {
		t.Fatalf("FindSessionByIdempotency = %+v, %v", found, err)
	}
	if current, err := s.LiveSession(value.TaskID); err != nil || current == nil || current.Session.ID != value.ID {
		t.Fatalf("LiveSession = %+v, %v", current, err)
	}
}

func TestSessionUpdatePreservesUnknownFields(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	value := storedSession()
	if _, err := s.CreateSession(context.Background(), value.Actor, value, nil); err != nil {
		t.Fatal(err)
	}
	path := s.sessionPath(value.ID)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, []byte("future_field: keep-me\n")...)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := s.GetSession(value.ID)
	if err != nil {
		t.Fatal(err)
	}
	finished, err := session.Finish(d.Session, "Implemented sessions", "def456", session.Usage{OutputTokens: 9}, value.StartedAt.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	d.Replace(finished)
	if err := s.SaveSession(context.Background(), value.Actor, d); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "future_field: keep-me") {
		t.Fatalf("unknown field lost:\n%s", after)
	}
	reloaded, err := s.GetSession(value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Session.Status != session.StatusFinished || reloaded.Session.Summary != "Implemented sessions" {
		t.Fatalf("updated session = %+v", reloaded.Session)
	}
}

func TestCreateSessionRejectsSecondLiveSession(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	first := storedSession()
	if _, err := s.CreateSession(context.Background(), first.Actor, first, nil); err != nil {
		t.Fatal(err)
	}
	second := first
	second.ID = "ses_002"
	second.IdempotencyKey = "begin-002"
	if _, err := s.CreateSession(context.Background(), second.Actor, second, nil); !errors.Is(err, ErrLiveSession) {
		t.Fatalf("second CreateSession error = %v, want ErrLiveSession", err)
	}
}

func TestSessionOptimisticConflict(t *testing.T) {
	s := New(repo(t, map[string]string{"PROJ-001": minimalTask}))
	value := storedSession()
	if _, err := s.CreateSession(context.Background(), value.Actor, value, nil); err != nil {
		t.Fatal(err)
	}
	a, _ := s.GetSession(value.ID)
	b, _ := s.GetSession(value.ID)
	aFinished, _ := session.Finish(a.Session, "first", "", session.Usage{}, value.StartedAt.Add(time.Hour))
	a.Replace(aFinished)
	if err := s.SaveSession(context.Background(), value.Actor, a); err != nil {
		t.Fatal(err)
	}
	bCanceled, _ := session.Cancel(b.Session, "second", value.StartedAt.Add(time.Hour))
	b.Replace(bCanceled)
	if err := s.SaveSession(context.Background(), value.Actor, b); !errors.Is(err, ErrSessionConflict) {
		t.Fatalf("stale save error = %v, want ErrSessionConflict", err)
	}
}

func TestRepositoryWriteLockHonorsContext(t *testing.T) {
	root := repo(t, map[string]string{"PROJ-001": minimalTask})
	first := New(root)
	second := New(root)
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- first.Write(context.Background(), "agent:a", "hold", func(*WriteTx) error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	err := second.Write(ctx, "agent:b", "contend", func(*WriteTx) error { return nil })
	if !errors.Is(err, ErrLockTimeout) {
		t.Fatalf("contended Write error = %v, want ErrLockTimeout", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first Write: %v", err)
	}

	if err := second.Write(context.Background(), "agent:b", "after release", func(*WriteTx) error { return nil }); err != nil {
		t.Fatalf("Write after release: %v", err)
	}
}
