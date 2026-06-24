---
id: PROJ-01kvxmmqk4pjbepx
title: Embed cairn agent-loop block in AGENTS.md/CLAUDE.md on init
status: in_progress
priority: medium
labels: [backend, init, dx]
checks:
  - desc: Go tests pass
    cmd: go test ./...
    timeout: 180
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-24T20:20:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T20:20:21Z', did: began session ses_45e4280ed463816c9a4d8582}
assignee: agent:claude
active_attempt: att_45e4280ed463816c9a4d8582
---
Agents skip the workflow because AGENTS.md/CLAUDE.md only *link* to `.cairn/WORKFLOW.md`, and linked files aren't auto-loaded into agent context. Make `repo.Init` embed a concise, self-contained "Agent loop — required" block inline instead of just a reference.

## Scope
- `internal/repo/repo.go` — replace `ensureWorkflowRef` with `ensureAgentLoop`: a marker-wrapped (`<!-- cairn:agent-loop:start/end -->`) block containing the required loop (identity → find ready → begin → heartbeat → note → run_checks → finish → close) + link to WORKFLOW.md. Re-init refreshes content between markers in place; file content outside markers untouched; append if no markers; create with header if file absent.
- `internal/repo/repo_test.go` — update the idempotency test (heading → marker count == 1).

## Acceptance
- Init on a fresh dir creates AGENTS.md + CLAUDE.md with the marked block.
- Init twice → exactly one marked block (in-place refresh).
- Existing file content preserved. go test green.