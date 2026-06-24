---
id: PROJ-045
title: Replace task counter with time-ordered conflict-free IDs
status: done
priority: high
labels: [backend, ids, git]
checks:
  - desc: go tests pass
    cmd: go test ./...
    cwd: .
    result: pass
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T16:27:02Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T16:27:10Z', did: began session ses_fe1f32ee5d98a71fdbd02cef}
  - {who: 'agent:claude', at: '2026-06-24T16:38:06Z', did: note, text: 'Done. New ids are minted in internal/store/id.go (mintTaskID): "<prefix>-<10-char base32 UnixMilli><6-char crypto/rand tail>", lowercase Crockford base32 so lexical order == chronological. Create (store.go) now mints + does an existence-check retry and no longer writes config.yaml, so concurrent creators in separate clones never collide on id or filename. config.NewID removed; Counter field kept (deprecated) for lossless parse. Web effectiveRank (filter.ts) decodes the time prefix instead of the numeric suffix; legacy PROJ-0NN ids still decode to their small ints and sort before new ids. Note: create now emits SSE task-changed (single file) instead of tasks-changed; the web invalidates the task list on every event so the board still updates — updated the SSE test to cover the list path via two concurrent creates. Docs updated: SPEC §2/§3/§5/§8/§9, task-files.md, mcp-tools.md, WORKFLOW.md, FRICTION.md #8. Verified: go test ./... + web pnpm build both green; gofmt + go vet clean. Caveat: the cairn MCP server bound to this session is the pre-change binary, so its own create still used the old counter (it minted this very id, PROJ-045) — rebuild + reconnect to mint time-ordered ids live.'}
  - {who: 'agent:claude', at: '2026-06-24T16:38:18Z', did: finished session ses_fe1f32ee5d98a71fdbd02cef, text: 'Replaced the sequential counter with time-ordered, collision-free task ids. mintTaskID (internal/store/id.go) emits "<prefix>-<base32 UnixMilli><crypto/rand tail>"; Create no longer writes config.yaml and guards uniqueness with an existence-check retry, so two clones creating tasks offline never conflict on the id, the counter line, or the filename. config.NewID removed (Counter kept, deprecated). Web effectiveRank decodes the time prefix; legacy PROJ-0NN ids still sort first. SSE create signal is now task-changed (the web invalidates the list on any event); SSE test updated to cover the list path. Docs/spec/FRICTION updated. go test ./... and web pnpm build both pass; gofmt + go vet clean. Review focus: id length/entropy (10+6 base32), the effectiveRank decode for mixed old/new ids, and the SSE event-semantics change. Heads-up: the session''s cairn binary predates this change, so rebuild + reconnect before the new ids mint live.'}
  - {who: 'human:shaho', at: '2026-06-24T16:39:00Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T16:40:37Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T16:40:45Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_fe1f32ee5d98a71fdbd02cef
---
Sequential counter IDs (`prefix`+`counter` in `config.yaml`) cause unavoidable git conflicts when two clones create tasks offline: both bump the counter and both write the same `PROJ-NNN.md`. Replace with a time-ordered, collision-resistant ID minted at create; stop writing the counter.

## Scope
- New minter `internal/store/id.go`: `mintTaskID(prefix, at)` = 10-char base32 of `UnixMilli` + short `crypto/rand` tail (lexical == chronological). Reuse rand pattern from `internal/session/doc.go`.
- `internal/store/store.go` `Create`: use minter, drop the `config.Save` counter write, add an existence-check retry so a create never overwrites.
- `internal/config/config.go`: remove `NewID`, deprecate (keep) the `Counter` field for lossless load.
- Web `web/src/lib/filter.ts` `effectiveRank`: stop parsing numeric ID suffix; fall back to lexical id / created time.
- Tests: rewrite `TestCreateMintsIDAndIncrementsCounter`, update config test, add `mintTaskID` test.
- Docs: `SPEC.md`, `docs/task-files.md`, `docs/mcp-tools.md`, `tools.go` schema example, `.cairn/WORKFLOW.md` (monotonic wording).

## Acceptance
- Two offline clones each create a task → merge with no conflict in `config.yaml` and two distinct task files.
- Existing `PROJ-001..044` untouched.