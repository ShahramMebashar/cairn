---
id: PROJ-007
title: 'Web: check output panel in TaskDetail'
status: done
deps: [PROJ-004, PROJ-006]
checks:
  - desc: typecheck + build
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: check output renders and refreshes after a run; matches Linear aesthetic
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T19:48:09Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:21:05Z', did: note, text: 'Code built ahead of the deps gate (PROJ-006 still in_review, so this task stays backlog/not-ready and wasn''t transitioned). Done: TaskDetail.tsx CheckRow rewritten — cmd checks with a captured run are an expandable Collapsible showing the latest run (matched by cmd) with exit/duration/timedout header + output in a ScrollArea (monospace); manual+pending checks get inline Attest pass/fail buttons. Added useRuns + useAttest hooks (queries.ts), api.getRuns/attestTask + Run type. Live refresh via PROJ-006''s SSE invalidation of the runs key. pnpm build passes. To close 006 then 007: rebuild+reconnect cairn (so attest verb + SSE/runs endpoints are live), run the app, verify both manual checks, attest.'}
  - {who: 'agent:claude', at: '2026-06-21T20:29:47Z', did: claimed}
  - {who: 'human:web', at: '2026-06-21T20:29:55Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:29:59Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:30:43Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-21T20:32:02Z', did: attested, text: check 1 pass}
  - {who: 'human:web', at: '2026-06-21T20:32:11Z', did: transitioned to done}
assignee: agent:claude
---
Show captured check output per check.

- Add the runs query hook consuming `GET /api/tasks/{id}/runs`.
- In `TaskDetail.tsx`, each cmd check gets an expandable panel (shadcn `Collapsible` + `ScrollArea`, monospace) showing the latest run matched by `cmd`, plus header (exit/duration/timedout).
- Add shadcn components via `pnpm dlx shadcn@latest add` — no hand-rolling; style with design tokens only.
- Refreshes live via the SSE-invalidated runs query.

See spec §B (UI).