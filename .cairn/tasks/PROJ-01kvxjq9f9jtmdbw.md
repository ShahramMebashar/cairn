---
id: PROJ-01kvxjq9f9jtmdbw
title: Edit & delete tasks, subtasks, and notes (web + MCP)
status: done
priority: high
labels: [feature, backend, web, mcp]
provenance:
  - {who: 'agent:claude', at: '2026-06-24T19:46:42Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T19:46:49Z', did: began session ses_a6232e1ffa1dd54251e60ef0}
  - {who: 'agent:claude', at: '2026-06-24T19:50:21Z', did: note, text: 'Backend done (builds + existing tests pass). Decisions: notes get a stable `n_…` id stamped only on `did=="note"` entries so system provenance stays byte-identical; note edit/delete address by id with a 0-based index fallback for legacy id-less notes (seq.Content[i] stays 1:1 with d.Provenance[i]). Delete-block rule lives in pure `task.ValidateDeletable` (children via parent, dependents via deps) so it''s testable without a store; `store.DeleteTask` scans + unlinks under the write lock to avoid a TOCTOU child-create race. Task edit reuses the `update` path (one `updated` provenance entry); checks replace the whole list, carrying `result` forward on retained checks. Next: backend tests for the new paths, then web UI.'}
  - {who: 'agent:claude', at: '2026-06-24T20:01:24Z', did: note, text: 'Done. Backend: `store` gains `Provenance.ID/EditedAt` (note-only id via `mintNoteID`), `SetTitle/SetBody/SetChecks`, `DeleteTask`, `EditNote/DeleteNote` (id-or-index addressing, splice keeps node↔slice aligned); `task.ValidateDeletable` blocks delete on children/dependents; `mcp.Service` extends `UpdateFields` + adds `Delete/EditNote/DeleteNote`; MCP tools `delete`,`edit_note`,`delete_note` + `update` now does title/body/checks; HTTP `DELETE /api/tasks/{id}`, `PATCH|DELETE /api/tasks/{id}/notes/{note}` (?index= legacy fallback), 422/404 mapping. Web: api/queries hooks, TaskDetail inline title/body/checks editors + ••• delete (AlertDialog), note edit/delete in ActivityEntry with "(edited)" marker, delete on SubTasks/TaskRow/Board card; added shadcn alert-dialog + ConfirmDeleteDialog. Verified: `go test ./...` (new store/task/mcp/server tests), `go vet`, `gofmt`, `pnpm build`, and a live HTTP smoke (create→edit→delete-blocked-by-child→note edit/delete→edit-system-422→leaf deletes). Note: the running MCP server is the pre-change binary, so the new `delete`/`edit_note`/`delete_note` MCP tools require a `cairn serve` restart to appear.'}
  - {who: 'agent:claude', at: '2026-06-24T20:01:36Z', did: finished session ses_a6232e1ffa1dd54251e60ef0, text: 'Implemented edit/delete for tasks, subtasks, and activity notes across store→service→MCP→HTTP→web. Edit task = title+body+checks (reuses `update`); delete task is hard-delete blocked by children/dependents (`task.ValidateDeletable`); notes get a stable id and can be edited (in place + `editedAt`) or deleted by anyone, with a legacy index fallback. New MCP tools: `delete`, `edit_note`, `delete_note` (require `cairn serve` restart to appear). New HTTP: `DELETE /api/tasks/{id}`, `PATCH|DELETE /api/tasks/{id}/notes/{note}`. Web: inline title/body/checks editors, ••• delete with AlertDialog on TaskDetail/Board/list/sub-tasks, note edit/delete in ActivityEntry. Verified via go test ./... (added store/task/mcp/server tests), go vet, gofmt, pnpm build, and a live HTTP smoke covering create→edit→blocked-delete→note edit/delete→system-entry-422→leaf deletes. To review: run `go test ./...` and `pnpm build`; restart `cairn serve` to expose the new MCP tools.'}
  - {who: 'human:shaho', at: '2026-06-24T20:20:55Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T20:20:55Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_a6232e1ffa1dd54251e60ef0
rank: !!float 891164924159
---
Add full edit/delete for tasks (= subtasks, which are just tasks with a `parent`) and edit/delete for activity notes, exposed to humans (web) and agents (MCP). Plan: `~/.claude/plans/wewant-the-ability-to-shimmering-origami.md`.

## Scope
- **Edit task**: title + body + checks via the existing `update` path (priority/labels/parent already work). `internal/store/store.go` (`SetTitle`/`SetBody`/`SetChecks`), `internal/mcp/service.go` (`UpdateFields`), `tools.go` (`updateIn`/`checkIn`), `server.go` (`updateReq`).
- **Delete task**: hard delete, **blocked** if the task has children (`parent==id`) or dependents (`deps∋id`). `task.ValidateDeletable` + `store.DeleteTask` + `svc.Delete` + `delete` MCP tool + `DELETE /api/tasks/{id}`.
- **Edit/delete note**: only `did=="note"` entries; anyone may edit/delete any. Edit in place + `editedAt`. Stable note `id` added to `Provenance` (index fallback for legacy). `EditNote`/`DeleteNote` across store→service→MCP (`edit_note`/`delete_note`)→HTTP (`PATCH`/`DELETE /api/tasks/{id}/notes/{note}`).
- **Web UI**: inline title edit, body editor, checks editor, delete (•••+AlertDialog) on TaskDetail/Board/SubTasks, note edit/delete in `ActivityEntry`.

## Acceptance
- `go test ./...` + new store/task/server tests; `go build ./...`; `pnpm build`.
- MCP smoke: update title/body/checks, delete (blocked by child), note→edit_note→delete_note.

Decisions: delete is block-not-cascade; notes editable by anyone; edit-in-place + editedAt; only note entries editable.