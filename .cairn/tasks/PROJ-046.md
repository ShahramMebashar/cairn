---
id: PROJ-046
title: Enforce check-gate at handoff and re-verify on close
status: done
priority: high
labels: [backend, workflow, gates]
checks:
  - desc: go tests pass
    cmd: go test ./...
    cwd: .
    result: pass
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T16:52:12Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T16:52:19Z', did: began session ses_7c0c1b09ae5f06e7fdbbf768}
  - {who: 'agent:claude', at: '2026-06-24T16:57:53Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T16:58:07Z', did: note, text: 'Done. #1: task.CanTransition now gates entry into the review state (Rules.Review) on COMMAND checks passing; manual checks exempt (attested during review). FinishSession enforces this gate BEFORE ending the session, so a failing/pending-cmd handoff refuses and leaves the session active. Key design call: finish gates on the *recorded* result and refuses (agent runs run_checks, which executes off the write lock) rather than executing builds itself — running a multi-minute build inside the FinishSession write-tx would hold the repo flock and block every other writer. #2: Transition into review/closed always re-runs cmd checks fresh (dropped the trust-stored-pass fast path); a stale recorded pass can''t close. Chose re-run over git-fingerprint freshness because cairn core deliberately doesn''t shell git (clients pass HEAD) — fingerprinting belongs to the deferred #4/#5 git layer. Tests: task_test TestCanTransitionReviewGate (cmd-pending blocks, manual-pending allows, close still needs manual); mcp TestFinishRefusesUnrunChecks (regression guard for the PROJ-045 miss), TestFinishAllowsPendingManualCheck, TestTransitionDoesNotTrustStaleStoredPass (stored pass + failing cmd → close refused, result corrected to fail). Docs: SPEC §5 (two-entry checks gate + freshness), transition row, WORKFLOW.md steps 6-7. go test ./... + pnpm build green via run_checks; gofmt + vet clean. Caveat: the session''s cairn binary predates this; rebuild + reconnect for the live finish/transition to enforce the new gate.'}
  - {who: 'agent:claude', at: '2026-06-24T16:58:18Z', did: finished session ses_7c0c1b09ae5f06e7fdbbf768, text: 'Made the checks gate un-skippable at the engine level. #1: entering the review state now requires every command check to pass (task.Rules.Review + CanTransition); finish enforces it before ending the session, so a pending/failing-cmd handoff refuses and the session stays active. Manual checks remain exempt at handoff (attested during review). #2: transitions into review/closed always re-run cmd checks fresh — a stale recorded pass can no longer close a task. Finish gates on recorded results and pushes the actual run to run_checks (off the write lock) to avoid holding the repo lock during a build. Tests added in internal/task and internal/mcp (incl. a regression guard for the exact PROJ-045 miss). Docs: SPEC §5, WORKFLOW.md. go test ./... + pnpm build both pass (run via run_checks); gofmt + vet clean. Review focus: the recorded-vs-rerun split between finish and close, and whether re-running on every close is acceptable latency vs. the deferred git-fingerprint optimization. Heads-up: this session''s cairn binary predates the change, so rebuild + reconnect before the gate is live in the running server.'}
  - {who: 'agent:claude', at: '2026-06-24T16:58:22Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_7c0c1b09ae5f06e7fdbbf768
---
Make the checks gate impossible for an agent to skip — the lesson from PROJ-045's botched handoff (finished with checks still `pending` because they were run out-of-band). Move enforcement from WORKFLOW.md prose into the engine.

## Scope (engine-level, ships in the binary → every project)
- **#1 review gate** — `task.CanTransition` gates entry into the review state on *command* checks passing (manual checks stay exempt until attested during review). `finish` runs the cmd checks and enforces the gate *before* it ends the session, so a failure aborts cleanly (session stays active) instead of handing off with un-run/failing checks.
- **#2 close freshness** — `transition` into the review **or** closed state always re-runs cmd checks fresh; it never trusts a previously-recorded `pass`, which can be stale vs. the current code. (Re-run, not git fingerprinting — cairn core deliberately doesn't shell git; that's the deferred #4/#5 layer.)

## Files
- `internal/task/task.go` — `Rules.Review`; review-entry checks gate (cmd-only).
- `internal/mcp/service.go` — `rulesOf` sets `Review`; `Transition` re-runs fresh for gated targets.
- `internal/mcp/sessions.go` — `FinishSession` enforces the review gate up front.

## Acceptance
- `finish` with a failing cmd check refuses and leaves the session active.
- `finish` with a pending manual check still hands off to review (attested later).
- `transition` to a closed state re-runs checks and refuses if a stored `pass` is now stale/failing.