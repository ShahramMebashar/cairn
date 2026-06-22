---
id: PROJ-015
title: 'init: write generic WORKFLOW.md and wire AGENTS.md/CLAUDE.md'
status: done
checks:
  - desc: go check passes
    cmd: make check
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T21:02:42Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T21:02:42Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T21:02:42Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T21:07:28Z', did: note, text: 'repo.Init now embeds a generic WORKFLOW.md (workflow.md) and writes it on init (never overwriting an existing one), then ensures AGENTS.md/CLAUDE.md reference it (append section if present, create stub if absent; idempotent). Made this repo''s own WORKFLOW.md generic and moved the cairn-specific friction log to .cairn/FRICTION.md. Tests: TestInitWritesWorkflowAndAgentDocs, ...DoesNotOverwriteWorkflow, ...AppendsRefToExistingAgentDocOnce.'}
  - {who: 'agent:claude', at: '2026-06-21T21:07:29Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T21:07:29Z', did: transitioned to done}
assignee: agent:claude
---
Make the cairn workflow a generic template shipped by init.

- `repo.Init` writes `.cairn/WORKFLOW.md` from an embedded generic template (idempotent: never overwrites an existing one)
- Ensure root `AGENTS.md` and `CLAUDE.md` reference the workflow: append a section if the file exists (and lacks the ref), create a stub if it does not
- Generic content only — no cairn-repo-specific friction log
- Tests in internal/repo