---
id: PROJ-040
title: Ship observable agent sessions
status: done
priority: urgent
labels: [epic, sessions]
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:32:30Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:32:38Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:32:44Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T19:11:44Z', did: note, text: 'Shipped Slice 1 end to end: repository-wide flock serialization, durable/live session stores, identity-safe MCP and HTTP lifecycle, SSE supervision, high-signal UI, generated workflow/docs, and real agent:codex dogfood. Verified with make check, frontend build/lint, race tests for session/store/MCP/server, and dark/light browser review.'}
  - {who: 'agent:codex', at: '2026-06-22T19:11:51Z', did: transitioned to done}
assignee: agent:codex
---
Deliver Slice 1 from the session control-plane design as an end-to-end, dogfoodable product loop.

## Scope
- Cross-process write serialization and durable/live session storage.
- Session lifecycle through MCP and HTTP with realtime updates.
- Active, stalled, and awaiting-review supervision UI.
- Agent setup guidance and dogfood verification.

## Design
- See docs/superpowers/specs/2026-06-22-agent-sessions-and-trust-loop-design.md.
