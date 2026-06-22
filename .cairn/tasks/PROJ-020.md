---
id: PROJ-020
title: Labels / priority / hierarchy UI
status: done
deps: [PROJ-017]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: matches Linear aesthetic (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:33:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T07:58:30Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T07:58:30Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T08:06:30Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-22T08:07:30Z', did: note, text: 'PriorityIcon (signal bars; urgent=alert/destructive). Row shows priority glyph (leading col) + label chips. Detail Properties: priority Select, labels editor (chips+add), parent Select -> useUpdateTask. Hierarchy: parent breadcrumb + Sub-tasks section (children + n/m progress + Add sub-task -> CreateTaskDialog with parent preset). Create dialog gained priority + labels. Verified on board + PROJ-017 detail (sub-task PROJ-022).'}
  - {who: 'agent:claude', at: '2026-06-22T08:07:30Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T08:07:33Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T08:07:33Z', did: transitioned to done}
assignee: agent:claude
priority: high
labels: [frontend, ui]
---
PriorityIcon + label chips in TaskRow/TaskDetail/CommandPalette. Edit from detail Properties (priority Select, labels combobox, parent picker) + CreateTaskDialog via useUpdateTask. Hierarchy: parent breadcrumb + Sub-tasks section (children + progress + add). Optional group-by-epic on board.