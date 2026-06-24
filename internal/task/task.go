// Package task is the heart of cairn: the Task type plus the pure gate logic that
// decides what transitions are legal. It has no side effects and no dependency on
// config, store, or check — MCP verbs and the future UI both call these functions so
// the rules physically cannot diverge (SPEC §0, §9).
package task

import (
	"errors"
	"fmt"
	"slices"
)

// Check is a single gate-closing verification on a task (SPEC §5, §6).
// A Check with no Cmd is manual: its Result is set by attestation, not execution.
// Result is engine-managed and is one of: pending | pass | fail.
type Check struct {
	Desc    string
	Cmd     string
	Type    string // "manual" for a check with no Cmd; otherwise empty
	Result  string // pending | pass | fail
	Cwd     string // relative to repo root; defaults to repo root
	Timeout int    // seconds; falls back to config check_timeout_default
}

// Passed reports whether this check has succeeded.
func (c Check) Passed() bool { return c.Result == "pass" }

// Task is the gate-relevant view of a task file. The store parses task files into
// this struct for reads; lossless writes operate on the raw YAML node, not this type
// (SPEC §8), so unknown frontmatter keys are preserved outside of here.
type Task struct {
	ID       string
	Title    string
	Status   string
	Assignee string
	Deps     []string
	Checks   []Check
	// Optional, caller-owned organization fields. None gate transitions (only deps do);
	// Parent is grouping/rollup (epics & sub-tasks).
	Labels   []string
	Priority string  // "" | low | medium | high | urgent
	Parent   string  // id of the parent task, or ""
	Rank     float64 // manual board ordering; 0 = unset (falls back to id order)
	// ActiveAttempt identifies the session attempt currently eligible for review.
	ActiveAttempt string
}

// Rules are the config-derived inputs the gate logic needs. They are passed in (rather
// than importing the config package) so task stays a leaf in the dependency graph
// (SPEC §9). The mcp/store layer builds a Rules from config.Config.
type Rules struct {
	Initial string   // the state new tasks start in
	Closed  []string // states considered "closed" (subset of States)
	States  []string // all valid states; empty disables target-state validation
	Review  string   // the review/handoff state, or ""; entry is gated on command checks
}

// IsClosed reports whether status is one of the closed states.
func (r Rules) IsClosed(status string) bool {
	return slices.Contains(r.Closed, status)
}

// IsState reports whether status is a valid state. When States is empty, validation is
// opt-out and every status is accepted.
func (r Rules) IsState(status string) bool {
	if len(r.States) == 0 {
		return true
	}
	return slices.Contains(r.States, status)
}

// Gate sentinels. CanTransition and ValidateDeps wrap these with offending detail via
// %w, so callers match with errors.Is — e.g. the mcp transition verb auto-runs checks
// when it sees ErrChecksNotPassed (SPEC §5, §7).
var (
	ErrUnknownState    = errors.New("unknown target state")
	ErrDepsNotClosed   = errors.New("dependencies not closed")
	ErrChecksNotPassed = errors.New("checks not passed")
	ErrDanglingDep     = errors.New("dangling dependency")
	ErrCycle           = errors.New("dependency cycle")
	ErrParentMissing   = errors.New("parent not found")
	ErrParentCycle     = errors.New("parent cycle")
	ErrInvalidPriority = errors.New("invalid priority")
)

// Priorities are the allowed non-empty priority values (highest first).
var Priorities = []string{"urgent", "high", "medium", "low"}

// ValidPriority reports whether p is "" (none) or one of Priorities.
func ValidPriority(p string) bool { return p == "" || slices.Contains(Priorities, p) }

// Closed reports whether t is currently in a closed state.
func Closed(t Task, r Rules) bool { return r.IsClosed(t.Status) }

// DepResolver looks up a single task by id. ok is false when no such task exists. It lets
// the deps gate fetch only a task's listed dependencies instead of loading the whole board.
type DepResolver func(id string) (Task, bool)

// ReadyFunc reports whether every dep of t resolves to a closed task, fetching each dep on
// demand via resolve. Readiness is derived, never stored (SPEC §4). A dep that does not
// resolve yields false defensively; the real dangling-dep error is raised once, at load, by
// ValidateDeps.
func ReadyFunc(t Task, resolve DepResolver, r Rules) bool {
	for _, id := range t.Deps {
		dep, ok := resolve(id)
		if !ok || !r.IsClosed(dep.Status) {
			return false
		}
	}
	return true
}

// Ready is ReadyFunc backed by a fully-loaded task map.
func Ready(t Task, all map[string]Task, r Rules) bool {
	return ReadyFunc(t, mapResolver(all), r)
}

// mapResolver adapts a loaded task map to a DepResolver.
func mapResolver(all map[string]Task) DepResolver {
	return func(id string) (Task, bool) { t, ok := all[id]; return t, ok }
}

// CanTransition returns nil if moving t to state `to` is allowed, otherwise a wrapped
// sentinel. The two gates (SPEC §5):
//
//  1. Deps gate  — cannot LEAVE the initial state unless all deps are closed.
//  2. Checks gate — cannot ENTER a closed state unless ALL checks pass; cannot ENTER the
//     review state unless all COMMAND checks pass (manual checks are attested during
//     review, per SPEC §6, so they're exempt at review entry). Zero checks passes
//     vacuously.
//
// Gating review entry is what makes handoff un-skippable: an agent cannot finish a session
// into review while its command checks are still pending or failing.
//
// All other transitions are free, including reopening a closed task. Both gates can
// apply at once (e.g. backlog -> done directly); the deps gate is reported first.
func CanTransition(t Task, to string, all map[string]Task, r Rules) error {
	return CanTransitionFunc(t, to, mapResolver(all), r)
}

// CanTransitionFunc is CanTransition with the deps gate resolving dependencies on demand via
// resolve, so a caller need not load the whole board to move one task.
func CanTransitionFunc(t Task, to string, resolve DepResolver, r Rules) error {
	if !r.IsState(to) {
		return fmt.Errorf("%w: %q", ErrUnknownState, to)
	}

	// Deps gate: triggered only when actually leaving the initial state.
	if t.Status == r.Initial && to != r.Initial && !ReadyFunc(t, resolve, r) {
		return fmt.Errorf("%w: %s cannot leave %q with unclosed deps", ErrDepsNotClosed, t.ID, r.Initial)
	}

	// Checks gate: closed entry requires every check; review entry requires every command
	// check (manual checks are attested later, during review).
	gatedClosed := r.IsClosed(to)
	gatedReview := r.Review != "" && to == r.Review && !gatedClosed
	if gatedClosed || gatedReview {
		for i, c := range t.Checks {
			if gatedReview && c.Cmd == "" {
				continue // manual check: attested during review, not before entry
			}
			if !c.Passed() {
				return fmt.Errorf("%w: check %d (%q) is %q", ErrChecksNotPassed, i, c.Desc, c.Result)
			}
		}
	}

	return nil
}

// ValidateParents checks the parent chain of every task: each parent must exist and the
// chain must not cycle (a task can't be its own ancestor). Parents are grouping only and
// never gate transitions.
func ValidateParents(all map[string]Task) error {
	for id, t := range all {
		seen := map[string]bool{id: true}
		for cur := t.Parent; cur != ""; {
			p, ok := all[cur]
			if !ok {
				return fmt.Errorf("%w: %s -> %s", ErrParentMissing, id, cur)
			}
			if seen[cur] {
				return fmt.Errorf("%w: %s", ErrParentCycle, id)
			}
			seen[cur] = true
			cur = p.Parent
		}
	}
	return nil
}

// ValidateDeps checks the whole task set as a graph, for the store to call on load
// (SPEC §4): any dep id absent from all is a dangling-dep error; any cycle (including a
// self-loop) is a cycle error. Both are loud, non-recoverable load failures.
func ValidateDeps(all map[string]Task) error {
	// Dangling deps first: a missing node also can't be walked for cycle detection.
	for id, t := range all {
		for _, dep := range t.Deps {
			if _, ok := all[dep]; !ok {
				return fmt.Errorf("%w: %s -> %s", ErrDanglingDep, id, dep)
			}
		}
	}

	// Cycle detection via DFS with three colors (white=unseen, gray=on-stack, black=done).
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int, len(all))

	var visit func(id string) error
	visit = func(id string) error {
		color[id] = gray
		for _, dep := range all[id].Deps {
			switch color[dep] {
			case gray:
				return fmt.Errorf("%w: %s -> %s", ErrCycle, id, dep)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		color[id] = black
		return nil
	}

	for id := range all {
		if color[id] == white {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}
