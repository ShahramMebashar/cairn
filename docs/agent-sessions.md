# Agent sessions

Agent sessions make work observable without turning an agent's self-report into proof.
Each attempt links a task to an actor, client/model, heartbeat, Git context, and final
summary. The task remains the workflow object; the session records who is working and what
they report.

## Lifecycle

1. Call `identity` and read the connection's bound `actor` and `client`.
2. Call `begin` with that exact actor as `expected_actor` and a unique
   `idempotency_key`. Cairn claims the task and enters the configured working state.
3. Call `heartbeat` periodically with concise progress. Progress is status, not
   chain-of-thought.
4. Call `finish` with a useful review summary and final Git head; or call `cancel` with a
   reason.
5. Run checks and explicitly transition the task closed only after review.

```jsonc
// identity
{}

// begin
{
  "id": "PROJ-043",
  "expected_actor": "agent:codex",
  "client": "codex",
  "model": "gpt-5",
  "worktree": "/work/cairn",
  "branch": "codex/agent-sessions",
  "head": "0123abcd",
  "idempotency_key": "PROJ-043-attempt-1"
}

// heartbeat
{
  "session": "ses_…",
  "progress": "Session filters and task-detail supervision panel are implemented."
}

// finish
{
  "session": "ses_…",
  "summary": "Added Active, Stalled, and Awaiting review views plus session history in task detail.",
  "head": "89abcdef"
}
```

## Supervision states

Workflow status and execution state are separate:

- **Active** — an active session has a recent heartbeat.
- **Stalled** — the heartbeat is older than `session_stale_after` (default `5m`). This is
  derived from local time; no repair write is required.
- **Awaiting review** — the latest attempt finished and the task is in the configured review
  state.

The web UI exposes these as first-class sidebar views. Task rows carry a compact execution
signal, and task detail shows progress, heartbeat age, actor/client/model, branch/worktree,
cancellation reason, and final summary.

## Storage and concurrency

- `.cairn/sessions/<session>.yaml` is the durable, Git-friendly attempt record.
- `.cairn/live/<session>.json` is ephemeral heartbeat state and is gitignored.
- `.cairn/write.lock` uses an OS advisory lock to serialize cross-process writes; the file
  contains diagnostics but file existence is never treated as ownership.

Session YAML preserves unknown fields during updates. Absolute worktree paths and live
heartbeat timestamps remain in local live files rather than durable records.

## Trust boundary

`finish` means “the agent has stopped and supplied a review handoff.” It does not mean the
work is correct. Checks and explicit workflow closure remain separate. A later trust-loop
slice will bind check evidence to a Git tree snapshot and invalidate stale results; until
then, the existing check gate remains authoritative.
