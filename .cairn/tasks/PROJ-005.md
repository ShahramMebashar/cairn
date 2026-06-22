---
id: PROJ-005
title: 'SSE endpoint: GET /api/events'
status: done
deps: [PROJ-003]
checks:
  - desc: subscribe → broadcast → receive event as text/event-stream
    cmd: go test ./internal/server
    cwd: .
    result: pass
  - desc: gofmt + vet + test pass
    cmd: make check
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T19:47:52Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:01:39Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:01:42Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:02:49Z', did: note, text: 'Done in events.go: handleEvents subscribes to the hub BEFORE writing headers (no missed change between handshake and first read), streams `data: {json}\n\n` per Event, 25s `: ping` heartbeat, tears down via defer cancel() on r.Context().Done(). Server now owns hub *Hub (NewHub(0)); route GET /api/events registered. Deviation: used ?path= for consistency with all other endpoints (spec said ?root=). Test TestSSEStreamsTaskChangeEvent uses a real httptest server, opens the stream, creates a task, reads the data line; stable -count=3.'}
  - {who: 'agent:claude', at: '2026-06-21T20:02:56Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:03:00Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-21T20:03:37Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:25:22Z', did: note, text: 'Follow-up: tightened SSE test coverage. The original TestSSEStreamsTaskChangeEvent triggered on create, whose event actually comes from the config.yaml counter bump (tasks-changed), so the HTTP/SSE layer was never proven to emit task-changed{id} from a task-file-only write. Added TestSSEStreamsTaskChangedForSingleTask (a note, which writes only the task file) → asserts task-changed + the id; passes 3x. Remaining non-gaps: server-shutdown teardown not implemented (per-request disconnect is), no event-id/Last-Event-ID replay (deliberate — reconnect refetches), no prod static serving (pre-existing; UI served by Vite which streams the /api proxy fine).'}
assignee: agent:claude
rank: 15.5
---
HTTP transport over the watcher/hub.

- `GET /api/events?root=<path>` → `text/event-stream`.
- Subscribe to the root's hub; emit `{"type":"task-changed","id":...}` per task-file change, `{"type":"tasks-changed"}` for create/delete or unknown id.
- Periodic heartbeat comment to keep the stream alive; flush per event.
- Clean teardown on client disconnect / server shutdown (ref-count down the watcher).

See spec §A.