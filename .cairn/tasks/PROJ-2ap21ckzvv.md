---
id: PROJ-2ap21ckzvv
title: Add Connect page with one-click agent auto-config
status: in_review
priority: high
checks:
  - desc: Go build, vet, and tests pass
    cmd: make check
    timeout: 300
    result: pass
  - desc: Web typecheck/build passes
    cmd: pnpm --dir web build
    timeout: 300
    result: pass
  - desc: 'End-to-end: Connect page writes a valid agent config and flips to Connected'
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-25T10:53:32Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T10:53:37Z', did: began session ses_fde7b1192ee82f7f2d8feb77}
  - {id: n_wek7gpmm, who: 'agent:claude', at: '2026-06-25T10:59:06Z', did: note, text: 'Backend done. New `internal/connect` package: registry of agents + format adapters (mcpServersJSON for Claude/Cursor/Windsurf, codexTOML for Codex, openCodeJSON for OpenCode). Writes are safe-merge (preserve siblings) + atomic + `.bak` + write-and-verify. Binary path comes from os.Executable()+EvalSymlinks so the desktop sidecar path is embedded (never bare `cairn`). Added go-toml/v2 dep. Endpoints in internal/server/connect.go: GET /api/connect, POST /api/connect/{agent}, GET /api/connect/{agent}/manual. Codex/Kilo paths are project-local (.codex/config.toml, .kilocode/mcp.json) matching repo evidence; Antigravity/Pi/Kilo are manual-only until paths verified. Tests green.'}
  - {who: 'agent:claude', at: '2026-06-25T11:04:05Z', did: ran checks}
  - {id: n_8e40thca, who: 'agent:claude', at: '2026-06-25T11:04:15Z', did: note, text: 'Frontend done + verified end-to-end. New `web/src/pages/Connect.tsx` (detected agents first, full catalog with one-click Connect + expandable manual guide). Routing wired in App.tsx (`#/<slug>/connect`, View union, render branch); sidebar gained a Connect nav entry + the dropdown "Connect an agent…" now navigates to the page; ConnectAgentDialog deleted (superseded), its CodeBlock extracted to `web/src/components/CodeBlock.tsx`. api.ts/queries.ts: listIntegrations/connectAgent/agentManual + useIntegrations/useConnectAgent/useAgentManual (invalidates integrations on connect so the card flips to Connected). Live test via `cairn web`: GET /api/connect detected all 5 auto agents (Windsurf→global path, others project-local), POST /api/connect/cursor wrote .cursor/mcp.json with the absolute binary path + --repo and verified connected:true, codex manual returned valid TOML. make check + pnpm web build both pass.'}
  - {who: 'agent:claude', at: '2026-06-25T11:04:29Z', did: finished session ses_fde7b1192ee82f7f2d8feb77, text: "Adds frictionless agent integrations: a new Connect page + one-click MCP auto-config written by the cairn process itself.\n\nBackend: new `internal/connect` package — agent registry + per-format adapters (mcpServersJSON: Claude/Cursor/Windsurf; codexTOML: Codex; openCodeJSON: OpenCode). Writes are safe-merge (preserve siblings), atomic (temp+rename), back up to `.bak`, and write-and-verify. Binary path resolved via os.Executable()+EvalSymlinks so the desktop sidecar abs path is embedded (never bare `cairn`). v1 auto: Claude, Cursor, Codex, Windsurf, OpenCode (project-local paths; Windsurf global). Manual-only catalog: Kilo, Antigravity, Pi. HTTP in internal/server/connect.go: GET /api/connect, POST /api/connect/{agent}, GET /api/connect/{agent}/manual. Added go-toml/v2.\n\nFrontend: web/src/pages/Connect.tsx (installed-first sections + catalog, one-click Connect, expandable manual guide, agent-identity field). Routing in App.tsx, sidebar Connect entry, dropdown repointed to the page. Shared CodeBlock extracted; ConnectAgentDialog removed. api/queries hooks added; connect invalidates detection so cards flip to Connected live.\n\nVerification: make check (vet+tests, incl. new connect tests for merge idempotency/.bak/detection) green; pnpm web build green. Live `cairn web` run: detection found all installed agents with correct scoped paths; POST cursor wrote a valid .cursor/mcp.json (abs bin path + --repo, verified); codex manual returned valid TOML.\n\nReview notes: (1) manual-only agent config paths (Kilo/Antigravity/Pi) not yet verified against their docs — intentionally degraded to guide-only. (2) Codex TOML emits an empty `[mcp_servers]` parent header before `[mcp_servers.cairn]` — valid and parses, minor cosmetic. (3) Desktop sidecar path embedding is verified only via os.Executable() logic + dev-binary run, not yet exercised inside the packaged Tauri app."}
  - {who: 'agent:claude', at: '2026-06-25T11:47:10Z', did: began session ses_d33d96940de92297fe9cb0f4}
  - {id: n_5khmjant, who: 'agent:claude', at: '2026-06-25T11:48:13Z', did: note, text: 'Promoted Kilo + Pi to auto after doc verification (user asked to confirm before guessing). Kilo: project `.kilocode/mcp.json`, standard mcpServers (created-if-missing, precedence over global) — kilo.ai docs. Pi: project `.mcp.json`, mcpServers shape (shared with Claude''s file — idempotent) — pi.dev/pi-mcp-adapter. Antigravity stays manual: mcpServers shape confirmed but exact path (`~/.gemini/config/mcp_config.json`, single source) and stdio command/args not doc-confirmed, so guide-only with an accurate path hint (target set for guide display; Connect refuses). Added detection for all three. New test TestConnectAutoAgentsWriteExpectedPaths. Live-verified via cairn web: kilo/pi write valid configs + connected:true; antigravity refuses auto and returns a correct mcpServers JSON guide. No frontend change — page derives Connect vs Manual from mode.'}
  - {who: 'agent:claude', at: '2026-06-25T11:48:23Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T11:55:07Z', did: began session ses_7929787f40209bbf5040600a}
  - {id: n_g0c3fr5c, who: 'agent:claude', at: '2026-06-25T11:57:28Z', did: note, text: 'Fixed agent identity: each agent now connects as its own actor (`agent:<id>`, e.g. agent:cursor) by default instead of the human operator''s name. Backend: new resolveActor(agentID, actor) in connect.go defaults empty→`agent:<id>`; server handlers no longer fall back to s.actorFor(r)/X-Cairn-Actor (that was leaking the human identity into agent configs and provenance). Frontend Connect.tsx: removed the shared human-seeded "Agent identity" field; each AgentCard owns an editable identity defaulting to `agent:<id>` (inline muted input, supports overrides like agent:cursor-2). New test TestConnectDefaultsToPerAgentIdentity. Live-verified: POST connect with no actor + a human header writes agent:cursor (no leak); override→agent:codex-2; manual guide defaults to agent:windsurf. make check + web build green.'}
  - {who: 'agent:claude', at: '2026-06-25T11:57:40Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T12:06:05Z', did: began session ses_ba386bc76686674c3b2b6bc0}
  - {id: n_ghtc2per, who: 'agent:claude', at: '2026-06-25T12:08:43Z', did: note, text: 'Added Disconnect (remove cairn from an agent''s config). Backend: format.remove(existing) on each adapter drops only the `cairn` key (jsonNestedDelete for JSON shapes — also prunes an emptied mcpServers/mcp map; codexTOML prunes an emptied mcp_servers table). connect.Disconnect(agentID, repo): no-op if file absent/not-connected, else atomic write + .bak + verify-gone. Endpoint DELETE /api/connect/{agent}. Frontend: useDisconnectAgent hook; connected cards show an Unplug ghost icon-button next to Reconnect (destructive hover) that disconnects and invalidates detection so the card flips back. Tests TestDisconnectRemovesOnlyCairnAndKeepsSiblings + TestDisconnectMissingFileIsNoop. Live-verified: connect added cairn beside a sibling `other` server, disconnect removed only cairn (kept `other`), wrote .bak, detection→connected:false. make check + web build green.'}
  - {who: 'agent:claude', at: '2026-06-25T12:08:57Z', did: ran checks}
assignee: agent:claude
active_attempt: att_ba386bc76686674c3b2b6bc0
---
Frictionless agent integrations. The cairn Go process detects installed AI agents and writes their MCP config files directly (project-local, stdio, sidecar-safe abs binary path). New sidebar **Connect** page: detect installed agents + full catalog with manual guides.

Plan: `~/.claude/plans/i-want-to-make-smooth-sutton.md`.

## Scope
- Backend `internal/connect`: agent registry + format adapters (`mcpServersJSON`, `codexTOML`, `openCodeJSON`), safe-merge writes with `.bak`, `Detect` + `Connect` (write-and-verify).
- HTTP: `GET /api/connect`, `POST /api/connect/{agent}`, `GET /api/connect/{agent}/manual` in `internal/server`.
- Frontend: `web/src/pages/Connect.tsx`, route wiring in `App.tsx`, sidebar entry, `api.ts`/`queries.ts` hooks; extract shared `CodeBlock`.

## v1 auto
Claude, Cursor, Codex, Windsurf, OpenCode. Manual-only: Antigravity, Pi, Kilo.

## Acceptance
- `make check` green; `go test ./internal/connect/...` covers merge idempotency + `.bak` + detection.
- Connect page lists agents, one-click writes a valid config with abs binary path + `--repo`, card flips to Connected.