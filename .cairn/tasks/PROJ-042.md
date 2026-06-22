---
id: PROJ-042
title: Expose observable session protocol
status: done
priority: urgent
labels: [backend, mcp, sessions]
parent: PROJ-040
deps: [PROJ-041]
checks:
  - desc: Go checks pass
    cmd: make check
    result: pass
  - desc: race detector passes
    cmd: go test -race ./internal/mcp ./internal/server
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:33:04Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:40:42Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:40:47Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T18:51:55Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T18:52:04Z', did: note, text: 'Exposed observable sessions through shared service orchestration, MCP, HTTP, and SSE. Added identity, begin, heartbeat, finish, cancel, session reads/lists, exact expected-actor protection, deterministic idempotent begin IDs, task active_attempt/execution state, active/stalled/awaiting-review filtering, and retry-safe terminal operations. Added --client binding and HTTP/MCP lifecycle tests; verified make check and race tests for internal/mcp and internal/server.'}
  - {who: 'agent:codex', at: '2026-06-22T18:52:13Z', did: transitioned to done}
assignee: agent:codex
---
Make session lifecycle available equally to agents and the web client.

## Scope
- Service orchestration for identity, begin, heartbeat, finish, cancel, get, and list.
- MCP tools with expected-actor protection and HTTP mirrors.
- Session-aware task DTOs, execution filters, SSE invalidation, and focused tests.

## Acceptance
- MCP and HTTP tests exercise the same lifecycle and identity mismatch fails before mutation.
