---
id: PROJ-041
title: Build session domain and durable storage
status: done
priority: urgent
labels: [backend, sessions]
parent: PROJ-040
checks:
  - desc: Go checks pass
    cmd: make check
    result: pass
  - desc: race detector passes
    cmd: go test -race ./internal/session ./internal/store
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:32:53Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:33:39Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:33:44Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T18:40:13Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T18:40:20Z', did: note, text: 'Implemented the session foundation in internal/session and internal/store: pure active/finished/canceled lifecycle and health derivation, cumulative usage, random IDs, lossless durable YAML, atomic gitignored live state, active-session/idempotency lookup, optimistic session writes, and repository-wide POSIX flock transactions. Repo init now creates/ignores session runtime paths; config writes are atomic and tasks carry active_attempt. Verified with make check, race tests, and lock/contention/store round-trip tests.'}
  - {who: 'agent:codex', at: '2026-06-22T18:40:25Z', did: transitioned to done}
assignee: agent:codex
---
Create the Slice 1 session foundation with cross-process-safe writes.

## Scope
- POSIX repository write lock used by every Cairn mutation.
- Pure session lifecycle types and validation.
- Durable YAML sessions plus gitignored atomic live-state files.
- Random session/attempt IDs, one-live-session validation, health derivation, and idempotency.

## Acceptance
- Table-driven lifecycle/store tests cover round trips, stale health, duplicate begin, concurrent writers, and crash-safe retry ordering.
