---
title: HTTP API
---

# HTTP API

`cairn web` serves this HTTP API. It mirrors the [MCP tools](/reference/mcp-tools) over
HTTP by reusing the same `mcp.Service` rule-set, so the web and agent front-ends can't
drift.

This is a **local, single-user tool**. There is **no authentication** — the trust model is
like a git author: writes are stamped with an actor, but anyone who can reach the port can
call the API. Don't expose it to a network you don't control.

## Conventions

- **`?path=<repo>`** — every endpoint accepts the project folder to operate on, falling
  back to the server's launch `--repo`. The path must resolve to an existing directory;
  `~` is expanded.
- **Actor stamping** — writes are stamped with the actor from the `X-Cairn-Actor` header
  (URL-encoded; sanitized to a single bounded line). It falls back to `?actor=`, then to
  the server default (`human:web` unless overridden).
- **Responses** — JSON. A task response is the task DTO (id, title, status, assignee, deps,
  derived `ready`, checks, provenance, body, execution state). Errors are
  `{ "error": "…" }` with a mapped status: `404` not found, `409` conflict (already
  claimed, live session), `422` rule violation (deps gate, checks gate, bad input).

## Status & init

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/status` | workspace status: `{ initialized, root, prefix?, suggestedPrefix, states, closed, initial, review, actor, suggestedActor }` |
| `GET` | `/api/identity` | the server's bound actor |
| `POST` | `/api/init` | scaffold `.cairn/` — body `{ path?, prefix? }` |
| `GET` | `/healthz` | readiness probe (no `?path`); returns `ok` |

## Tasks

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/tasks` | list tasks — query `?status`, `?assignee`, `?ready`, `?execution` |
| `POST` | `/api/tasks` | create — body `{ title, body?, deps?, checks?, labels?, priority?, parent? }` |
| `GET` | `/api/tasks/{id}` | full task |
| `DELETE` | `/api/tasks/{id}` | delete — refused if it has children/dependents |
| `GET` | `/api/tasks/{id}/runs` | check-run log history: `{ runs: [{ file, at, cmd, cwd, exit, timedout, duration, output }] }` |
| `POST` | `/api/tasks/{id}/update` | edit fields — body `{ title?, body?, checks?, priority?, labels?, parent? }` |
| `POST` | `/api/tasks/{id}/reorder` | set board rank — body `{ rank }` |
| `POST` | `/api/tasks/{id}/transition` | move state (deps + checks gates) — body `{ to }` |
| `POST` | `/api/tasks/{id}/claim` | assign to the request actor |
| `POST` | `/api/tasks/{id}/run_checks` | run `cmd` checks — body `{ only? }` (check indices) |
| `POST` | `/api/tasks/{id}/attest` | attest a manual check — body `{ index, pass? }` (`pass` defaults true) |
| `POST` | `/api/tasks/{id}/note` | append a provenance note — body `{ text }` |
| `PATCH` | `/api/tasks/{id}/notes/{note}` | edit a note — body `{ text }` (use `{note}="-"` + `?index=` for a legacy note) |
| `DELETE` | `/api/tasks/{id}/notes/{note}` | delete a note (use `{note}="-"` + `?index=` for a legacy note) |

See [Task files](/guides/task-files) for the on-disk shape and
[Checks and gates](/guides/checks-and-gates) for transition rules.

## Sessions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/tasks/{id}/sessions` | sessions for one task |
| `POST` | `/api/tasks/{id}/sessions/begin` | open a session on a task |
| `GET` | `/api/sessions` | list sessions (filters) |
| `GET` | `/api/sessions/{session}` | one session |
| `POST` | `/api/sessions/{session}/heartbeat` | report progress |
| `POST` | `/api/sessions/{session}/finish` | finish into review (summary required) |
| `POST` | `/api/sessions/{session}/cancel` | cancel the session (reason required) |

See [Sessions](/guides/sessions) for the lifecycle and health derivation.

## Events

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/events` | `text/event-stream` of change signals — query `?path` |

The server watches the file-based store, so changes from **any** actor (including MCP
agents in a separate process) push to connected UIs. See [Events (SSE)](/reference/events)
for the message shape and event types.

## Integrations

One-click agent wiring: the server detects installed agents and writes their MCP config
files itself (it already runs locally with the user's permissions).

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/connect` | detect agents — query `?path` → `{ agents: [AgentStatus] }` |
| `POST` | `/api/connect/{agent}` | write the agent's MCP config — body `{ path?, actor? }` → `{ connected, path }` |
| `DELETE` | `/api/connect/{agent}` | remove the agent's cairn config — query `?path` → `{ connected:false, path }` |
| `GET` | `/api/connect/{agent}/manual` | manual setup guide — query `?path`, `?actor` → `{ path, lang, config }` |

For `POST /api/connect/{agent}`, an empty `actor` defaults to `agent:<id>` — the request
actor is deliberately **not** used as a fallback, since an agent's writes must be stamped
as that agent.

## MCP

| Method | Path | Description |
|---|---|---|
| `*` | `/mcp` | MCP over Streamable HTTP — query `?repo=<abs path>&actor=<actor>` (`&client=` optional) |

Each connection binds its own repo and identity via the query and reuses the same rule-set
as `cairn serve`. Both `?repo` and `?actor` are required; a missing or bad value returns a
`400`. The workspace is auto-initialized on connect.
