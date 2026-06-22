---
id: PROJ-036
title: 'Authoring convention: structured task bodies'
status: in_progress
labels: [docs, workflow]
checks:
  - desc: go check passes
    cmd: make check
    result: pending
  - desc: convention reads well (manual)
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-22T16:39:17Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T16:39:25Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T16:39:30Z', did: transitioned to in_progress}
assignee: agent:claude
---
Guide LLM agents to write task bodies that are concise yet structured — good for humans and agents.

## Scope
- `internal/repo/workflow.md`: add a short **Body style** subsection under *Authoring tasks* — lead with a one-line summary; use `##` sections only when the task has structure; inline code for identifiers/paths; tight bullets; token-efficient. Include a ~6-line example.
- `AGENTS.md`: one-line pointer to the convention.

## Notes
Ships to every workspace via `cairn init`. Lightweight rules, not a rigid template.