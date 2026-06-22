---
id: PROJ-014
title: 'Task list: search, hover actions, inline status, relative time'
status: done
deps: [PROJ-013]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: matches Linear aesthetic (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:45:53Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:56:00Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:56:00Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T21:01:13Z', did: note, text: 'Board search filter; TaskRow shows relative updated time (timeAgo), hover Claim, and inline status change via the status glyph menu (row is a div so nested menu/buttons are valid). Verified in light+dark.'}
  - {who: 'agent:claude', at: '2026-06-21T21:01:13Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-21T21:01:16Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T21:01:16Z', did: transitioned to done}
assignee: agent:claude
---
Make the list more capable (depends on updatedAt).

- Board: search input filtering by id/title
- TaskRow: relative updated time (timeAgo util)
- Hover quick-actions: claim + status menu
- Inline status change via the status glyph