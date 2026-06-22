---
id: PROJ-036
title: 'Authoring convention: structured task bodies'
status: done
labels: [docs, workflow]
checks:
  - desc: go check passes
    cmd: make check
    result: pass
  - desc: convention reads well (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T16:39:17Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T16:39:25Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T16:39:30Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T16:41:04Z', did: note, text: 'Added a "Body style" subsection to internal/repo/workflow.md (the embedded template cairn init writes) under Authoring tasks: one-line intent, optional ## sections (Scope/Dependencies/Acceptance), inline code for identifiers/paths, tight bullets, token-efficient — with a ~6-line example. Mirrored it into this repo''s .cairn/WORKFLOW.md (init won''t overwrite an existing one) and added a one-line pointer in AGENTS.md. Lightweight rules, not a rigid template; ships to every workspace via init. make check green.'}
  - {who: 'agent:claude', at: '2026-06-22T16:41:10Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T16:41:17Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T16:41:17Z', did: transitioned to done}
assignee: agent:claude
---
Guide LLM agents to write task bodies that are concise yet structured — good for humans and agents.

## Scope
- `internal/repo/workflow.md`: add a short **Body style** subsection under *Authoring tasks* — lead with a one-line summary; use `##` sections only when the task has structure; inline code for identifiers/paths; tight bullets; token-efficient. Include a ~6-line example.
- `AGENTS.md`: one-line pointer to the convention.

## Notes
Ships to every workspace via `cairn init`. Lightweight rules, not a rigid template.