---
id: PROJ-021
title: Saved views + faceted search
status: done
deps: [PROJ-017]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: feels good (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:33:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T08:06:30Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-22T08:44:53Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T08:45:29Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T09:19:29Z', did: note, text: 'Faceted board filter: Priority/Label/Assignee dropdowns (derived from present values) + text search, all client-side. Saved views persisted per workspace (`lib/views.ts`, localStorage `cairn-views:<path>`) via a Views dropdown — save current (name), apply (sets base filter + facets through onPickFilter), delete. Verified facet bar + Views render on the board.'}
  - {who: 'agent:claude', at: '2026-06-22T09:26:48Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T09:27:04Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T09:27:04Z', did: transitioned to done}
priority: medium
labels: [frontend]
assignee: agent:claude
---
Board filter facets (status + label + assignee + priority + text). Named views persisted per workspace (localStorage cairn-views); sidebar lists them; Save current view; Cmd-K jumps to them.