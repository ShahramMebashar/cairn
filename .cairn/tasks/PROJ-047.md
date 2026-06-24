---
id: PROJ-047
title: Remove session usage/token tracking entirely
status: done
priority: medium
labels: [backend, web, cleanup, sessions]
checks:
  - desc: go tests pass
    cmd: go test ./...
    cwd: .
    result: pass
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: no usage symbol remains in code/web/live-docs (the lone removeKey purge line is allowed)
    cmd: '! grep -rni usage internal web/src docs/agent-sessions.md docs/mcp-tools.md --include=*.go --include=*.ts --include=*.tsx --include=*.md | grep -v node_modules | grep -vi ''design.*token\|semantic token'' | grep -v ''removeKey(m, "usage")'''
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T17:35:05Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T17:35:12Z', did: began session ses_e8efcf1da3611407da465f22}
  - {who: 'agent:claude', at: '2026-06-24T17:42:52Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T17:43:05Z', did: note, text: 'Removed usage/token tracking end-to-end. Go: deleted session.Usage type + Merge, Session.Usage, Live.Usage, and the usage param of session.Finish; dropped usageIn/toSession (mcp/tools.go), Heartbeat/FinishSessionInput.Usage (mcp/sessions.go), usageReq/sessionUsage + HTTP fields (server/sessions.go); removed now-dead int64Node (store/store.go). store/session.go Replace now removeKey(m, "usage") to actively purge the deprecated key from any session file it rewrites. Web: dropped the Usage type + usage fields (api.ts) and all token rendering + the Wrench import (SessionTimeline.tsx). Docs: agent-sessions.md, mcp-tools.md, internal/repo/workflow.md (the init template) and .cairn/WORKFLOW.md; left the dated design specs under docs/superpowers/specs as historical record. Tests updated in session/store/mcp. Verified: go test ./... + pnpm build green; gofmt + vet clean; the no-usage grep check passes (only the intentional removeKey purge line remains, explicitly allowed). Residue note: the 3 .cairn/sessions/*.yaml still carry an ignored `usage:` key written by the pre-change binary; the new code purges it on any rewrite, but finished session files won''t be rewritten, so after rebuild+reconnect I can strip them manually if you want them gone.'}
  - {who: 'agent:claude', at: '2026-06-24T17:43:15Z', did: finished session ses_e8efcf1da3611407da465f22, text: 'Removed session usage/token tracking entirely, per decision that an unfed field is worse than no field. Deleted the Usage type, Session/Live usage fields, the Finish usage param, MCP usageIn + Heartbeat/Finish usage args, HTTP usageReq, and the web Usage type + token rendering; store now purges the deprecated `usage:` key on any session rewrite. Docs (agent-sessions, mcp-tools, the init workflow template + .cairn/WORKFLOW.md) updated; dated design specs left as historical record. go test ./... + pnpm build green, gofmt/vet clean, and a grep check enforces no usage symbol remains (bar the one intentional purge line). Review focus: confirm dropping the field from the session schema is acceptable (it was optional/omitempty, so no breakage) and whether the 3 historical .cairn/sessions/*.yaml should have their now-ignored usage keys stripped. Heads-up: this session''s cairn binary still predates PROJ-045/046/047 — rebuild + reconnect for any of these to be live (and to stop the binary re-writing usage onto session files).'}
  - {who: 'agent:claude', at: '2026-06-24T17:43:19Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T17:43:47Z', did: ran checks}
assignee: agent:claude
active_attempt: att_e8efcf1da3611407da465f22
---
Usage was never populated (no client reports it; cairn can't measure it) and rendered as a misleading `0`. Remove it end-to-end rather than ship a permanently-empty field.

## Scope
- `internal/session/doc.go` — drop `Usage` type + `Merge`, `Session.Usage`, `Live.Usage`, and the `usage` param of `Finish`.
- `internal/store/session.go` — remove `setUsage`; in `Replace` actively `removeKey(m, "usage")` so the deprecated key is purged from any rewritten session file.
- `internal/mcp/sessions.go`, `internal/mcp/tools.go` — drop usage from Heartbeat/Finish inputs + schemas; `usageIn`/`toSession`.
- `internal/server/sessions.go` — drop `usageReq`/`sessionUsage` and the HTTP usage fields.
- `web/src/lib/api.ts`, `web/src/components/SessionTimeline.tsx` — drop the `Usage` type and all token rendering.
- Docs: `docs/agent-sessions.md`, `docs/mcp-tools.md`, `internal/repo/workflow.md` (template) + `.cairn/WORKFLOW.md` — drop "cumulative usage" language. Leave dated design specs as historical record.
- Tests in `internal/session`, `internal/store`, `internal/mcp` that reference `Usage`.

## Acceptance
- No `Usage`/`usage` symbol left in engine code, web, or live docs (historical specs excepted).
- go test + web build green.