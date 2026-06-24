---
id: PROJ-01kvxmmqk4pjbepx
title: Embed cairn agent-loop block in AGENTS.md/CLAUDE.md on init
status: done
priority: medium
labels: [backend, init, dx]
checks:
  - desc: Go tests pass
    cmd: go test ./...
    timeout: 180
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T20:20:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T20:20:21Z', did: began session ses_45e4280ed463816c9a4d8582}
  - {who: 'agent:claude', at: '2026-06-24T20:21:33Z', did: note, text: 'Done. repo.go: replaced ensureWorkflowRef with ensureAgentLoop — embeds a marker-wrapped (<!-- cairn:agent-loop:start/end -->) concise "Agent loop — required" block (identity→find→begin→heartbeat→note→run_checks→finish→close + link to WORKFLOW.md) inline in AGENTS.md/CLAUDE.md. Re-init replaces only between markers; appends if no markers; creates with header if file absent — content outside markers untouched. Updated Init call site + package doc + repo_test idempotency check (now asserts agentLoopStart count==1 and the block heading). go test ./... green; verified rendered output via `cairn init` in a temp dir. Why inline not link: harnesses auto-load AGENTS.md/CLAUDE.md but don''t follow the WORKFLOW.md link, so the loop has to be in-file. Rebuilt bin/cairn (running server needs restart to apply on future inits).'}
  - {who: 'agent:claude', at: '2026-06-24T20:21:39Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T20:21:49Z', did: finished session ses_45e4280ed463816c9a4d8582, text: "repo.Init now embeds a cairn-managed agent-loop block inline in AGENTS.md/CLAUDE.md instead of just linking to WORKFLOW.md — fixing agents skipping the workflow (harnesses auto-load these docs but don't follow links).\n\n- internal/repo/repo.go: ensureWorkflowRef → ensureAgentLoop. Marker-wrapped (<!-- cairn:agent-loop:start/end -->) concise \"Agent loop — required\" block: the 8-step loop (identity → find → begin → heartbeat → note → run_checks → finish → close) + a link to .cairn/WORKFLOW.md for depth. Re-init replaces only the content between markers (in-place refresh, so the block can evolve with cairn); appends once if markers absent; creates the file with a header if missing. Content outside the markers is never touched.\n- internal/repo/repo_test.go: idempotency test now asserts exactly one marker pair after two inits and that the block is present.\n\nVerified: go test ./... green; rendered AGENTS.md via `cairn init` in a temp dir looks correct. Rebuilt bin/cairn — the running MCP/web server must restart for future inits to use it. Review focus: wording/length of the embedded block (it lands in every agent's auto-loaded context)."}
  - {who: 'human:shaho', at: '2026-06-24T20:21:57Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T20:21:57Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_45e4280ed463816c9a4d8582
rank: !!float 1
---
Agents skip the workflow because AGENTS.md/CLAUDE.md only *link* to `.cairn/WORKFLOW.md`, and linked files aren't auto-loaded into agent context. Make `repo.Init` embed a concise, self-contained "Agent loop — required" block inline instead of just a reference.

## Scope
- `internal/repo/repo.go` — replace `ensureWorkflowRef` with `ensureAgentLoop`: a marker-wrapped (`<!-- cairn:agent-loop:start/end -->`) block containing the required loop (identity → find ready → begin → heartbeat → note → run_checks → finish → close) + link to WORKFLOW.md. Re-init refreshes content between markers in place; file content outside markers untouched; append if no markers; create with header if file absent.
- `internal/repo/repo_test.go` — update the idempotency test (heading → marker count == 1).

## Acceptance
- Init on a fresh dir creates AGENTS.md + CLAUDE.md with the marked block.
- Init twice → exactly one marked block (in-place refresh).
- Existing file content preserved. go test green.