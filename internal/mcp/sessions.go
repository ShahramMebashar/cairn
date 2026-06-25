package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"cairn/internal/gitctx"
	"cairn/internal/session"
	"cairn/internal/store"
	"cairn/internal/task"
)

const ServiceVersion = "0.2.0"

const (
	ExecutionActive         = "active"
	ExecutionStalled        = "stalled"
	ExecutionAwaitingReview = "awaiting_review"
)

var (
	ErrIdentityMismatch    = errors.New("session actor does not match the bound connection actor")
	ErrClientMismatch      = errors.New("session client does not match the bound connection client")
	ErrIdempotencyRequired = errors.New("session idempotency key is required")
	ErrSessionActor        = errors.New("session belongs to another actor")
	ErrTaskClosed          = errors.New("cannot begin a session on a closed task")
)

// Identity describes the actor/client stamped on session writes.
type Identity struct {
	Actor   string `json:"actor"`
	Client  string `json:"client,omitempty"`
	Version string `json:"version"`
}

// SessionView combines durable session fields with ephemeral health.
type SessionView struct {
	session.Session
	Health session.Health `json:"health"`
	Live   *session.Live  `json:"live,omitempty"`
	// HeartbeatIntervalSeconds is the cadence at which the agent should heartbeat to keep
	// the session from going stale. Derived from config so clients need not guess.
	HeartbeatIntervalSeconds int `json:"heartbeatIntervalSeconds,omitempty"`
}

type BeginSessionInput struct {
	TaskID         string
	ExpectedActor  string
	Client         string
	Model          string
	Worktree       string
	Branch         string
	Head           string
	IdempotencyKey string
}

type HeartbeatInput struct {
	SessionID string
	Progress  string
}

type FinishSessionInput struct {
	SessionID string
	Summary   string
	Head      string
}

type CancelSessionInput struct {
	SessionID string
	Reason    string
}

// Identity returns the connection identity without mutating state.
func (svc *Service) Identity() Identity {
	return Identity{Actor: svc.actor, Client: svc.client, Version: ServiceVersion}
}

// BeginSession atomically claims/starts a task and creates its observable session.
func (svc *Service) BeginSession(ctx context.Context, in BeginSessionInput) (SessionView, error) {
	if in.ExpectedActor != svc.actor {
		return SessionView{}, fmt.Errorf("%w: expected %q, bound as %q", ErrIdentityMismatch, in.ExpectedActor, svc.actor)
	}
	if svc.client != "" && in.Client != "" && in.Client != svc.client {
		return SessionView{}, fmt.Errorf("%w: expected %q, bound as %q", ErrClientMismatch, in.Client, svc.client)
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return SessionView{}, ErrIdempotencyRequired
	}

	startedAt := svc.now().UTC()
	if in.Branch == "" || in.Head == "" {
		if ref, err := gitctx.Current(ctx, svc.store.Root()); err == nil {
			if in.Branch == "" {
				in.Branch = ref.Branch
			}
			if in.Head == "" {
				in.Head = ref.Head
			}
		}
	}
	sessionID := stableSessionID("ses_", in.TaskID, svc.actor, in.IdempotencyKey)
	attemptID := stableSessionID("att_", in.TaskID, svc.actor, in.IdempotencyKey)
	var result *store.SessionDoc

	err := svc.store.Write(ctx, svc.actor, "begin session", func(tx *store.WriteTx) error {
		existing, err := tx.FindSessionByIdempotency(in.TaskID, in.IdempotencyKey)
		if err != nil {
			return err
		}

		// Idempotent retry: a previously-successful begin returns its original session
		// regardless of any later human action (closing or reclaiming the task). Only a
		// genuinely new session (existing == nil) is subject to the begin guards below.
		if existing != nil {
			result = existing
			if live, err := tx.ReadLive(existing.Session.ID); err != nil {
				return err
			} else if live == nil {
				return tx.WriteLive(session.Live{SessionID: existing.Session.ID, HeartbeatAt: startedAt, Worktree: in.Worktree})
			}
			return nil
		}

		doc, err := tx.GetTask(in.TaskID)
		if err != nil {
			return err
		}
		cfg, err := tx.Config()
		if err != nil {
			return err
		}
		if slices.Contains(cfg.Closed, doc.Task.Status) {
			return fmt.Errorf("%w: %s", ErrTaskClosed, in.TaskID)
		}
		if doc.Task.Assignee != "" && doc.Task.Assignee != svc.actor {
			return fmt.Errorf("%w: held by %s", ErrAlreadyClaimed, doc.Task.Assignee)
		}
		live, err := tx.LiveSession(in.TaskID)
		if err != nil {
			return err
		}
		if live != nil {
			return fmt.Errorf("%w: %s", store.ErrLiveSession, live.Session.ID)
		}

		taskChanged := false
		if doc.Task.Status == cfg.Initial {
			all, err := tx.Tasks()
			if err != nil {
				return err
			}
			if err := task.CanTransition(doc.Task, cfg.Working(), all, rulesOf(cfg)); err != nil {
				return err
			}
			doc.SetStatus(cfg.Working())
			taskChanged = true
		}
		if doc.Task.Assignee != svc.actor {
			doc.SetAssignee(svc.actor)
			taskChanged = true
		}
		if doc.Task.ActiveAttempt != attemptID {
			doc.SetActiveAttempt(attemptID)
			taskChanged = true
		}
		if taskChanged {
			doc.AppendProvenance(svc.actor, "began session "+sessionID, "", startedAt)
			if err := tx.SaveTask(doc); err != nil {
				return err
			}
		}

		value := session.Session{
			ID:             sessionID,
			TaskID:         in.TaskID,
			AttemptID:      attemptID,
			Actor:          svc.actor,
			Client:         firstNonEmpty(in.Client, svc.client),
			Model:          in.Model,
			Status:         session.StatusActive,
			IdempotencyKey: in.IdempotencyKey,
			StartedAt:      startedAt,
			Branch:         in.Branch,
			HeadStarted:    in.Head,
		}
		created, err := tx.CreateSession(value)
		if err != nil {
			return err
		}
		if err := tx.WriteLive(session.Live{SessionID: sessionID, HeartbeatAt: startedAt, Worktree: in.Worktree}); err != nil {
			return err
		}
		result = created
		return nil
	})
	if err != nil {
		return SessionView{}, err
	}
	return svc.sessionView(result)
}

// Heartbeat refreshes ephemeral progress for an active session.
func (svc *Service) Heartbeat(ctx context.Context, in HeartbeatInput) (SessionView, error) {
	var result *store.SessionDoc
	err := svc.store.Write(ctx, svc.actor, "heartbeat session", func(tx *store.WriteTx) error {
		d, err := tx.GetSession(in.SessionID)
		if err != nil {
			return err
		}
		if err := svc.ownsSession(d.Session); err != nil {
			return err
		}
		if d.Session.Status != session.StatusActive {
			return fmt.Errorf("%w: %s is %s", session.ErrTerminal, d.Session.ID, d.Session.Status)
		}
		live, err := tx.ReadLive(in.SessionID)
		if err != nil {
			return err
		}
		if live == nil {
			live = &session.Live{SessionID: in.SessionID}
		}
		live.HeartbeatAt = svc.now().UTC()
		if strings.TrimSpace(in.Progress) != "" {
			live.Progress = strings.TrimSpace(in.Progress)
		}
		if err := tx.WriteLive(*live); err != nil {
			return err
		}
		result = d
		return nil
	})
	if err != nil {
		return SessionView{}, err
	}
	return svc.sessionView(result)
}

// FinishSession records a final summary and moves the task into review when configured.
func (svc *Service) FinishSession(ctx context.Context, in FinishSessionInput) (SessionView, error) {
	var result *store.SessionDoc
	if in.Head == "" {
		if ref, err := gitctx.Current(ctx, svc.store.Root()); err == nil {
			in.Head = ref.Head
		}
	}
	err := svc.store.Write(ctx, svc.actor, "finish session", func(tx *store.WriteTx) error {
		d, err := tx.GetSession(in.SessionID)
		if err != nil {
			return err
		}
		if err := svc.ownsSession(d.Session); err != nil {
			return err
		}
		taskDoc, err := tx.GetTask(d.Session.TaskID)
		if err != nil {
			return err
		}
		cfg, err := tx.Config()
		if err != nil {
			return err
		}
		review := cfg.Review()
		movingToReview := review != "" && !slices.Contains(cfg.Closed, taskDoc.Task.Status) && taskDoc.Task.Status != review

		// Enforce the review checks gate UP FRONT, before ending the session: a handoff into
		// review requires every COMMAND check to be recorded `pass` (manual checks are exempt
		// — they're attested during review). If they're pending or failing, refuse the whole
		// finish so the session stays active; the agent runs `run_checks` (which executes
		// outside this write lock) and retries. This makes "run checks before handoff" a hard
		// engine gate, not a documented suggestion — while keeping the build off the lock.
		if movingToReview {
			all, err := tx.Tasks()
			if err != nil {
				return err
			}
			if gateErr := task.CanTransition(taskDoc.Task, review, all, rulesOf(cfg)); gateErr != nil {
				return gateErr
			}
		}

		finished := d.Session
		if d.Session.Status != session.StatusFinished {
			finished, err = session.Finish(d.Session, strings.TrimSpace(in.Summary), in.Head, svc.now())
			if err != nil {
				return err
			}
			d.Replace(finished)
			if err := tx.SaveSession(d); err != nil {
				return err
			}
		}
		if err := tx.DeleteLive(d.Session.ID); err != nil {
			return err
		}

		if movingToReview {
			taskDoc.SetStatus(review)
			taskDoc.AppendProvenance(svc.actor, "finished session "+d.Session.ID, finished.Summary, svc.now())
			if err := tx.SaveTask(taskDoc); err != nil {
				return err
			}
		}
		result = d
		return nil
	})
	if err != nil {
		return SessionView{}, err
	}
	return svc.sessionView(result)
}

// CancelSession abandons a live attempt and releases its task assignment.
func (svc *Service) CancelSession(ctx context.Context, in CancelSessionInput) (SessionView, error) {
	var result *store.SessionDoc
	err := svc.store.Write(ctx, svc.actor, "cancel session", func(tx *store.WriteTx) error {
		d, err := tx.GetSession(in.SessionID)
		if err != nil {
			return err
		}
		if err := svc.ownsSession(d.Session); err != nil {
			return err
		}
		canceled := d.Session
		if d.Session.Status != session.StatusCanceled {
			canceled, err = session.Cancel(d.Session, strings.TrimSpace(in.Reason), svc.now())
			if err != nil {
				return err
			}
			d.Replace(canceled)
			if err := tx.SaveSession(d); err != nil {
				return err
			}
		}
		if err := tx.DeleteLive(d.Session.ID); err != nil {
			return err
		}

		taskDoc, err := tx.GetTask(d.Session.TaskID)
		if err != nil {
			return err
		}
		if taskDoc.Task.Assignee == svc.actor && taskDoc.Task.ActiveAttempt == d.Session.AttemptID {
			taskDoc.SetAssignee("")
			// An abandoned attempt must no longer be the task's active attempt, so the task
			// derives no execution state (finish, by contrast, retains it for awaiting_review).
			taskDoc.SetActiveAttempt("")
			taskDoc.AppendProvenance(svc.actor, "canceled session "+d.Session.ID, canceled.CancelReason, svc.now())
			if err := tx.SaveTask(taskDoc); err != nil {
				return err
			}
		}
		result = d
		return nil
	})
	if err != nil {
		return SessionView{}, err
	}
	return svc.sessionView(result)
}

// GetSession returns one session with current live health.
func (svc *Service) GetSession(id string) (SessionView, error) {
	d, err := svc.store.GetSession(id)
	if err != nil {
		return SessionView{}, err
	}
	return svc.sessionView(d)
}

// ListSessions returns newest-first sessions filtered by durable or derived fields.
func (svc *Service) ListSessions(taskID, actor, status, health string) ([]SessionView, error) {
	docs, err := svc.store.ListSessions()
	if err != nil {
		return nil, err
	}
	out := make([]SessionView, 0, len(docs))
	for _, d := range docs {
		if taskID != "" && d.Session.TaskID != taskID {
			continue
		}
		if actor != "" && d.Session.Actor != actor {
			continue
		}
		if status != "" && string(d.Session.Status) != status {
			continue
		}
		view, err := svc.sessionView(d)
		if err != nil {
			return nil, err
		}
		if health != "" && string(view.Health) != health {
			continue
		}
		out = append(out, view)
	}
	return out, nil
}

func (svc *Service) sessionView(d *store.SessionDoc) (SessionView, error) {
	live, err := svc.store.ReadLive(d.Session.ID)
	if err != nil {
		return SessionView{}, err
	}
	cfg, err := svc.store.Config()
	if err != nil {
		return SessionView{}, err
	}
	return SessionView{
		Session:                  d.Session,
		Live:                     live,
		Health:                   session.DeriveHealth(d.Session, live, svc.now(), cfg.SessionStaleDuration()),
		HeartbeatIntervalSeconds: int(cfg.SessionHeartbeatDuration().Seconds()),
	}, nil
}

func (svc *Service) executionForTask(t task.Task) (string, string) {
	docs, err := svc.store.TaskSessions(t.ID)
	if err != nil {
		// A read failure (e.g. a corrupt session file) must not masquerade as "no
		// session" — surface it so the degraded state is visible, then degrade gracefully.
		log.Printf("mcp: execution state for %s: read sessions: %v", t.ID, err)
		return "", ""
	}
	if len(docs) == 0 {
		return "", ""
	}
	cfg, err := svc.store.Config()
	if err != nil {
		log.Printf("mcp: execution state for %s: read config: %v", t.ID, err)
		return "", ""
	}
	state, sessionID, err := svc.executionFor(t, docs[0], cfg)
	if err != nil {
		log.Printf("mcp: execution state for %s: derive: %v", t.ID, err)
		return "", ""
	}
	return state, sessionID
}

// ExecutionOf returns a task's derived supervision state and latest relevant session.
func (svc *Service) ExecutionOf(t task.Task) (string, string) {
	return svc.executionForTask(t)
}

func (svc *Service) ownsSession(s session.Session) error {
	if s.Actor != svc.actor {
		return fmt.Errorf("%w: %s is owned by %s", ErrSessionActor, s.ID, s.Actor)
	}
	return nil
}

func stableSessionID(prefix, taskID, actor, key string) string {
	sum := sha256.Sum256([]byte(taskID + "\x00" + actor + "\x00" + key))
	return prefix + hex.EncodeToString(sum[:12])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
