---
id: PROJ-2apxkyvcqd
title: Render Code Context on task detail
status: done
priority: high
labels: [web, git, sessions, ui]
parent: PROJ-2apxk5hz23
deps: [PROJ-2apxkpxre4]
checks:
  - desc: Web build passes
    cmd: cd web && pnpm build
    timeout: 300
    result: pass
  - desc: Go checks pass
    cmd: make check
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-25T18:44:14Z', did: created}
  - {who: 'agent:codex', at: '2026-06-25T19:09:24Z', did: began session ses_65a3044eee053f620087209e}
  - {id: n_qs3n2ycz, who: 'agent:codex', at: '2026-06-25T19:09:34Z', did: note, text: 'UI decision: render Code Context inside `TaskDetail` next to the existing session timeline. The panel stays compact (branch/refs, warnings, changed files, commits, uncommitted files) and avoids becoming a diff viewer.'}
  - {who: 'agent:codex', at: '2026-06-25T19:09:39Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:11:35Z', did: finished session ses_65a3044eee053f620087209e, text: 'Rendered Code Context in `TaskDetail`: added typed API/query support, task-level context fetching, and a compact panel showing branch, refs, warnings, changed files, commits, and uncommitted files. Verified with `cd web && pnpm build`, `make check`, live endpoint curl, and a browser DOM check showing populated Code Context.'}
  - {who: 'agent:codex', at: '2026-06-25T19:11:43Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:11:43Z', did: transitioned to done}
assignee: agent:codex
active_attempt: att_65a3044eee053f620087209e
---
Show Git evidence in the web UI where humans review agent work.

## Scope
- Add a Code Context panel to `TaskDetail`/session detail.
- Show branch, start/end/current commits, changed files, commits, dirty state, and warnings.
- Keep the panel compact and Linear-like, reusing shadcn/ui components and design tokens.
- Include degraded states for missing Git, missing refs, and no session anchors.

## Acceptance
- Finished sessions display branch, commits, files changed, and review warnings.
- Active sessions display current working changes and dirty state.
- The UI remains readable in the existing task detail split layout.
- No raw colors or one-off component patterns are introduced.