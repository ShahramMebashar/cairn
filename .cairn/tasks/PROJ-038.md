---
id: PROJ-038
title: Specify agent sessions and evidence-gated completion
status: done
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
  - {who: 'agent:codex', at: '2026-06-22T18:07:25Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:07:30Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:07:34Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T18:08:04Z', did: note, text: 'Kept session tracking agent-neutral and split durable session/evidence files from gitignored heartbeat/log state. Also recorded the identity mismatch that created PROJ-037 under agent:claude; PROJ-038 is the correctly attributed replacement and the proposed begin contract now treats actor/client mismatch as a first-class failure.'}
  - {who: 'agent:codex', at: '2026-06-22T18:08:46Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T18:08:53Z', did: note, text: 'Completed the proposed v1 design in docs/superpowers/specs/2026-06-22-agent-sessions-and-trust-loop-design.md and logged the MCP actor mismatch in .cairn/FRICTION.md. The design defines durable sessions, ephemeral heartbeats, typed verified evidence, workspace-bound check freshness, completion facts, MCP/HTTP contracts, UI states, failure recovery, and phased delivery. Added an expected_actor handshake so mismatched connections fail before writing. Verified with task checks (make check and document existence) plus git diff --check.'}
  - {who: 'agent:codex', at: '2026-06-22T18:08:58Z', did: transitioned to done}
assignee: agent:codex
---
Define the durable session protocol and trust gates that make agent work observable, recoverable, and verifiable.

## Scope
- Session lifecycle, storage, MCP/HTTP contracts, evidence, heartbeat-derived state, and UI surfaces.
- Git-aware check freshness and completion semantics.
- Preserve Cairn as an agent-neutral control plane; agent launching stays out of core.

## Acceptance
- Design identifies invariants, failure handling, migration, phased implementation, and rejected alternatives.
