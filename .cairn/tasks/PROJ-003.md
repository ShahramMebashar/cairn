---
id: PROJ-003
title: fsnotify watcher + event hub
status: done
checks:
  - desc: watcher emits a single debounced event on file write
    cmd: go test ./internal/server
    cwd: .
    result: pass
  - desc: gofmt + vet + test pass
    cmd: make check
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T19:47:28Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T19:57:18Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T19:57:23Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T19:59:30Z', did: note, text: 'Split into pure core + fsnotify glue. Core done in events.go: classify(path)+coalesce() — trailing debounce collapses the temp-create/rename/chmod burst of one atomic save into a single Event; one task touched in a window → task-changed{id}, config or multiple → tasks-changed; .tmp-* and non-.md ignored. Must watch the tasks DIR (atomic temp+rename swaps the inode). 4 deterministic tests pass with an injected channel. Next: Hub (Subscribe + ref-counted per-root watcher) wiring real fsnotify.'}
  - {who: 'agent:claude', at: '2026-06-21T20:00:53Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:01:01Z', did: note, text: 'Done. Hub in events.go: NewHub(debounce), Subscribe(root)→(chan Event, cancel, err) lazily starts one fsnotify watcher per root over .cairn + .cairn/tasks, ref-counted teardown on last cancel (idempotent via sync.Once). broadcast drops to full subscriber buffers so a slow client can''t stall the watcher. Tests: TestHubEmitsOnTaskFileWrite (real fs), TestHubRefCountedTeardown, + 4 coalescer tests; stable over -count=3. fsnotify v1.10.1 added. Hub not yet wired to Server — that''s PROJ-005 (SSE endpoint).'}
  - {who: 'agent:claude', at: '2026-06-21T20:01:08Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T13:38:33Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-22T13:40:07Z', did: transitioned to in_progress}
  - {who: 'human:web', at: '2026-06-22T13:40:09Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T18:45:43Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:45:43Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T18:46:16Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:46:16Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T18:49:58Z', did: note, text: test}
  - {who: 'human:shaho', at: '2026-06-24T18:50:04Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:50:04Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T18:50:07Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:50:07Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T18:50:19Z', did: updated}
  - {who: 'human:shaho', at: '2026-06-24T18:50:22Z', did: updated}
  - {who: 'human:shaho', at: '2026-06-24T19:09:22Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:09:22Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:09:25Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:09:25Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:10:16Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:10:16Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:10:18Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:10:18Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:10:24Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:10:24Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:10:27Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:10:27Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:11:58Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:11:58Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:12:00Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:12:00Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:13:14Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-24T19:13:17Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:13:17Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:13:35Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:13:35Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:14:43Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-24T19:14:45Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:14:45Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:14:47Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:14:47Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T20:20:35Z', did: updated}
assignee: agent:claude
---
Backend foundation for real-time board sync. New `internal/server/events.go`.

- `fsnotify` watcher on `.cairn/tasks/` and `.cairn/config.yaml` per project root.
- Debounce \~100ms; coalesce duplicate write/chmod events. Parse task id from filename.
- Per-root subscriber hub; broadcast on debounced change.
- Watcher started lazily on first subscriber for a root, stopped (ref-counted) on last disconnect.
- Pure transport/adapter — no gate-logic changes. Add `fsnotify` dependency.

See docs/superpowers/specs/2026-06-21-sse-realtime-and-check-output-design.md