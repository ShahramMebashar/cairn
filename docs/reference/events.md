---
title: Events (SSE)
---

# Events (SSE)

`GET /api/events?path=<repo>` is a [Server-Sent Events](https://developer.mozilla.org/docs/Web/API/Server-sent_events)
stream of change signals. The server watches the file-based store, so changes made by
**any** actor — including MCP agents in a separate process — push to connected clients in
real time. See the [HTTP API](/reference/http-api) for the surrounding endpoints.

The stream carries **no task data**: each message is a signal telling the client *what to
refetch* via the REST endpoints. This keeps the stream from drifting from the DTOs.

## Transport

- `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`.
- On connect the server immediately sends a `: connected` comment so an idle stream doesn't
  look dead, then a `: ping` comment every 15s to keep the connection (and proxies) alive.
  EventSource clients ignore comment lines.
- Each event arrives as a `data:` line with a JSON payload.
- Bursts are debounced (~100ms): one atomic save fans out into several filesystem events,
  which are coalesced into a single message.

## Message shape

```json
{ "type": "task-changed", "id": "PROJ-01j8x2k7q7f3az" }
```

| Field | Present when | Meaning |
|---|---|---|
| `type` | always | event type (see below) |
| `id` | `task-changed` | the task id that changed |
| `session` | `session-changed` | the session id that changed (advisory) |

## Event types

| `type` | Emitted when | Client should refetch |
|---|---|---|
| `task-changed` | exactly one task file changed in the window | that task (detail, runs, and its sessions) — `GET /api/tasks/{id}` |
| `tasks-changed` | board config changed, or more than one task changed | the whole list — `GET /api/tasks` |
| `session-changed` | only session/live writes (no task file touched) | sessions for the path — `GET /api/sessions` (the `session` id is advisory; refresh all) |

A `task-changed` already implies the task's session queries are stale, so a coincident
session write in the same window is covered by refetching that task.
