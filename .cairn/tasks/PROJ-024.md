---
id: PROJ-024
title: 'Backend: rank field + provenance-free reorder'
status: done
labels: [backend, board]
checks:
  - desc: go check passes
    cmd: make check
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T13:07:52Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T13:08:45Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T13:08:56Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:15:39Z', did: note, text: 'Added Task.Rank (float64, optional, never gates) through task→store→mcp→server→api. store: floatNode, parse, write-at-create, SetRank (clears at 0). mcp.Reorder saves WITHOUT provenance (cosmetic) + `reorder` verb; rank on taskOut/view/list. server POST /api/tasks/{id}/reorder + rank on taskDTO. web: Task.rank, reorderTask, silent useReorder (invalidate on settle). Tests: rank round-trip+clear, Reorder adds no provenance, reorder endpoint. make check + pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T13:15:54Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T13:15:54Z', did: transitioned to done}
assignee: agent:claude
---
Persisted manual ordering for the Kanban board.

- `task.Task.Rank float64` (optional; 0=unset; never gates)
- store: parse `rank`, `floatNode`, write at create if set, `SetRank` setter (removes key at 0)
- mcp: `Reorder(id, rank)` — saves WITHOUT provenance (reorder is cosmetic); `reorder` verb
- server: `POST /api/tasks/{id}/reorder {rank}`; `rank` on taskDTO + mcp taskOut/view
- api.ts/queries: `Task.rank`, `reorderTask`, silent `useReorder`
- tests: rank round-trip+clear, Reorder adds no provenance, server endpoint</parameter>
<parameter name="priority">high