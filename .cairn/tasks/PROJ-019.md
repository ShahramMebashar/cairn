---
id: PROJ-019
title: Dependency graph view (React Flow + dagre)
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: graph reads clearly (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:32:54Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T07:50:21Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T07:50:21Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T07:57:30Z', did: note, text: 'Graph page via #/<slug>/graph (Cmd-K + sidebar). @xyflow/react nodes from tasks, edges from deps (dep->dependent), @dagrejs/dagre LR auto-layout. Custom node: StatusIcon+id+title, ready highlighted (brand ring), closed dimmed, animated edges into ready tasks. Click node -> open task; pan/zoom/fit/controls. Verified 20 nodes laid out.'}
  - {who: 'agent:claude', at: '2026-06-22T07:57:30Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T07:57:33Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T07:57:33Z', did: transitioned to done}
assignee: agent:claude
---
#/<slug>/graph route + sidebar nav. @xyflow/react nodes from tasks, edges from deps, dagre LR layout. Status colors, ready highlighted, click->open, hover highlights ancestors/descendants, pan/zoom/fit.