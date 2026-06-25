---
id: PROJ-2anjy1mqaq
title: Package cairn as a cross-platform Tauri desktop app
status: in_review
priority: high
labels: [desktop, tauri, build]
provenance:
  - {who: 'agent:claude', at: '2026-06-25T06:35:45Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T06:35:53Z', did: began session ses_c9e4783b5cb331eed0069d3e}
  - {id: n_b6gc700k, who: 'agent:claude', at: '2026-06-25T06:37:36Z', did: note, text: 'Explored cairn + dbugd. Confirmed APIs: SDK has mcp.NewStreamableHTTPHandler(getServer func(*http.Request) *Server, opts) in go-sdk@v1.6.1; mcp.NewServer/NewServiceWithClient exist; server.Handler() builds the mux + resolveRoot validates ?path. Building Go core first on branch feat/desktop-tauri.'}
  - {id: n_nm0xe8e0, who: 'agent:claude', at: '2026-06-25T06:40:03Z', did: note, text: 'Go core done + verified: web/embed.go (//go:embed all:dist); server serves SPA at /, GET /healthz returns ok, /mcp Streamable-HTTP with per-conn repo+actor (MCP initialize returns serverInfo cairn + tools). cairn web prints CAIRN_WEB_URL on stdout, port falls back 7777->7778 when busy, --parent-watch exits on stdin EOF. go vet + go test ./... green.'}
  - {who: 'agent:claude', at: '2026-06-25T06:53:25Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-25T06:53:36Z', did: ran checks}
  - {id: n_2drh8qxk, who: 'agent:claude', at: '2026-06-25T06:53:57Z', did: note, text: 'Found + fixed a sharp edge: vite empties dist/ on build, deleting the embed placeholder. Added scripts/keep-dist.mjs run at the end of `pnpm build` so web/dist/.gitkeep always survives (go:embed stays green on fresh clone/CI). Verified packaged Cairn.app: launches, spawns sidecar, /healthz ok, embedded UI served, MCP initialize works, sidecar dies on quit. Local DMG step fails only due to headless Finder/AppleScript (CI macos-14 has a GUI session). CI lint step dropped — repo has pre-existing eslint debt unrelated to this work.'}
  - {who: 'agent:claude', at: '2026-06-25T06:54:13Z', did: finished session ses_c9e4783b5cb331eed0069d3e, text: "Packaged cairn as a cross-platform Tauri desktop app (branch feat/desktop-tauri, not yet committed).\n\nGo core: web/embed.go embeds the built UI; internal/server now serves the SPA at /, GET /healthz, and MCP over Streamable HTTP at /mcp?repo=&actor= (per-connection repo+actor, 400 on missing/invalid). cmd/cairn web binds 127.0.0.1:7777 with next-free-port fallback, prints CAIRN_WEB_URL on stdout, and --parent-watch shuts down on stdin EOF.\n\nDesktop: web/src-tauri/ (tauri.conf.json, Cargo.toml, build.rs, src/main.rs, src/lib.rs, capabilities, icons from app-icon.svg). lib.rs spawns the cairn sidecar, reads the URL from stdout, polls /healthz, navigates the webview, and kills the child on quit. scripts/build-sidecar.mjs builds the Go binary as the triple-named sidecar; scripts/set-desktop-version.mjs stamps tag versions; scripts/keep-dist.mjs preserves the embed placeholder across vite builds.\n\nFrontend: native folder picker in OpenProject (web/src/lib/tauri.ts), and a \"Connect an agent\" dialog (HTTP + stdio snippets) wired into AppSidebar.\n\nBuild/CI: Makefile desktop/desktop-dev/sidecar targets; .github/workflows/ci.yml (go test/vet, frontend build, no-bundle desktop check) and release.yml (mac/linux/windows installers on v* tags, unsigned MVP).\n\nVerified: go vet + go test green; pnpm build green; full pnpm tauri:build produced Cairn.app + sidecar; launched the packaged app — it spawns the server, serves the UI, answers MCP initialize, and the sidecar dies on quit. Open gates: not committed yet; signing deferred; local DMG packaging needs a GUI runner (works on CI macos-14); pre-existing repo eslint debt left out of CI."}
  - {who: 'agent:claude', at: '2026-06-25T07:12:51Z', did: began session ses_07d73d8770a7816997bca35a}
  - {who: 'agent:claude', at: '2026-06-25T07:17:01Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T08:33:13Z', did: ran checks}
assignee: agent:claude
active_attempt: att_07d73d8770a7816997bca35a
checks:
  - desc: Go vet + tests pass (server SPA/healthz/mcp endpoints)
    cmd: go vet ./... && go test ./...
    result: pass
  - desc: Frontend typechecks + builds (UI + new components)
    cmd: cd web && pnpm build
    result: pass
  - desc: Desktop bundle builds installers on CI (mac DMG needs a GUI runner; verified Cairn.app + sidecar + launch locally)
    type: manual
    result: pending
---
Wrap cairn as a light, native desktop app (macOS/Linux/Windows) using Tauri, mirroring the dbugd recipe.

Plan: /Users/shaho/.claude/plans/inherited-nibbling-blossom.md

Scope:
- Embed web/dist into the Go binary (`web/embed.go`, `//go:embed all:dist`).
- Server: serve SPA at `/`, add `GET /healthz`, add `/mcp` Streamable-HTTP endpoint with per-connection repo+actor binding.
- `cairn web`: fixed port 7777 + fallback, print `CAIRN_WEB_URL=`, `--parent-watch` stdin-EOF shutdown.
- Tauri scaffold under `web/src-tauri/` (sidecar = cairn binary), native folder dialog, "Connect an agent" panel.
- GitHub Actions: release on `v*` tag (mac/linux/windows installers), build-check on PRs. Unsigned MVP.

Decisions: native folder picker; unsigned builds; tag+PR CI; MCP over HTTP+stdio both; fixed port 7777 + fallback.