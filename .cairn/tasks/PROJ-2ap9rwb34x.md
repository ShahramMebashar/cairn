---
id: PROJ-2ap9rwb34x
title: Build VitePress documentation site (mirroring parakit)
status: in_review
priority: high
checks:
  - desc: VitePress build is clean (no dead internal links)
    cmd: npm --prefix docs install && npm --prefix docs run build
    timeout: 420
    result: pass
  - desc: No references to removed/renamed docs files remain
    cmd: '! rg -n ''docs/(getting-started|connect-claude|agent-sessions)\.md'' --glob ''!docs/**'' --glob ''!.claude/**'''
    timeout: 60
    result: pass
  - desc: 'Manual: agent pages reviewed for accuracy vs live /api/connect manual output'
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-25T13:05:32Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T13:05:39Z', did: began session ses_6de3a97b7e1be6c7dc1fa08a}
  - {who: 'agent:claude', at: '2026-06-25T13:13:18Z', did: ran checks}
  - {id: n_2y6mm2qp, who: 'agent:claude', at: '2026-06-25T13:13:34Z', did: note, text: 'VitePress docs site complete under docs/. Scaffold: .vitepress/config.mts (base /cairn/, local search, editLink, Guide/Agents/Reference nav+sidebar), package.json (vitepress ^1.6.4), .gitignore. Pages: index.md (hero + supported-agents table), introduction, installation, quickstart; guides/ (task-files, agent-loop, sessions, checks-and-gates); agents/ overview + 8 per-agent pages (claude/cursor/codex/windsurf/opencode/kilo/pi auto, antigravity manual) using exact snippets captured live from GET /api/connect/<agent>/manual; reference/ (mcp-tools, cli, http-api, events) sourced from cmd/cairn/main.go + internal/server. Migrated then removed the 6 old flat docs (getting-started, connect-claude, agent-sessions, mcp-tools, task-files, README); no external refs pointed at them. Deploy: .github/workflows/docs.yml → GitHub Pages on push to main. `npm run build` clean (VitePress dead-link check passes). Bulk pages written by 3 parallel subagents from exact data; spot-checked codex + http-api for accuracy. One-time user action: enable Pages with Source=GitHub Actions.'}
  - {who: 'agent:claude', at: '2026-06-25T13:13:50Z', did: finished session ses_6de3a97b7e1be6c7dc1fa08a, text: "Built a full VitePress documentation site under docs/, modeled on the parakit project.\n\nStructure (docs/):\n- .vitepress/config.mts — title/description, base '/cairn/', cleanUrls, lastUpdated, local search, GitHub editLink + social, Guide/Agents/Reference nav + grouped sidebar.\n- package.json (vitepress ^1.6.4, dev/build/preview), .gitignore for dist/cache/node_modules.\n- Home (index.md, layout: home) — hero, feature cards, supported-agents table.\n- Getting started: introduction, installation, quickstart.\n- Core concepts (guides/): task-files, agent-loop, sessions, checks-and-gates.\n- Agents: overview (agents/index.md) + one page each for claude, cursor, codex, windsurf, opencode, kilo, pi, antigravity. Per-agent config snippets are the EXACT output captured live from GET /api/connect/<agent>/manual, so they match the registry (internal/connect/registry.go).\n- Reference: cli, mcp-tools, http-api, events — sourced from cmd/cairn/main.go and internal/server/{server,connect,events}.go.\n\nMigrated the six old flat docs into the new tree and removed them (getting-started, connect-claude, agent-sessions, mcp-tools, task-files, README) — confirmed nothing outside docs/ referenced them.\n\nDeploy: .github/workflows/docs.yml builds VitePress and deploys to GitHub Pages on push to main (Node 22, fetch-depth 0 for lastUpdated).\n\nVerification: `npm --prefix docs run build` is clean — VitePress's dead-internal-link check passes (would fail the build otherwise). The no-stale-reference rg check passes. Bulk content authored by 3 parallel subagents from exact data; spot-checked agents/codex.md and reference/http-api.md for accuracy.\n\nReview notes / follow-ups: (1) Manual check still open — a human should skim the 8 agent pages once. (2) One-time repo setting required before the site goes live: Settings → Pages → Source = \"GitHub Actions\". (3) Site will publish at https://shahrammebashar.github.io/cairn/. (4) Optional: add a docs link to the root README and run `npm --prefix docs run dev` for a local visual pass."}
  - {who: 'agent:claude', at: '2026-06-25T13:18:31Z', did: began session ses_42980a7c02f8551da745827c}
  - {who: 'agent:claude', at: '2026-06-25T13:20:01Z', did: ran checks}
  - {id: n_g0qge0t2, who: 'agent:claude', at: '2026-06-25T13:20:13Z', did: note, text: 'Added branding + desktop-app docs. Logo: copied web/public/favicon.svg → docs/public/logo.svg (vivid #863bff mark, theme-safe on light+dark; the desktop cairn-logo.svg uses dark fills that vanish on a dark navbar). Wired as themeConfig.logo (nav), hero.image (index.md), and the head favicon (/cairn/logo.svg with base). Desktop app: new "## Desktop app" section in installation.md framed as the easy path — `make desktop` (Tauri installer, embeds UI + cairn sidecar) and `make desktop-dev`, with a placeholder ::: warning that prebuilt signed installers aren''t published yet (build from source until the first tagged release lands on Releases + auto-updates). Renamed "## Build" → "## Build the binary"; added a 5th home feature card ("Run it however you like"); noted Node/Rust requirements. Build clean; logo.svg published to dist and referenced.'}
assignee: agent:claude
active_attempt: att_42980a7c02f8551da745827c
---
Stand up a VitePress docs site under `docs/`, modeled on the parakit project: home hero, Guide + Agents + Reference sections, one page per agent integration, local search, GitHub Pages deploy.

Plan: `~/.claude/plans/i-want-to-make-smooth-sutton.md`.

## Scope
- `docs/.vitepress/config.mts` + `docs/package.json` (vitepress) + `.gitignore` for dist/cache.
- Pages: `index.md` (hero + supported-agents table), introduction, installation; `guides/` (task-files, agent-loop, sessions, checks); `agents/` overview + 8 per-agent pages; `reference/` (mcp-tools, cli, http-api, events).
- Migrate existing `docs/*.md` content into the new structure; update cross-refs in AGENTS.md/README/CLAUDE.md/.cairn/WORKFLOW.md.
- `.github/workflows/docs.yml` Pages deploy (base `/cairn/`).

## Accuracy
Per-agent config paths/snippets must match `internal/connect/registry.go` and the live `GET /api/connect/<agent>/manual`.

## Acceptance
- `cd docs && npm install && npm run build` is clean (VitePress fails on dead internal links).
- 8 agent pages render with correct config snippets; no stale `docs/*.md` references remain.