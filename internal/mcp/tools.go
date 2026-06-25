package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/store"
	"cairn/internal/task"
)

// NewServer builds the MCP task and observable-session tools over the given Service.
// The actor identity is already baked into svc; the server itself is identity-agnostic.
func NewServer(svc *Service) *mcpsdk.Server {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "cairn", Version: ServiceVersion}, nil)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "identity",
		Description: "Return this Cairn connection's bound actor and client. Check before session writes."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, _ struct{}) (*mcpsdk.CallToolResult, Identity, error) {
			return nil, svc.Identity(), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "list",
		Description: "List tasks, optionally filtered by status, assignee, and readiness. ready=true with status=<initial> answers 'what can I start now'."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in listIn) (*mcpsdk.CallToolResult, listOut, error) {
			views, err := svc.ListWithExecution(in.Status, in.Assignee, in.Ready, in.Execution)
			if err != nil {
				return nil, listOut{}, err
			}
			out := listOut{Tasks: make([]taskOut, 0, len(views))}
			for _, v := range views {
				out.Tasks = append(out.Tasks, taskOut{ID: v.ID, Title: v.Title, Status: v.Status,
					Assignee: v.Assignee, Deps: v.Deps, Ready: v.Ready, UpdatedAt: v.UpdatedAt, Rank: v.Rank,
					Labels: v.Labels, Priority: v.Priority, Parent: v.Parent, ActiveAttempt: v.ActiveAttempt,
					ExecutionState: v.ExecutionState, SessionID: v.SessionID, Checks: toCheckOut(v.Checks)})
			}
			return nil, out, nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "get",
		Description: "Get one task: fields, checks with results, provenance, and body."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in idIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Get(in.ID)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "create",
		Description: "Create a task. The engine assigns the id and sets the initial status. Deps must already exist."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in createIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Create(store.Draft{
				Title: in.Title, Body: in.Body, Deps: in.Deps, Checks: fromCheckIn(in.Checks),
				Labels: in.Labels, Priority: in.Priority, Parent: in.Parent,
			})
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "reorder",
		Description: "Set a task's board ordering rank (manual Kanban order). Cosmetic — does not append provenance."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in reorderIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Reorder(in.ID, in.Rank)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "update",
		Description: "Edit a task: title, body, checks, priority, labels, or parent (epics/sub-tasks). Omitted fields are left unchanged; title must be non-empty; checks replaces the whole list; parent must exist and not create a cycle."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in updateIn) (*mcpsdk.CallToolResult, taskOut, error) {
			f := UpdateFields{Priority: in.Priority, Labels: in.Labels, Parent: in.Parent, Title: in.Title, Body: in.Body}
			if in.Checks != nil {
				checks := fromCheckIn(*in.Checks)
				f.Checks = &checks
			}
			doc, err := svc.Update(in.ID, f)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "delete",
		Description: "Delete a task. Refuses if other tasks name it as parent (children) or list it in deps (dependents); reparent/remove those first."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in idIn) (*mcpsdk.CallToolResult, deleteOut, error) {
			if err := svc.Delete(in.ID); err != nil {
				return nil, deleteOut{}, err
			}
			return nil, deleteOut{ID: in.ID, Deleted: true}, nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "claim",
		Description: "Claim a task for this actor. Fails if already claimed by someone else."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in idIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Claim(in.ID)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "transition",
		Description: "Move a task to a new state. Applies the deps and checks gates; closing auto-runs checks and refuses on failure."},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in transitionIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.TransitionContext(ctx, in.ID, in.To)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "run_checks",
		Description: "Run a task's cmd checks (all by default, or the given indices) and record results. Manual checks are skipped."},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in runChecksIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.RunChecksContext(ctx, in.ID, in.Only)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "note",
		Description: "Append a free-text provenance entry to a task."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in noteIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Note(in.ID, in.Text)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "edit_note",
		Description: "Edit a note's text in place (marks it editedAt). Anyone may edit any note; only note entries are editable, not system provenance."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in editNoteIn) (*mcpsdk.CallToolResult, taskOut, error) {
			idx := -1
			if in.Index != nil {
				idx = *in.Index
			}
			doc, err := svc.EditNote(in.ID, in.Note, idx, in.Text)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "delete_note",
		Description: "Delete a note. Anyone may delete any note; only note entries are deletable, not system provenance."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in deleteNoteIn) (*mcpsdk.CallToolResult, taskOut, error) {
			idx := -1
			if in.Index != nil {
				idx = *in.Index
			}
			doc, err := svc.DeleteNote(in.ID, in.Note, idx)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "attest",
		Description: "Attest a manual check (one with no cmd): set its result to pass (default) or fail. Command checks are run by run_checks, not attested."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in attestIn) (*mcpsdk.CallToolResult, taskOut, error) {
			doc, err := svc.Attest(in.ID, in.Index, in.Pass == nil || *in.Pass)
			if err != nil {
				return nil, taskOut{}, err
			}
			return nil, svc.view(doc), nil
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "begin",
		Description: fmt.Sprintf("Begin an observable work session. This connection writes as %s; expected_actor must match exactly. Heartbeat at least every heartbeatIntervalSeconds (from the result) to keep the session from going stale.", svc.actor)},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in beginSessionIn) (*mcpsdk.CallToolResult, SessionView, error) {
			out, err := svc.BeginSession(ctx, BeginSessionInput{
				TaskID: in.ID, ExpectedActor: in.ExpectedActor, Client: in.Client, Model: in.Model,
				Worktree: in.Worktree, Branch: in.Branch, Head: in.Head, IdempotencyKey: in.IdempotencyKey,
			})
			return nil, out, err
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "heartbeat",
		Description: "Refresh an active session with concise progress."},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in heartbeatIn) (*mcpsdk.CallToolResult, SessionView, error) {
			out, err := svc.Heartbeat(ctx, HeartbeatInput{SessionID: in.Session, Progress: in.Progress})
			return nil, out, err
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "finish",
		Description: "Finish an active session with a review summary. This requests review; it never closes the task."},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in finishSessionIn) (*mcpsdk.CallToolResult, SessionView, error) {
			out, err := svc.FinishSession(ctx, FinishSessionInput{SessionID: in.Session, Summary: in.Summary, Head: in.Head})
			return nil, out, err
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "cancel",
		Description: "Cancel an active session, release its task assignment, and keep the task open."},
		func(ctx context.Context, _ *mcpsdk.CallToolRequest, in cancelSessionIn) (*mcpsdk.CallToolResult, SessionView, error) {
			out, err := svc.CancelSession(ctx, CancelSessionInput{SessionID: in.Session, Reason: in.Reason})
			return nil, out, err
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "get_session",
		Description: "Get one durable session with current heartbeat health."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in sessionIDIn) (*mcpsdk.CallToolResult, SessionView, error) {
			out, err := svc.GetSession(in.Session)
			return nil, out, err
		})

	mcpsdk.AddTool(srv, &mcpsdk.Tool{Name: "list_sessions",
		Description: "List observable sessions newest first, optionally filtered by task, actor, status, or health."},
		func(_ context.Context, _ *mcpsdk.CallToolRequest, in listSessionsIn) (*mcpsdk.CallToolResult, sessionsOut, error) {
			views, err := svc.ListSessions(in.Task, in.Actor, in.Status, in.Health)
			return nil, sessionsOut{Sessions: views}, err
		})

	return srv
}

// --- tool I/O schemas (jsonschema tags drive the MCP input schema) ---

type idIn struct {
	ID string `json:"id" jsonschema:"the task id, e.g. PROJ-01j8x2k7q7f3az"`
}

type listIn struct {
	Status    string `json:"status,omitempty" jsonschema:"filter by status"`
	Assignee  string `json:"assignee,omitempty" jsonschema:"filter by assignee, e.g. agent:claude-1"`
	Ready     *bool  `json:"ready,omitempty" jsonschema:"filter by derived deps-satisfied readiness"`
	Execution string `json:"execution,omitempty" jsonschema:"filter by derived execution state: active, stalled, or awaiting_review"`
}

type beginSessionIn struct {
	ID             string `json:"id" jsonschema:"task id"`
	ExpectedActor  string `json:"expected_actor" jsonschema:"actor the caller expects this connection to use; must match identity exactly"`
	Client         string `json:"client,omitempty" jsonschema:"agent client, e.g. codex or claude"`
	Model          string `json:"model,omitempty" jsonschema:"model identifier when known"`
	Worktree       string `json:"worktree,omitempty" jsonschema:"local worktree path"`
	Branch         string `json:"branch,omitempty" jsonschema:"current Git branch"`
	Head           string `json:"head,omitempty" jsonschema:"current Git HEAD"`
	IdempotencyKey string `json:"idempotency_key" jsonschema:"unique retry key for this begin operation"`
}

type heartbeatIn struct {
	Session  string `json:"session" jsonschema:"session id"`
	Progress string `json:"progress,omitempty" jsonschema:"concise current progress, not chain-of-thought"`
}

type finishSessionIn struct {
	Session string `json:"session" jsonschema:"session id"`
	Summary string `json:"summary" jsonschema:"review-ready summary of the completed attempt"`
	Head    string `json:"head,omitempty" jsonschema:"ending Git HEAD"`
}

type cancelSessionIn struct {
	Session string `json:"session" jsonschema:"session id"`
	Reason  string `json:"reason" jsonschema:"why the attempt was abandoned"`
}

type sessionIDIn struct {
	Session string `json:"session" jsonschema:"session id"`
}

type listSessionsIn struct {
	Task   string `json:"task,omitempty" jsonschema:"filter by task id"`
	Actor  string `json:"actor,omitempty" jsonschema:"filter by actor"`
	Status string `json:"status,omitempty" jsonschema:"filter by durable status"`
	Health string `json:"health,omitempty" jsonschema:"filter by derived health"`
}

type sessionsOut struct {
	Sessions []SessionView `json:"sessions"`
}

type createIn struct {
	Title    string    `json:"title" jsonschema:"task title"`
	Body     string    `json:"body,omitempty" jsonschema:"markdown body / intent"`
	Deps     []string  `json:"deps,omitempty" jsonschema:"task ids that must be closed before this can start"`
	Checks   []checkIn `json:"checks,omitempty" jsonschema:"gate-closing checks"`
	Labels   []string  `json:"labels,omitempty" jsonschema:"free-text labels"`
	Priority string    `json:"priority,omitempty" jsonschema:"one of: low, medium, high, urgent"`
	Parent   string    `json:"parent,omitempty" jsonschema:"parent task id (epic/sub-task)"`
}

type reorderIn struct {
	ID   string  `json:"id" jsonschema:"the task id"`
	Rank float64 `json:"rank" jsonschema:"new ordering rank (use a value between two neighbors)"`
}

type updateIn struct {
	ID       string     `json:"id" jsonschema:"the task id"`
	Priority *string    `json:"priority,omitempty" jsonschema:"set priority (empty clears); omit to leave unchanged"`
	Labels   *[]string  `json:"labels,omitempty" jsonschema:"set labels (empty clears); omit to leave unchanged"`
	Parent   *string    `json:"parent,omitempty" jsonschema:"set parent id (empty clears); omit to leave unchanged"`
	Title    *string    `json:"title,omitempty" jsonschema:"set the title (must be non-empty); omit to leave unchanged"`
	Body     *string    `json:"body,omitempty" jsonschema:"set the markdown body; omit to leave unchanged"`
	Checks   *[]checkIn `json:"checks,omitempty" jsonschema:"replace the full checks list (carry result on retained checks); omit to leave unchanged"`
}

type deleteOut struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type editNoteIn struct {
	ID    string `json:"id" jsonschema:"the task id"`
	Note  string `json:"note,omitempty" jsonschema:"the note's stable id; omit only for a legacy note addressed by index"`
	Index *int   `json:"index,omitempty" jsonschema:"0-based provenance index; used only when the note id is absent"`
	Text  string `json:"text" jsonschema:"the replacement note text"`
}

type deleteNoteIn struct {
	ID    string `json:"id" jsonschema:"the task id"`
	Note  string `json:"note,omitempty" jsonschema:"the note's stable id; omit only for a legacy note addressed by index"`
	Index *int   `json:"index,omitempty" jsonschema:"0-based provenance index; used only when the note id is absent"`
}

type checkIn struct {
	Desc    string `json:"desc" jsonschema:"what the check verifies"`
	Cmd     string `json:"cmd,omitempty" jsonschema:"shell command; omit for a manual check"`
	Type    string `json:"type,omitempty" jsonschema:"set to 'manual' for an attested check"`
	Cwd     string `json:"cwd,omitempty" jsonschema:"working dir relative to repo root"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"timeout in seconds"`
	Result  string `json:"result,omitempty" jsonschema:"carry an existing result (pending|pass|fail) on edit; omit for a new check (defaults pending)"`
}

type transitionIn struct {
	ID string `json:"id" jsonschema:"the task id"`
	To string `json:"to" jsonschema:"the target state"`
}

type runChecksIn struct {
	ID   string `json:"id" jsonschema:"the task id"`
	Only []int  `json:"only,omitempty" jsonschema:"check indices to run; omit to run all"`
}

type noteIn struct {
	ID   string `json:"id" jsonschema:"the task id"`
	Text string `json:"text" jsonschema:"the note text"`
}

type attestIn struct {
	ID    string `json:"id" jsonschema:"the task id"`
	Index int    `json:"index" jsonschema:"0-based index of the manual check to attest"`
	Pass  *bool  `json:"pass,omitempty" jsonschema:"attestation result; omit or true = pass, false = fail"`
}

type listOut struct {
	Tasks []taskOut `json:"tasks"`
}

type taskOut struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Status         string             `json:"status"`
	Assignee       string             `json:"assignee,omitempty"`
	Deps           []string           `json:"deps,omitempty"`
	Ready          bool               `json:"ready"`
	UpdatedAt      string             `json:"updatedAt,omitempty"`
	Rank           float64            `json:"rank,omitempty"`
	Labels         []string           `json:"labels,omitempty"`
	Priority       string             `json:"priority,omitempty"`
	Parent         string             `json:"parent,omitempty"`
	ActiveAttempt  string             `json:"activeAttempt,omitempty"`
	ExecutionState string             `json:"executionState,omitempty"`
	SessionID      string             `json:"sessionId,omitempty"`
	Checks         []checkOut         `json:"checks,omitempty"`
	Provenance     []store.Provenance `json:"provenance,omitempty"`
	Body           string             `json:"body,omitempty"`
}

type checkOut struct {
	Desc   string `json:"desc"`
	Cmd    string `json:"cmd,omitempty"`
	Type   string `json:"type,omitempty"`
	Result string `json:"result"`
	Cwd    string `json:"cwd,omitempty"`
}

// view builds the full single-task response, computing the derived ready flag.
func (svc *Service) view(doc *store.Doc) taskOut {
	updatedAt := ""
	if n := len(doc.Provenance); n > 0 {
		updatedAt = doc.Provenance[n-1].At
	}
	executionState, sessionID := svc.executionForTask(doc.Task)
	return taskOut{ID: doc.Task.ID, Title: doc.Task.Title, Status: doc.Task.Status,
		Assignee: doc.Task.Assignee, Deps: doc.Task.Deps, Ready: svc.ReadyOf(doc.Task),
		UpdatedAt: updatedAt, Rank: doc.Task.Rank, Labels: doc.Task.Labels, Priority: doc.Task.Priority,
		Parent: doc.Task.Parent, ActiveAttempt: doc.Task.ActiveAttempt, ExecutionState: executionState,
		SessionID: sessionID, Checks: toCheckOut(doc.Task.Checks), Provenance: doc.Provenance, Body: doc.Body}
}

// ReadyOf computes a task's derived readiness best-effort; on a load error it reports
// false rather than failing the response. Used by the web adapter for single-task DTOs.
func (svc *Service) ReadyOf(t task.Task) bool {
	cfg, err := svc.store.Config()
	if err != nil {
		return false
	}
	// Resolve only this task's listed deps instead of scanning the whole board — this runs on
	// the response path of every single-task endpoint (get/note/update/transition).
	return task.ReadyFunc(t, svc.depResolver(), rulesOf(cfg))
}

func toCheckOut(checks []task.Check) []checkOut {
	if len(checks) == 0 {
		return nil
	}
	out := make([]checkOut, len(checks))
	for i, c := range checks {
		out[i] = checkOut{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Result: c.Result, Cwd: c.Cwd}
	}
	return out
}

func fromCheckIn(in []checkIn) []task.Check {
	if len(in) == 0 {
		return nil
	}
	out := make([]task.Check, len(in))
	for i, c := range in {
		result := c.Result
		if result == "" {
			result = "pending"
		}
		out[i] = task.Check{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Cwd: c.Cwd, Timeout: c.Timeout, Result: result}
	}
	return out
}
