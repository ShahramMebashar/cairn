---
id: PROJ-025
title: Board view (Kanban) with drag-and-drop
status: done
labels: [frontend, board]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: smooth + feature-rich (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T13:08:16Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T13:16:05Z', did: claimed}
  - {who: 'human:web', at: '2026-06-22T13:16:18Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:16:36Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:25:34Z', did: note, text: 'Kanban board shipped. New view `#/<slug>/board` (BoardView) via @dnd-kit: columns = config states, cards via useSortable, DragOverlay for a smooth lift. Cards show priority/id/title/labels/deps/checks/assignee; click opens task. Drag within a column reorders (midpoint rank → silent useReorder, persisted, NO provenance); drag across columns transitions (gate-aware: optimistic move, transition then reorder, refusal toasts + refetch snaps back). Reuses search + facets (extracted matches()/FILTER_LABEL→lib/filter.ts, Facet→components/Facet.tsx, also adopted by the list). Sidebar Board nav (SquareKanban) + ⌘K "Board" command. Verified via Playwright: board renders all columns/cards; dragged PROJ-004 above PROJ-003, persisted across reload, PROJ-004 file gained `rank: !!float 2` with provenance unchanged (7 entries). make check + pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T13:25:38Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T13:25:47Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T13:25:47Z', did: transitioned to done}
  - {who: 'agent:claude', at: '2026-06-22T13:34:02Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-22T13:34:09Z', did: transitioned to done}
assignee: agent:claude
---
Kanban board: columns per status, dnd-kit drag within + across columns.

- routing `#/<slug>/board` + sidebar Board nav + Cmd-K command
- extract `matches()`/FILTER_LABEL to lib/filter.ts and Facet to components/Facet.tsx (shared)
- BoardView: DndContext + SortableContext per status column; cards sorted by effective rank (rank||idNum); DragOverlay
- card: priority/id/title/labels/checks/assignee; click opens task
- onDragEnd: midpoint rank; cross-column => transition (gate-aware, optimistic, revert+toast on refuse) then reorder; same column => reorder
- reuse search/facets + New task + SSE live updates</parameter>
<parameter name="deps">["PROJ-024"]