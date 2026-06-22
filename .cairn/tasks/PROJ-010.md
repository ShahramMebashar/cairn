---
id: PROJ-010
title: Optimistic concurrency on task Save (no last-write-wins)
status: done
checks:
  - desc: concurrent Save conflicts instead of clobbering
    cmd: go test ./internal/store ./internal/mcp
    cwd: .
    result: pass
  - desc: gofmt + vet + test
    cmd: make check
    cwd: .
    result: pass
  - desc: retry-vs-reject policy reviewed by a human
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:32:37Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:37:13Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:39:30Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:39:39Z', did: note, text: 'Implemented reject-with-409 (per decision). store: Doc gains a `version` (sha256 of file bytes captured at Get); save() refuses with ErrConflict if the on-disk file''s hash changed since read (or was deleted), then advances version after a successful write so the same Doc can be re-saved. Create has no version → skipped. server writeErr maps store.ErrConflict→409 (alongside ErrAlreadyClaimed). Tests: TestSaveConflictsOnStaleDoc, TestSaveSucceedsAfterReread. Existing single-Get→Save flows unaffected. Caveat: tiny TOCTOU window remains between the version re-read and the rename — narrows the real seconds-apart race, not a hard cross-process lock. Manual check (policy review) satisfied by the reject-409 decision.'}
  - {who: 'agent:claude', at: '2026-06-21T20:39:55Z', did: ran checks}
  - {who: 'human:web', at: '2026-06-21T20:40:05Z', did: attested, text: check 2 pass}
  - {who: 'agent:claude', at: '2026-06-21T20:40:10Z', did: transitioned to done}
assignee: agent:claude
---
Concurrent writers clobber state: when the web server (human:web) and an MCP agent (agent:claude) act on the same task within a moment, Save load-mutate-writes the whole doc with no version check, so one overwrites the other. Observed PROJ-007's in_progress reverting to backlog after a near-simultaneous ran-checks (friction #5 in .cairn/WORKFLOW.md).

- Add an optimistic guard in internal/store: capture the file's mtime/size (or a content hash) at Get, and on Save reject if it changed underneath (ErrConflict), so the caller re-reads instead of clobbering.
- Decide caller behavior: surface ErrConflict (409) over HTTP/MCP, or auto-retry the mutation against the fresh doc.
- Tests: two services mutating the same id; second Save conflicts rather than silently reverting.

NOTE: design-heavier — confirm the retry-vs-reject policy before implementing.