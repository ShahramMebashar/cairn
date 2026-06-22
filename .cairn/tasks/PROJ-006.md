---
id: PROJ-006
title: 'Web: EventSource client + query invalidation'
status: done
deps: [PROJ-005]
checks:
  - desc: typecheck + build
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: board updates live when another process changes a task
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T19:48:02Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:04:33Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:04:38Z', did: transitioned to in_progress}
  - {who: 'human:web', at: '2026-06-21T20:05:43Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:06:05Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:06:15Z', did: note, text: 'Code done + typecheck passes. Added useTaskEvents(path) in queries.ts (one EventSource → invalidates tasks list always, plus task+runs keys on task-changed{id}); mounted in App Workspace where path is active+initialized; browser auto-reconnects. Also added useRuns hook + api.getRuns/Run type (for PROJ-007). BLOCKED from done: the manual check ''board updates live'' is pending and cairn has NO attestation verb (SetCheckResult is only called for cmd checks) — so it can never pass, and done/canceled are both gated. Moving to in_review pending a decision on attestation.'}
  - {who: 'agent:claude', at: '2026-06-21T20:06:25Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-21T20:22:07Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-21T20:29:39Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T13:04:14Z', did: note, text: "```\nconsole.log(\"Hello, World!\");\n```"}
assignee: agent:claude
---
Wire the SSE stream into React Query.

- Open one `EventSource` for the active project root; close/reopen on root change.
- On `task-changed`: invalidate that task's key + its runs key. On `tasks-changed`: invalidate the tasks-list key.
- Reuse existing query keys in `web/src/lib/queries.ts`; add the runs query hook.
- Rely on `EventSource` auto-reconnect.

See spec §A (Client).