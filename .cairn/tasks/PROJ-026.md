---
id: PROJ-026
title: 'Fix: cannot drop a card into an empty board column'
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: empty-column drop works (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T13:31:58Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T13:32:03Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T13:32:07Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:33:20Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-22T13:33:41Z', did: note, text: 'Fixed via a custom collisionDetection in BoardView: pointerWithin first (resolves an empty column under the cursor, which closestCorners missed), falling back to rectIntersection for the keyboard sensor. Verified with Playwright: dragging PROJ-026 into the empty "In review" column transitioned it in_progress→in_review (toast confirmed, card now in that column).'}
  - {who: 'agent:claude', at: '2026-06-22T13:33:45Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T13:33:51Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T13:33:53Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-22T13:33:59Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T13:33:59Z', did: transitioned to done}
  - {who: 'agent:claude', at: '2026-06-22T13:34:11Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-22T13:34:46Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T13:35:42Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-22T13:35:44Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T13:40:11Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-22T13:40:14Z', did: transitioned to in_progress}
  - {who: 'human:web', at: '2026-06-22T13:40:15Z', did: transitioned to done}
assignee: agent:claude
---
Dragging a card into an empty status column does nothing — closestCorners doesn't resolve an empty droppable under the pointer as `over`, so no transition fires.

Fix: custom collision detection in BoardView — pointerWithin first (detects the empty column under the cursor), fall back to rectIntersection. Verify dropping into an empty column transitions the task.</body>
<parameter name="labels">["frontend", "board", "bug"]