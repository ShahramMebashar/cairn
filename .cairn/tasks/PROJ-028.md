---
id: PROJ-028
title: 'Board drag: insertion indicator + fix drop blink'
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: indicator + no blink (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T13:44:17Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T13:44:23Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T13:44:31Z', did: transitioned to in_progress}
  - {who: 'human:web', at: '2026-06-22T13:46:27Z', did: ran checks}
  - {who: 'human:web', at: '2026-06-22T13:46:30Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-22T13:47:05Z', did: note, text: 'Removed activeId from the cols-rebuild effect deps (kept the in-effect guard) so a drop no longer rebuilds from stale server data — fixes the snap-back blink; optimistic order holds until the mutation settles. Added a DropIndicator (3px brand line + centered ring) rendered in place of the active card while dragging, so the insertion point is clearly marked Linear-style. User confirmed it feels nice. pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T13:47:17Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T13:47:33Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T13:47:56Z', did: transitioned to canceled}
  - {who: 'human:web', at: '2026-06-22T13:47:58Z', did: transitioned to done}
assignee: agent:claude
rank: !!float 1
---
Two refinements after PROJ-027.

1. Drop blink: the cols-rebuild effect lists `activeId` in its deps, so the instant the drag ends it rebuilds from stale (pre-mutation) server data → the card flashes back to its old slot before the refetch moves it. Fix: drop `activeId` from the deps (keep the in-effect guard) so cols rebuild only when server data actually changes; the optimistic cols hold until then.

2. Insertion indicator: while dragging, render a 3px blue line with a centered ring in the drop slot (the active card's optimistic position) instead of an empty gap — clear Linear-style "drops here". Implemented by rendering a DropIndicator in SortableCard when isDragging; neighbors animate to make room.</body>
<parameter name="labels">["frontend", "board", "polish"]