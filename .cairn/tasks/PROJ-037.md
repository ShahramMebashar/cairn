---
id: PROJ-037
title: Specify agent sessions and evidence-gated completion
status: canceled
priority: urgent
labels: [design, sessions, trust]
checks:
  - desc: repository checks pass
    cmd: make check
    result: pass
  - desc: design document exists
    cmd: test -f docs/superpowers/specs/2026-06-22-agent-sessions-and-trust-loop-design.md
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T18:02:46Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T18:03:14Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T18:03:18Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T18:07:40Z', did: note, text: 'Created and started through an MCP connection configured as agent:claude while the active worker was Codex. Superseded by PROJ-038 so subsequent provenance is correctly attributed; no implementation work is retained on this task.'}
  - {who: 'agent:codex', at: '2026-06-22T18:07:46Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T18:07:46Z', did: transitioned to canceled}
assignee: agent:claude
---
Define the durable session protocol and trust gates that make agent work observable, recoverable, and verifiable.

## Scope
- Session lifecycle, storage, MCP/HTTP contracts, evidence, heartbeat-derived state, and UI surfaces.
- Git-aware check freshness and completion semantics.
- Preserve Cairn as an agent-neutral control plane; agent launching stays out of core.

## Acceptance
- Design identifies invariants, failure handling, migration, phased implementation, and rejected alternatives.
