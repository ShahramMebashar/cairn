---
id: PROJ-008
title: Manual-check attestation verb (attest)
status: done
checks:
  - desc: attest sets a manual check to pass and refuses cmd checks
    cmd: go test ./internal/mcp ./internal/server
    cwd: .
    result: pass
  - desc: gofmt + vet + full test suite
    cmd: make check
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:08:43Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:08:54Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:08:59Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:11:59Z', did: note, text: 'Done. mcp.Service.Attest(id, index, pass) refuses cmd checks (ErrNotManual) and out-of-range indices, sets result via Doc.SetCheckResult, stamps provenance "attested check N pass/fail". Registered MCP tool `attest` (pass *bool, omit=pass). HTTP POST /api/tasks/{id}/attest {index, pass?}; ErrNotManual→422. Tests: TestAttestManualCheck, TestAttestUnblocksClose (mcp), TestAttestEndpointUnblocksClose (server). NOTE: the cairn MCP server connected to this session is the pre-change binary, so mcp__cairn__attest isn''t live here yet — rebuild + reconnect cairn to attest PROJ-006/007''s manual checks (or use the web Attest button from PROJ-007).'}
  - {who: 'agent:claude', at: '2026-06-21T20:12:07Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:12:14Z', did: transitioned to done}
assignee: agent:claude
---
Close the spec-vs-impl gap: SPEC says manual checks' result is "set by attestation," but no verb sets it, so any task with a manual check can never close. Discovered while dogfooding PROJ-006/007.

- `internal/mcp/service.go`: `Attest(id string, index int, pass bool) (*store.Doc, error)` — validate index, refuse non-manual checks (those with a Cmd are engine-set), set result pass/fail via Doc.SetCheckResult, append provenance ("attested check N pass/fail"), save.
- `internal/mcp/tools.go`: register `attest` MCP tool (id, index, pass default true).
- `internal/server`: `POST /api/tasks/{id}/attest` { index, pass } → updated task DTO.
- Pure gate logic untouched; reuses the existing checks gate.
- Web attest UI is folded into PROJ-007's check panel, not here.

Unblocks PROJ-006 and PROJ-007 reaching done.