package mcp

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"cairn/internal/check"
	"cairn/internal/config"
	"cairn/internal/session"
	"cairn/internal/store"
	"cairn/internal/task"
)

// ErrAlreadyClaimed is returned when a task is claimed by a different actor (SPEC §7).
var ErrAlreadyClaimed = errors.New("task already claimed by another actor")

// ErrNotManual is returned when attesting a check that has a command — those are executed
// by the engine (RunChecks), never attested.
var ErrNotManual = errors.New("cannot attest a command check; it is run by the engine")

// Service implements task and session tools as thin orchestration over store + task +
// check. Gate logic is never reimplemented here — it calls task.Ready / task.CanTransition.
// Identity (actor) is fixed at construction, not passed per call; every write stamps a
// provenance entry with it.
type Service struct {
	store  *store.Store
	actor  string
	client string
	now    func() time.Time
}

// NewService binds the verbs to a store and an actor identity. now is injectable for
// deterministic provenance timestamps; nil uses the wall clock.
func NewService(s *store.Store, actor string, now func() time.Time) *Service {
	return NewServiceWithClient(s, actor, "", now)
}

// NewServiceWithClient binds the verbs to a store, actor, and declared client identity.
func NewServiceWithClient(s *store.Store, actor, client string, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: s, actor: actor, client: client, now: now}
}

func rulesOf(c config.Config) task.Rules {
	return task.Rules{Initial: c.Initial, Closed: c.Closed, States: c.States, Review: c.Review()}
}

// depResolver fetches a single task by id for the deps gate, reading just that file instead
// of scanning the whole board. A missing/unreadable dep resolves to not-found, which the gate
// treats as "not closed" — matching the loaded-map behaviour without the full List() cost.
func (svc *Service) depResolver() task.DepResolver {
	return func(id string) (task.Task, bool) {
		d, err := svc.store.Get(id)
		if err != nil {
			return task.Task{}, false
		}
		return d.Task, true
	}
}

// TaskView is a task plus derived fields: readiness (SPEC §4: computed, not stored) and the
// last-activity timestamp (newest provenance entry) for "updated X ago" displays.
type TaskView struct {
	task.Task
	Ready          bool   `json:"ready"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
	ExecutionState string `json:"executionState,omitempty"`
	SessionID      string `json:"sessionId,omitempty"`
}

// List returns tasks, optionally filtered by status, assignee, and readiness. A nil
// ready pointer means "don't filter on readiness". list(ready=true, status=initial) is
// the agent's "what can I start now" query (SPEC §7).
func (svc *Service) List(status, assignee string, ready *bool) ([]TaskView, error) {
	return svc.ListWithExecution(status, assignee, ready, "")
}

// ListWithExecution adds an optional derived execution-state filter.
func (svc *Service) ListWithExecution(status, assignee string, ready *bool, execution string) ([]TaskView, error) {
	docs, err := svc.store.ListDocs()
	if err != nil {
		return nil, err
	}
	cfg, err := svc.store.Config()
	if err != nil {
		return nil, err
	}
	rules := rulesOf(cfg)
	sessionDocs, err := svc.store.ListSessions()
	if err != nil {
		return nil, err
	}
	latestSession := make(map[string]*store.SessionDoc)
	for _, d := range sessionDocs {
		if latestSession[d.Session.TaskID] == nil {
			latestSession[d.Session.TaskID] = d
		}
	}

	all := make(map[string]task.Task, len(docs))
	for _, d := range docs {
		all[d.Task.ID] = d.Task
	}

	var out []TaskView
	for _, d := range docs {
		t := d.Task
		if status != "" && t.Status != status {
			continue
		}
		if assignee != "" && t.Assignee != assignee {
			continue
		}
		r := task.Ready(t, all, rules)
		if ready != nil && *ready != r {
			continue
		}
		executionState, sessionID, err := svc.executionFor(t, latestSession[t.ID], cfg)
		if err != nil {
			return nil, err
		}
		if execution != "" && execution != executionState {
			continue
		}
		out = append(out, TaskView{Task: t, Ready: r, UpdatedAt: lastActivity(d), ExecutionState: executionState, SessionID: sessionID})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (svc *Service) executionFor(t task.Task, d *store.SessionDoc, cfg config.Config) (string, string, error) {
	if d == nil || d.Session.AttemptID != t.ActiveAttempt {
		return "", "", nil
	}
	switch d.Session.Status {
	case session.StatusActive:
		live, err := svc.store.ReadLive(d.Session.ID)
		if err != nil {
			return "", "", err
		}
		health := session.DeriveHealth(d.Session, live, svc.now(), cfg.SessionStaleDuration())
		if health == session.HealthStalled {
			return ExecutionStalled, d.Session.ID, nil
		}
		return ExecutionActive, d.Session.ID, nil
	case session.StatusFinished:
		if !slices.Contains(cfg.Closed, t.Status) {
			return ExecutionAwaitingReview, d.Session.ID, nil
		}
	}
	return "", d.Session.ID, nil
}

// lastActivity returns the timestamp of the newest provenance entry, or "" if none.
func lastActivity(d *store.Doc) string {
	if n := len(d.Provenance); n > 0 {
		return d.Provenance[n-1].At
	}
	return ""
}

// Get returns the full task: typed fields, checks (+results), provenance, and body.
func (svc *Service) Get(id string) (*store.Doc, error) {
	return svc.store.Get(id)
}

// Create mints a new task in the initial state. Deps must already exist, or the graph
// would be born dangling (SPEC §4).
func (svc *Service) Create(d store.Draft) (*store.Doc, error) {
	if !task.ValidPriority(d.Priority) {
		return nil, fmt.Errorf("%w: %q", task.ErrInvalidPriority, d.Priority)
	}
	if len(d.Deps) > 0 || d.Parent != "" {
		all, err := svc.store.List()
		if err != nil {
			return nil, err
		}
		for _, dep := range d.Deps {
			if _, ok := all[dep]; !ok {
				return nil, fmt.Errorf("%w: %s", task.ErrDanglingDep, dep)
			}
		}
		if d.Parent != "" {
			if _, ok := all[d.Parent]; !ok {
				return nil, fmt.Errorf("%w: %s", task.ErrParentMissing, d.Parent)
			}
		}
	}
	return svc.store.Create(d, svc.actor, svc.now())
}

// ErrEmptyTitle is returned when an edit would set a task's title to blank.
var ErrEmptyTitle = errors.New("title cannot be empty")

// UpdateFields are the fields editable after create. A nil pointer leaves a field unchanged;
// a non-nil pointer sets it (empty clears, where clearing is meaningful). Title/Body/Checks
// edit the task's content; Priority/Labels/Parent are the organization fields.
type UpdateFields struct {
	Priority *string
	Labels   *[]string
	Parent   *string
	Title    *string
	Body     *string
	Checks   *[]task.Check
}

// Update edits a task's content and organization fields (SPEC §7-style write; appends one
// provenance entry). Parent changes are validated to exist and not create a cycle; a Title,
// when provided, must be non-empty.
func (svc *Service) Update(id string, f UpdateFields) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if f.Priority == nil && f.Labels == nil && f.Parent == nil && f.Title == nil && f.Body == nil && f.Checks == nil {
		return doc, nil // nothing to change — don't write a spurious provenance entry
	}
	if f.Priority != nil && !task.ValidPriority(*f.Priority) {
		return nil, fmt.Errorf("%w: %q", task.ErrInvalidPriority, *f.Priority)
	}
	if f.Title != nil && strings.TrimSpace(*f.Title) == "" {
		return nil, ErrEmptyTitle
	}
	if f.Parent != nil && *f.Parent != "" {
		all, err := svc.store.List()
		if err != nil {
			return nil, err
		}
		if _, ok := all[*f.Parent]; !ok {
			return nil, fmt.Errorf("%w: %s", task.ErrParentMissing, *f.Parent)
		}
		t := all[id]
		t.Parent = *f.Parent
		all[id] = t
		if err := task.ValidateParents(all); err != nil {
			return nil, err
		}
	}
	if f.Priority != nil {
		doc.SetPriority(*f.Priority)
	}
	if f.Labels != nil {
		doc.SetLabels(*f.Labels)
	}
	if f.Parent != nil {
		doc.SetParent(*f.Parent)
	}
	if f.Title != nil {
		doc.SetTitle(*f.Title)
	}
	if f.Body != nil {
		doc.SetBody(*f.Body)
	}
	if f.Checks != nil {
		doc.SetChecks(*f.Checks)
	}
	doc.AppendProvenance(svc.actor, "updated", "", svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Delete removes a task. It refuses when other tasks reference it (children via parent,
// dependents via deps); the caller must reparent/remove those first.
func (svc *Service) Delete(id string) error {
	return svc.store.DeleteTask(id, svc.actor)
}

// Reorder sets a task's board ordering rank. Reordering is cosmetic, so it deliberately
// does NOT append a provenance entry (keeps the activity log meaningful).
func (svc *Service) Reorder(id string, rank float64) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	doc.SetRank(rank)
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Claim sets the assignee to this actor. Re-claiming one's own task is a no-op; claiming
// a task held by someone else fails (SPEC §7).
func (svc *Service) Claim(id string) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if doc.Task.Assignee == svc.actor {
		return doc, nil
	}
	if doc.Task.Assignee != "" {
		return nil, fmt.Errorf("%w: held by %s", ErrAlreadyClaimed, doc.Task.Assignee)
	}
	doc.SetAssignee(svc.actor)
	doc.AppendProvenance(svc.actor, "claimed", "", svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// Transition applies the two gates (SPEC §5). When closing is blocked solely because
// checks have not passed, it auto-runs the checks and retries — refusing only if they
// still don't pass. Deps-gate and unknown-state failures are returned without side effects.
func (svc *Service) Transition(id, to string) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	cfg, err := svc.store.Config()
	if err != nil {
		return nil, err
	}
	rules := rulesOf(cfg)
	// The deps gate needs only this task's listed deps, so resolve them on demand rather than
	// scanning and re-validating the entire board on every status change.
	deps := svc.depResolver()

	// Report deps/unknown-state failures before touching checks (deps are reported first).
	gateErr := task.CanTransitionFunc(doc.Task, to, deps, rules)
	if gateErr != nil && !errors.Is(gateErr, task.ErrChecksNotPassed) {
		return nil, gateErr
	}

	// Verification-asserting transitions (entering the review state or a closed state) ALWAYS
	// re-run the command checks fresh and never trust a previously-recorded pass — a stored
	// result can be stale relative to the current code. Other transitions commit directly.
	gated := rules.IsClosed(to) || (rules.Review != "" && to == rules.Review)
	if !gated {
		return svc.commitTransition(doc, to) // gateErr is nil: only closed/review gate checks
	}

	if err := svc.runCmdChecks(doc, cfg, nil); err != nil {
		return nil, err
	}
	doc.AppendProvenance(svc.actor, "ran checks", "", svc.now())

	if again := task.CanTransitionFunc(doc.Task, to, deps, rules); again != nil {
		if saveErr := svc.store.Save(doc); saveErr != nil { // persist the recorded results
			return nil, saveErr
		}
		return doc, again
	}
	doc.SetStatus(to)
	doc.AppendProvenance(svc.actor, "transitioned to "+to, "", svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (svc *Service) commitTransition(doc *store.Doc, to string) (*store.Doc, error) {
	doc.SetStatus(to)
	doc.AppendProvenance(svc.actor, "transitioned to "+to, "", svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// RunChecks runs the cmd checks (all by default, or the indices in `only`) and writes
// their results. Manual checks have no cmd and are skipped (SPEC §6, §7).
func (svc *Service) RunChecks(id string, only []int) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	cfg, err := svc.store.Config()
	if err != nil {
		return nil, err
	}
	var filter map[int]bool
	if len(only) > 0 {
		filter = make(map[int]bool, len(only))
		for _, i := range only {
			filter[i] = true
		}
	}
	if err := svc.runCmdChecks(doc, cfg, filter); err != nil {
		return nil, err
	}
	doc.AppendProvenance(svc.actor, "ran checks", "", svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// runCmdChecks executes each cmd check (optionally filtered) and records pass/fail on the
// doc. It mutates but does not save.
func (svc *Service) runCmdChecks(doc *store.Doc, cfg config.Config, only map[int]bool) error {
	runner := check.Runner{Root: svc.store.Root(), LogDir: svc.store.RunsDir(), Now: svc.now}
	for i, c := range doc.Task.Checks {
		if only != nil && !only[i] {
			continue
		}
		if c.Cmd == "" {
			continue // manual check: result set by attestation, not execution
		}
		res, err := runner.Run(doc.Task.ID, check.Spec{Cmd: c.Cmd, Cwd: c.Cwd, Timeout: cfg.CheckTimeout(c.Timeout)})
		if err != nil {
			return err
		}
		result := "fail"
		if res.Pass {
			result = "pass"
		}
		if err := doc.SetCheckResult(i, result); err != nil {
			return err
		}
	}
	return nil
}

// Note appends a free-text provenance entry (SPEC §7).
// Attest sets a manual check's result (SPEC §6: a check with no command is set by
// attestation, not execution). It refuses checks that have a command and out-of-range
// indices. pass=false records a failed attestation.
func (svc *Service) Attest(id string, index int, pass bool) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(doc.Task.Checks) {
		return nil, fmt.Errorf("attest: check index %d out of range", index)
	}
	if doc.Task.Checks[index].Cmd != "" {
		return nil, ErrNotManual
	}
	result := "fail"
	if pass {
		result = "pass"
	}
	if err := doc.SetCheckResult(index, result); err != nil {
		return nil, err
	}
	doc.AppendProvenance(svc.actor, "attested", fmt.Sprintf("check %d %s", index, result), svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (svc *Service) Note(id, text string) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	doc.AppendProvenance(svc.actor, "note", text, svc.now())
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// EditNote edits a note's text in place and marks it editedAt. Anyone may edit any note;
// only note entries are editable. Address by note id (preferred) or, for a legacy note with
// no id, by 0-based provenance index (pass noteID=="").
func (svc *Service) EditNote(id, noteID string, index int, text string) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if err := doc.EditNote(noteID, index, text, svc.now()); err != nil {
		return nil, err
	}
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// DeleteNote removes a note. Anyone may delete any note; only note entries are deletable.
// Address by note id (preferred) or, for a legacy note with no id, by 0-based provenance
// index (pass noteID==""). No provenance entry is appended; the deletion leaves no trace.
func (svc *Service) DeleteNote(id, noteID string, index int) (*store.Doc, error) {
	doc, err := svc.store.Get(id)
	if err != nil {
		return nil, err
	}
	if err := doc.DeleteNote(noteID, index); err != nil {
		return nil, err
	}
	if err := svc.store.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}
