---
id: PROJ-022
title: Notification inbox (SSE-derived)
status: done
deps: [PROJ-017]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: notifications useful (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:33:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T08:06:30Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-22T08:07:33Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-22T09:27:34Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T09:27:48Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T09:52:31Z', did: note, text: 'In-app inbox: `lib/notifications.ts` diffs the SSE-refreshed tasks query vs the previous snapshot — ready (false→true), blocked (true→false), check-failed (fail count up), assigned-to-me (assignee===status.actor). Capped (50) localStorage list with read state. `NotificationBell` in the sidebar: unread badge, click→open+mark-read, auto-mark-read on open, clear. Verified: reopening PROJ-017 mid-session surfaced 3 "blocked" notifications for its dependents via SSE.'}
  - {who: 'agent:claude', at: '2026-06-22T09:52:40Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T09:52:54Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T09:52:54Z', did: transitioned to done}
priority: urgent
assignee: agent:claude
---
Bell/inbox in sidebar. Diff SSE-refreshed tasks vs previous snapshot: ready (false->true), check failed, assigned to me (assignee===status.actor), blocked (ready true->false). Capped localStorage list, unread badge, click->open, mark read. No OS notifications.