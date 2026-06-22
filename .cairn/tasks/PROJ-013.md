---
id: PROJ-013
title: 'List API: surface updatedAt (last activity timestamp)'
status: done
checks:
  - desc: go check passes
    cmd: make check
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:45:37Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:51:30Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:51:30Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:55:19Z', did: note, text: 'store.ListDocs() returns parsed Docs (with provenance); List builds on it. mcp TaskView gains UpdatedAt (newest provenance via lastActivity). server taskDTO.updatedAt set in handleList; api.ts Task.updatedAt. Test: TestListIncludesUpdatedAt. make check green.'}
  - {who: 'agent:claude', at: '2026-06-21T20:55:20Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:55:20Z', did: transitioned to done}
assignee: agent:claude
---
Backend so the list can show relative time.

- store.ListDocs() keeps Provenance; List builds on it
- mcp TaskView.UpdatedAt from newest provenance
- server taskDTO.updatedAt in handleList
- api.ts Task.updatedAt
- test: updatedAt reflects newest provenance