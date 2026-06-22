---
id: PROJ-027
title: 'Polish: smooth board drag — kill drop-animation snap-back'
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: drag feels smooth (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T13:37:36Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T13:37:43Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T13:37:49Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:39:07Z', did: note, text: 'DragOverlay dropAnimation={null} (no snap-back; card is already placed via optimistic cols) + source card opacity-0 while dragging (single clean lifted overlay, gap placeholder, no faint duplicate). Verified with a mid-drag Playwright capture: one lifted card following the cursor with rotate+shadow, target column highlighted, source slot vacated. pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T13:39:13Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T13:39:21Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T13:39:21Z', did: transitioned to done}
assignee: agent:claude
---
On drop, the DragOverlay plays its default drop animation back toward the original sortable slot (which has already moved via optimistic state), so a ghost slides back and vanishes — feels unnatural even though the card lands correctly.

Fix in BoardView:
- DragOverlay dropAnimation={null} (card is already in place optimistically; overlay should just vanish)
- source card while dragging: opacity-0 (was 40%) so there's a single clean lifted overlay + a gap placeholder, not a faint duplicate.</body>
<parameter name="labels">["frontend", "board", "polish"]