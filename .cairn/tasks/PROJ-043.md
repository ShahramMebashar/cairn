---
id: PROJ-043
title: Build session supervision UI
status: done
priority: high
labels: [frontend, sessions]
parent: PROJ-040
deps: [PROJ-042]
checks:
  - desc: frontend builds
    cmd: pnpm build
    cwd: web
    timeout: 300
    result: pass
  - desc: frontend lint passes
    cmd: pnpm lint
    cwd: web
    timeout: 300
    result: pass
  - desc: session states verified in browser
    type: manual
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:33:15Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:52:25Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:52:30Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T19:05:39Z', did: began session ses_a4ff1850db19336c1e897330}
  - {who: 'agent:codex', at: '2026-06-22T19:06:56Z', did: note, text: 'Implemented session-aware Active/Stalled/Awaiting review views, live SSE invalidation, compact task-row state, and a detail timeline with heartbeat, progress, runtime identity, Git context, usage, and summary. Full build/lint pass; visually verified active filtering and dark/light layouts against a real agent:codex session.'}
  - {who: 'agent:codex', at: '2026-06-22T19:07:07Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T19:07:13Z', did: attested, text: check 2 pass}
  - {who: 'agent:codex', at: '2026-06-22T19:07:20Z', did: finished session ses_a4ff1850db19336c1e897330, text: 'Built and verified the session supervision UI: Active/Stalled/Awaiting review views with live counts, task-row execution signals, and a session timeline showing heartbeat, progress, actor/client/model, Git context, usage, and handoff summary. Frontend build and lint pass; dark/light browser review passed.'}
  - {who: 'agent:codex', at: '2026-06-22T19:08:02Z', did: transitioned to done}
assignee: agent:codex
active_attempt: att_a4ff1850db19336c1e897330
---
Expose observable sessions as a quiet, high-signal supervision experience.

## Scope
- Active, Stalled, and Awaiting review sidebar views and counts.
- Session status, progress, heartbeat age, actor/client/model, usage, and summary in task detail.
- Realtime query invalidation and useful empty/loading/error states.

## Acceptance
- Compose existing shadcn primitives and design tokens; verify dark/light layout and keyboard navigation.
