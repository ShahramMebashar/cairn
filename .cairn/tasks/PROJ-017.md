---
id: PROJ-017
title: 'Data model: labels, priority, parent + update verb'
status: done
checks:
  - desc: go check passes
    cmd: make check
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:32:53Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T07:33:39Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T07:33:39Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T07:44:37Z', did: note, text: 'Added optional Labels/Priority/Parent across task->store->mcp->server->api. `task.ValidateParents` (exists + no ancestor cycle), validated in ListDocs alongside deps. store: Draft struct for Create, SetPriority/SetLabels/SetParent, removeKey. mcp: Create(Draft), Update + `update` verb. server: taskDTO/createReq fields, POST /update, `actor` in /status. parent never gates. Tests: ValidateParents, store org round-trip+clear, server update+actor. make check + pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T07:44:38Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T07:44:38Z', did: transitioned to done}
  - {who: 'agent:claude', at: '2026-06-22T09:31:11Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T09:31:14Z', did: transitioned to done}
assignee: agent:claude
---
Additive optional frontmatter fields wired store->mcp->server->api.

- task.Task: Labels []string, Priority string, Parent string; ValidateParents (exists + no ancestor cycle)
- store: parse + node builders + SetPriority/SetLabels/SetParent
- mcp: CreateParams, Update(id,fields); add `update` verb; validate parents in List
- server: taskDTO + createReq fields; POST /api/tasks/{id}/update; add actor to /api/status
- api.ts/queries.ts: Task fields, updateTask, Status.actor
- parent never gates; deps remain the only gate