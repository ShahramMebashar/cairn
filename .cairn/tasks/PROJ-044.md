---
id: PROJ-044
title: Dogfood and document observable sessions
status: done
priority: high
labels: [docs, sessions, dogfood]
parent: PROJ-040
deps: [PROJ-043]
checks:
  - desc: repository checks pass
    cmd: make check
    result: pass
  - desc: end-to-end session flow verified
    type: manual
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:33:23Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T19:08:43Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T19:08:48Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T19:08:56Z', did: began session ses_823084ee354c7075652c9fbe}
  - {who: 'agent:codex', at: '2026-06-22T19:09:32Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T19:10:11Z', did: attested, text: check 1 pass}
  - {who: 'agent:codex', at: '2026-06-22T19:10:11Z', did: note, text: 'Dogfooded identity → begin → heartbeat → finish with agent:codex. Verified durable session YAML, ephemeral live state, SSE-driven Active and Awaiting review counts, task-row signal, and live/final detail summaries. The run exposed and fixed cross-projection cache invalidation; recorded as friction #7.'}
  - {who: 'agent:codex', at: '2026-06-22T19:10:19Z', did: finished session ses_823084ee354c7075652c9fbe, text: 'Documented the observable session lifecycle and identity handshake in generated workflow and user guides, then dogfooded two real agent:codex sessions through HTTP, files, SSE, and UI. Repository checks, frontend build/lint, and browser verification pass; one stale-detail invalidation bug was found and fixed.'}
  - {who: 'agent:codex', at: '2026-06-22T19:10:24Z', did: transitioned to done}
assignee: agent:codex
active_attempt: att_823084ee354c7075652c9fbe
---
Make the first session slice usable by real Codex, Claude, and Cursor connections.

## Scope
- Update generated workflow and connection docs with identity/begin/heartbeat/finish conventions.
- Exercise a real session through the protocol and record friction.
- Verify the complete MCP → files → SSE → UI path.

## Acceptance
- A fresh agent can discover its bound identity and complete one observable session without hidden setup knowledge.
