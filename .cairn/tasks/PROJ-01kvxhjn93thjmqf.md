---
id: PROJ-01kvxhjn93thjmqf
title: Show "Running checks…" loading on Kanban cards during gated moves
status: done
priority: medium
labels: [frontend, ux]
checks:
  - desc: Web typecheck/build
    cmd: pnpm -C web build
    timeout: 240
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T19:26:41Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T19:26:48Z', did: began session ses_c29b59a03f9186ebd6ed051a}
  - {who: 'agent:claude', at: '2026-06-24T19:27:51Z', did: note, text: 'Implemented in BoardView.tsx only. Added Loader2 import; derived pendingId (transition.isPending && transition.variables?.id) and pendingLabel (gated→"Running checks…" via status.closed/review, else "Updating…"); threaded busy/busyLabel through Column→SortableCard→Card. Card gained a relative root + absolutely-positioned overlay (bg-panel/70 backdrop-blur, Loader2 spinner + label, z-10) and aria-busy. DragOverlay card unaffected (busy defaults off). Immediate-drop behavior unchanged — local cols state already holds the card in the target column until the mutation settles. pnpm build clean. Caveat: in_review labeling needs the rebuilt bin/cairn restarted to serve status.review; until then done shows "Running checks…", in_review shows "Updating…".'}
  - {who: 'agent:claude', at: '2026-06-24T19:28:00Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T19:28:10Z', did: finished session ses_c29b59a03f9186ebd6ed051a, text: "Kanban cards now show a loading overlay while a dragged card's transition is in flight, mirroring the task-detail control. Single file: web/src/pages/BoardView.tsx.\n\n- Derived the in-flight card from the shared useTransition (pendingId = isPending && variables.id; dnd-kit = one drag at a time) and a pendingLabel (\"Running checks…\" for gated closed/review targets via status.closed/status.review, else \"Updating…\").\n- Threaded busy/busyLabel through Column → SortableCard → Card.\n- Card got a relative root + absolutely-positioned overlay (translucent panel, backdrop blur, Loader2 spinner + label, z-10) and aria-busy. DragOverlay card unaffected (busy defaults off).\n- No change to drop behavior: the immediate drop already holds via local cols state until the mutation settles; on check failure the card snaps back and the existing error toast fires.\n\nVerified: pnpm build clean (check passed). Review focus: visual polish of the overlay; confirm in_review labels correctly after the dev server restarts with the rebuilt bin/cairn (serves status.review) — until then done shows \"Running checks…\", in_review shows \"Updating…\"."}
  - {who: 'human:shaho', at: '2026-06-24T19:30:18Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-24T19:30:24Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:30:24Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:30:29Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-24T19:30:34Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:30:34Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T19:33:25Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:33:25Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_c29b59a03f9186ebd6ed051a
rank: !!float 1
---
Board (`BoardView.tsx`) gives no feedback while a dragged card's gated transition (done/in_review) runs checks server-side (~seconds). Mirror the task-detail behavior: keep the immediate drop, add a per-card loading overlay.

## Scope
- Frontend only: `web/src/pages/BoardView.tsx`.
- Derive busy card from shared `useTransition` (`transition.isPending` + `transition.variables?.id`); dnd-kit = one drag at a time.
- Thread `busy`/`busyLabel` through Column → SortableCard → Card; overlay with Loader2 + "Running checks…" (gated) / "Updating…".
- Gating from `status.closed`/`status.review` (matches TaskDetail.tsx:216).

## Acceptance
- Drag to done/in_review: card drops immediately, shows "Running checks…" overlay until checks settle; snaps back on fail (existing toast).
- DragOverlay card unaffected; `pnpm build` clean.