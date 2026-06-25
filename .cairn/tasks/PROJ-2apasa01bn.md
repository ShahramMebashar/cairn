---
id: PROJ-2apasa01bn
title: Add README screenshot, SECURITY.md, and CHANGELOG (versioning)
status: in_review
priority: medium
checks:
  - desc: Meta files + screenshot asset exist
    cmd: test -f README.md && test -f SECURITY.md && test -f CHANGELOG.md && test -f docs/public/app-screenshot.png
    timeout: 30
    result: pass
  - desc: Docs build clean (screenshot on home)
    cmd: npm --prefix docs run build
    timeout: 300
    result: pass
  - desc: 'Manual: README + SECURITY + CHANGELOG reviewed'
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-25T13:22:50Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T13:23:24Z', did: began session ses_04418ccf7e0048fa10f6f709}
  - {who: 'agent:claude', at: '2026-06-25T13:27:56Z', did: ran checks}
  - {id: n_a80k0ehf, who: 'agent:claude', at: '2026-06-25T13:28:08Z', did: note, text: 'Done. Screenshot: copied app-screenshot.png → docs/public/, embedded on docs home (index.md body) and the new root README. Root README written (was empty): logo + screenshot + tagline + quickstart + run-it-your-way table + agents + dev + security + MIT, linking the docs site. SECURITY.md: cairn-specific — local no-auth trust model, the arbitrary-shell-via-checks warning (GitHub [!WARNING] alert, not a VitePress container since it renders on GitHub), identity/safe-merge/lock guarantees, report via GitHub Security Advisories. CHANGELOG.md: Keep a Changelog + SemVer, [Unreleased] (Connect integrations, Disconnect, docs site, SECURITY) + [0.1.0] initial; notes version lives in ServiceVersion + tauri.conf.json. Brand reconciled to #4c5edf indigo per user: switched docs logo from the purple favicon to the indigo cairn stones mark (light + dark variants for nav/hero), and set the VitePress theme accent to #4c5edf via .vitepress/theme/custom.css. Nav v0.x dropdown now links Changelog + Security policy. Build clean; logos + screenshot + indigo CSS all in dist.'}
  - {who: 'agent:claude', at: '2026-06-25T13:28:24Z', did: finished session ses_04418ccf7e0048fa10f6f709, text: "Repo-meta polish + indigo brand reconciliation.\n\nScreenshot: copied app-screenshot.png → docs/public/; embedded on the docs home (index.md body, below the hero) and in the new root README.\n\nREADME.md (was empty): centered indigo logo, full-width screenshot, tagline, quickstart, a \"run it your way\" table (desktop/web/MCP), supported agents, development commands, a security pointer, and MIT — all linking the docs site.\n\nSECURITY.md: cairn-specific. Local single-user / no-auth-by-design trust model; a prominent GitHub [!WARNING] alert that checks run arbitrary shell (closing a task auto-runs them) so only operate on trusted repos; guarantees (actor-stamping + begin's expected_actor refusal, safe-merge config writes with .bak + verify, advisory write.lock, no human-actor fallback for agents); report via GitHub Security Advisories. Used a GitHub alert, not a VitePress container, since it renders on GitHub.\n\nCHANGELOG.md: Keep a Changelog + SemVer; [Unreleased] (Connect integrations, Disconnect, docs site, SECURITY) and [0.1.0] initial; notes the version lives in internal/mcp/sessions.go ServiceVersion + desktop/src-tauri/tauri.conf.json and that pre-1.0 minors may break.\n\nBrand → #4c5edf indigo (per user): replaced the docs logo (was the purple #863bff/#7e14ff favicon mark) with the actual cairn stacked-stones mark, which uses #4c5edf — added light (docs/public/logo.svg) and dark (logo-dark.svg, dark stones lightened) variants wired to nav + hero. Set the VitePress theme accent to indigo via docs/.vitepress/theme/{index.ts,custom.css} (brand-1/2/3 + hero name gradient, with a .dark override). Nav v0.x dropdown now links Changelog + Security policy + SPEC.\n\nVerification: `npm --prefix docs run build` clean; logo.svg, logo-dark.svg, app-screenshot.png published to dist; indigo #4c5edf present 5× in the built CSS. Both command checks pass.\n\nOpen: manual review of README/SECURITY/CHANGELOG. Note the web app favicon (web/public/favicon.svg) is still the purple mark — not touched here; say the word to align it (and the Tauri icon) to #4c5edf for full brand consistency."}
assignee: agent:claude
active_attempt: att_04418ccf7e0048fa10f6f709
---
Repo-meta polish alongside the docs site.

## Scope
- Add `app-screenshot.png` to the root `README.md` (currently empty) and the docs home (`docs/index.md`); copy it into `docs/public/`.
- Write a proper root `README.md`: logo, screenshot, tagline, quickstart, links to the docs site + agents, license.
- `SECURITY.md` — cairn-specific (local single-user tool, no-auth trust model, actor stamping, safe-merge config writes, checks run arbitrary shell → trust your repos; report via GitHub Security Advisories).
- `CHANGELOG.md` — Keep a Changelog + SemVer; seed `[Unreleased]` + `[0.1.0]` (version per `internal/mcp/sessions.go` ServiceVersion + `desktop/src-tauri/tauri.conf.json`).
- Link CHANGELOG + SECURITY from the docs nav (`docs/.vitepress/config.mts` v0.x dropdown).

## Acceptance
- README/SECURITY/CHANGELOG exist; screenshot renders on docs home; `npm --prefix docs run build` clean.