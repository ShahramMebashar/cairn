---
id: PROJ-009
title: 'SSE: immediate connected comment + 15s heartbeat'
status: done
checks:
  - desc: stream emits an immediate keepalive comment on open
    cmd: go test ./internal/server
    cwd: .
    result: pass
  - desc: gofmt + vet + test
    cmd: make check
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:32:25Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:32:41Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:32:47Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:33:33Z', did: note, text: 'Done in events.go: handleEvents writes ": connected\n\n" + flush immediately after WriteHeader so an idle stream proves itself on open; sseHeartbeat 25s→15s. Test TestSSESendsImmediateCommentOnOpen asserts the first line is a comment with no change triggered (RED was a 3s read timeout). Existing SSE tests unaffected (readSSEData skips comment lines).'}
  - {who: 'agent:claude', at: '2026-06-21T20:33:40Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:33:44Z', did: transitioned to done}
assignee: agent:claude
---
An idle SSE stream looks dead on open because the first byte is the 25s heartbeat. Send an immediate `: connected` comment right after the headers flush so the connection visibly proves itself, and drop the heartbeat to ~15s.

- internal/server/events.go handleEvents: write `: connected\n\n` + flush immediately after WriteHeader; change sseHeartbeat 25s → 15s.
- Test: open the stream, assert the first line is a comment (`:`-prefixed) with no change triggered.