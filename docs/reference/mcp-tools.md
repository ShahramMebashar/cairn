---
title: MCP tools
---

# MCP tools

Every verb is a thin adapter over the same rule-set in `internal/task`. Identity is fixed
at server startup via `--actor`; it is **not** a tool argument. Every **write** appends one
`provenance` entry; reads never mutate.

All write verbs return the full task (the same shape as `get`).

| Verb | R/W | Arguments | Returns |
|---|---|---|---|
| [`identity`](#agent-sessions) | R | — | bound actor/client/version |
| [`list`](#list) | R | `status?`, `assignee?`, `ready?`, `execution?` | `{ tasks: [...] }` |
| [`get`](#get) | R | `id` | task |
| [`create`](#create) | W | `title`, `body?`, `deps?`, `checks?` | task |
| [`claim`](#claim) | W | `id` | task |
| [`transition`](#transition) | W | `id`, `to` | task |
| [`run_checks`](#run_checks) | W | `id`, `only?` | task |
| [`note`](#note) | W | `id`, `text` | task |
| [`begin`](#agent-sessions) | W | task, identity, runtime metadata, retry key | session |
| [`heartbeat`](#agent-sessions) | W | session, progress | session |
| [`finish`](#agent-sessions) | W | session, summary, head | session |
| [`cancel`](#agent-sessions) | W | session, reason | session |
| [`get_session`](#agent-sessions) | R | session | session |
| [`list_sessions`](#agent-sessions) | R | task/actor/status/health filters | `{ sessions: [...] }` |

## Task shape

```json
{
  "id": "PROJ-01j8x2k7q7f3az",
  "title": "Add idempotency keys",
  "status": "in_progress",
  "assignee": "agent:claude-1",
  "deps": ["PROJ-001"],
  "ready": false,
  "checks": [
    { "desc": "tests pass", "cmd": "go test ./...", "result": "pending", "cwd": "." }
  ],
  "provenance": [
    { "who": "human:shah", "at": "2026-06-21T10:00:00Z", "did": "created" }
  ],
  "body": "Prose intent…"
}
```

`ready` is **derived** (deps-satisfied), computed on read, never stored.

---

## list

Filter the task graph. Omit a filter to ignore it.

| Arg | Type | Meaning |
|---|---|---|
| `status` | string | only this status |
| `assignee` | string | only this assignee |
| `ready` | bool | only tasks whose deps are all closed (`true`) / not (`false`) |
| `execution` | string | `active`, `stalled`, or `awaiting_review` |

`list(ready=true, status=<initial>)` is the agent's **"what can I start now"** query.

```json
{ "ready": true, "status": "backlog" }
```

## get

Full task including `body`, `checks` (+results), and `provenance`.

```json
{ "id": "PROJ-001" }
```

## create

The engine assigns the `id` (a time-ordered, collision-resistant `prefix`-`<base32>` token)
and sets `status` to the configured `initial`. Deps must already exist or the call is rejected
(no dangling graph).

| Arg | Type | Meaning |
|---|---|---|
| `title` | string | required |
| `body` | string | markdown intent (immutable to the engine afterwards) |
| `deps` | string[] | task ids that must be closed before this can start |
| `checks` | object[] | gate-closing checks (see below) |

A check object:

| Field | Meaning |
|---|---|
| `desc` | what it verifies (required) |
| `cmd` | shell command line; **omit for a manual check** |
| `type` | `manual` for an attested check |
| `cwd` | working dir relative to repo root |
| `timeout` | seconds (falls back to `check_timeout_default`) |

```json
{
  "title": "Add idempotency keys",
  "body": "Dedupe webhook deliveries.",
  "deps": ["PROJ-001"],
  "checks": [
    { "desc": "tests pass", "cmd": "go test ./internal/payments" },
    { "desc": "reviewed by a human", "type": "manual" }
  ]
}
```

## claim

Sets `assignee` to the server's actor. Re-claiming your own task is a no-op; claiming a
task already held by someone else fails.

```json
{ "id": "PROJ-002" }
```

## transition

Move a task to state `to`, applying the two gates:

1. **Deps gate** — cannot leave the `initial` state until all `deps` are closed.
2. **Checks gate** — cannot enter a closed state unless all checks pass. If they haven't,
   `transition` **auto-runs** the `cmd` checks, then closes on all-pass or **refuses** on
   any fail. (Manual checks can't be auto-run; a pending one blocks the close.)

All other transitions are free (any state → any state, including reopening a closed task).

```json
{ "id": "PROJ-002", "to": "done" }
```

> Closing can block up to the checks' timeout — that latency is intentional; closing is
> exactly when verification belongs.

See [Checks and gates](/guides/checks-and-gates) for the full gate model.

## run_checks

Run a task's `cmd` checks and record results, without transitioning. By default runs all;
`only` limits to the given check indices. Manual checks are skipped.

| Arg | Type | Meaning |
|---|---|---|
| `id` | string | required |
| `only` | int[] | check indices to run; omit to run all |

```json
{ "id": "PROJ-002", "only": [0] }
```

Full stdout+stderr (tail) is written to `.cairn/runs/<id>-<timestamp>.log`; the task
file stores only `result: pass | fail`.

## note

Append a free-text `provenance` entry — an audit breadcrumb, no state change.

```json
{ "id": "PROJ-002", "text": "blocked on upstream API change" }
```

## Agent sessions

Session-aware agents use `identity → begin → heartbeat* → finish|cancel`. `begin` atomically
claims the task, moves an initial task into the configured working state, and creates one
durable attempt. Its `expected_actor` must exactly match `identity.actor`, and its
`idempotency_key` makes retries safe.

`finish` requires a review summary and moves the task to the configured review state. It
does **not** close the task: workflow completion still requires passing checks and an
explicit `transition`. `cancel` ends only the session, releases the assignment, and leaves
the task open.

See [Sessions](/guides/sessions) for schemas, examples, health derivation, and storage.
