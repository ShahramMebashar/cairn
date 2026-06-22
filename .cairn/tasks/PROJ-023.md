---
id: PROJ-023
title: Address code-review findings (power-pack)
status: done
labels: [review, backend, frontend]
provenance:
  - {who: 'agent:claude', at: '2026-06-22T11:59:07Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T11:59:42Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T11:59:50Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T12:50:37Z', did: note, text: 'Addressed review findings. Backend: dtoFromDoc now sets updatedAt (single-task responses); writeErr maps ParentMissing/ParentCycle/DanglingDep/Cycle/InvalidPriority → 422; mcp taskOut + list + view carry updatedAt; Update is a no-op (no provenance/save) when no fields change; priority validated against task.Priorities in Create+Update (task.ValidPriority/ErrInvalidPriority). Tests added: TestUpdateFields (mcp), TestUpdateValidationAndGetUpdatedAt (server). Frontend: notifications reset snapshot+inbox on workspace change, unique keys (seq counter), persist outside the state updater; Board resets keyboard cursor on filter change and ignores nav keys when a listbox/menu/dialog is open; NotificationBell clears its mark-read timeout on close/unmount; StatusSection reacts to defaultOpen; AppSidebar sets aria-current. Dismissed the reviewer''s "remove nested <Command>" — this CommandDialog variant doesn''t render Command, so the wrapper is required (it was the earlier crash fix). make check + pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T12:50:58Z', did: transitioned to done}
assignee: agent:claude
---
Fixes from the two-reviewer pass over PROJ-017…022.

**Backend**
- `dtoFromDoc` must set `updatedAt` (single-task responses returned "").
- `writeErr`: map `ErrParentMissing/ErrParentCycle/ErrDanglingDep/ErrCycle` → 422 (were 500).
- mcp `taskOut` + `list` verb: include `updatedAt`.
- `Update`: skip provenance/save on a no-op (all fields nil).
- Validate `priority` against the allowed set (create + update).
- Tests: `Update` service test, parent-missing → 422, `updatedAt` on GET.

**Frontend**
- `notifications`: reset `prev` on workspace (path) change; unique keys; persist outside the state updater.
- `Board`: reset keyboard selection on filter change; bail keyboard nav when a Radix listbox/menu is open.
- `NotificationBell`: clear the mark-read timeout on close/unmount.
- `StatusSection`: react to `defaultOpen` changes.
- `AppSidebar`: `aria-current` on the active nav item.

Note: the reviewer's "remove nested <Command>" is a false positive — this CommandDialog variant doesn't render Command, so the wrapper is required.</body>
<parameter name="priority">high