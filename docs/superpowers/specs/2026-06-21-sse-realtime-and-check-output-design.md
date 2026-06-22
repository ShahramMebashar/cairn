# Real-time board sync + check output viewing

Date: 2026-06-21
Status: approved (design)

## Goal

Two related capabilities for the cairn web UI:

1. **Real-time board sync** — the board and task views update live when *any*
   actor changes the task graph, including MCP agents running in a separate
   process.
2. **Check output viewing** — show the captured stdout/stderr of a task's checks
   in the UI, after the checks finish.

## Constraints discovered

- MCP (stdio) agents run in a **separate process** from the web server. Both
  mutate the file-based `.cairn/` store. An in-process event bus cannot observe
  MCP-driven changes, so realtime must be driven by **watching the filesystem**.
- `RunChecks` is synchronous; a check's output is captured (tail ~8KB) and written
  to `.cairn/runs/<id>-<timestamp>.log`. The log header records `cmd`, `cwd`,
  `exit`, `timedout`, `duration`.
- **SPEC §149 is frozen**: the task file stores only `result: pass|fail` per
  check; full output stays in the gitignored run log. The design must not add
  output (or a log reference) to the task frontmatter.
- Gate logic lives only in `internal/task` / `internal/mcp`. This feature is pure
  transport/adapter and must not touch gate rules.

## A. Real-time board sync (SSE + filesystem watch)

### Components

- **Watcher** — `fsnotify` watching `.cairn/tasks/` (and `.cairn/config.yaml`) for
  a project root. Events are debounced ~100ms and coalesced to absorb duplicate
  write/chmod notifications. The task id is parsed from the filename when
  possible.
- **Hub** — a per-root set of subscribers (SSE connections). A debounced watcher
  change broadcasts to all subscribers for that root. Watchers are started
  lazily on the first subscriber for a root and stopped (ref-counted) when the
  last subscriber for that root disconnects, so idle projects are not watched.
- **Endpoint** — `GET /api/events?root=<path>` returns `text/event-stream`. Sends
  a heartbeat comment periodically so proxies/`EventSource` keep the connection
  alive; `EventSource` auto-reconnects on drop.
- Lives in a new `internal/server/events.go`. New dependency: `fsnotify`.

### Event payload — signal only

Events carry a signal, not data:

- `{"type":"task-changed","id":"PROJ-003"}` — a single task file changed.
- `{"type":"tasks-changed"}` — create/delete (list membership changed) or any
  change where a specific id can't be determined.

The client invalidates the matching React Query keys and refetches via the
existing REST endpoints. This **reuses the existing invalidate-on-mutation
machinery**, avoids serializing DTOs over a second channel, and removes any risk
of the SSE payload diverging from the REST DTO.

Rejected alternative: full-payload SSE (push the whole task object). Saves one
roundtrip but duplicates serialization and risks divergence. Not worth it.

### Client

A single `EventSource` opened for the active project root. On `task-changed`,
invalidate that task's key (and the runs key, below); on `tasks-changed`,
invalidate the tasks-list key. Close/reopen when the active root changes.

## B. Check output viewing

### Endpoint

`GET /api/tasks/{id}/runs` returns the task's run logs, newest-first:

```json
{ "runs": [
  { "file": "PROJ-003-20260621-185736.523.log",
    "at": "2026-06-21T18:57:36Z",
    "cmd": "go test ./...",
    "cwd": ".",
    "exit": 0,
    "timedout": false,
    "duration": "1.2s",
    "output": "....tail...." }
] }
```

Implemented by listing `.cairn/runs/<id>-*.log`, parsing each log header, and
returning the body. Missing runs dir → empty list, not an error. No task-file
schema change, so SPEC §149 stays intact.

### Check → run mapping

Each log header already contains the `cmd`. The UI groups runs by `cmd` and shows
the latest run per check. This gives precise per-check output with no persisted
linkage.

### UI

In `TaskDetail.tsx`, each cmd check row gains an expandable output panel
(shadcn `Collapsible` + `ScrollArea`, monospace) showing its latest run's parsed
header and output. When SSE signals the task changed after checks run, the runs
query invalidates and the panel refreshes live.

## Combined data flow

1. An MCP agent (separate process) runs checks → writes `.cairn/runs/*.log` and
   updates the task file `result`.
2. `fsnotify` observes the task-file change → debounced → hub broadcasts
   `task-changed`.
3. The web client invalidates the task DTO query and the runs query → refetches.
4. The board status icon **and** the expanded check output update live.

## Error handling

- SSE: heartbeat comment keeps the stream open; `EventSource` auto-reconnects;
  streams close cleanly on server shutdown / client disconnect.
- Watcher: log and degrade gracefully — the UI still works via mutations and
  manual refetch if watching fails.
- Runs endpoint: absent/empty runs dir → empty list; unparseable log → skip or
  return raw body without a parsed header rather than erroring the whole request.

## Testing

- `internal/server`: SSE handler — subscribe, trigger a broadcast, assert the
  event is received and formatted as `text/event-stream`. Runs endpoint — seed a
  log file under a temp runs dir, assert parsed JSON.
- Watcher: write a file in a temp dir, assert a single debounced event is emitted.
- Gate-logic tests untouched.

## Scope cuts (YAGNI)

- No live line-by-line streaming of check output.
- No full-payload SSE.
- No persisted check→log linkage in task frontmatter.
- No auth on the SSE endpoint (matches the rest of the local API).
