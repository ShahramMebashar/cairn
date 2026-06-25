---
id: PROJ-2apf7aebwq
title: Scope release to desktop-only (drop GoReleaser/install.sh) + fix pnpm setup
status: done
priority: high
checks:
  - desc: release.yml valid, no goreleaser/install.sh refs; goreleaser+install.sh removed
    cmd: ruby -ryaml -e "YAML.load_file('.github/workflows/release.yml')" && ! test -f .goreleaser.yaml && ! test -f install.sh && ! grep -rqi 'goreleaser\|install.sh' .github/workflows/release.yml && echo OK
    timeout: 60
    result: pass
  - desc: Docs build clean
    cmd: npm --prefix docs run build
    timeout: 300
    result: pass
  - desc: 'Manual: tagged release builds desktop installers + latest.json (no binary archives)'
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-25T14:38:34Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T14:38:50Z', did: began session ses_b6ee9633861c73a33fc73c93}
  - {who: 'agent:claude', at: '2026-06-25T14:40:17Z', did: ran checks}
  - {id: n_a7t6gjv4, who: 'agent:claude', at: '2026-06-25T14:40:31Z', did: note, text: 'Release is now desktop-only. release.yml: removed the GoReleaser job entirely; the single Tauri desktop matrix (macOS arm64+x86_64, Linux, Windows) creates the release itself (no needs:), tauri-action with includeUpdaterJson:true → installers + signed latest.json; Apple signing + TAURI_SIGNING env unchanged. Fixed the failing run by pinning pnpm: added PNPM_VERSION=10 env and `with: version: ${{ env.PNPM_VERSION }}` on pnpm/action-setup (root has no packageManager field, lockfiles are v9.0, local pnpm 10.18.3). Deleted .goreleaser.yaml + install.sh. Docs: removed the "Install the binary" curl|sh section from installation.md (kept Desktop app + Build-from-source), rewrote the CHANGELOG entry to desktop-only (signed dmg / nsis / deb+appimage + latest.json, Go binary embedded as sidecar). README already clean. checks pass (release.yml valid, no goreleaser/install.sh refs, docs build clean). Note: the prior v0.1.0 tag run failed on the pnpm step before creating a release — user needs to commit these fixes and move/re-push the tag to retrigger. cairn `version` command kept (harmless; prints "dev" without GoReleaser injection).'}
  - {who: 'agent:claude', at: '2026-06-25T14:40:45Z', did: finished session ses_b6ee9633861c73a33fc73c93, text: "Scoped the release to desktop-only and fixed the pnpm failure.\n\n- release.yml: removed the GoReleaser job; the workflow is now a single Tauri desktop matrix (macOS arm64+x86_64, Linux, Windows) that creates the release itself and, via tauri-action includeUpdaterJson, publishes the installers plus a signed latest.json. Apple signing + TAURI_SIGNING env unchanged.\n- Fixed the failing run (\"No pnpm version is specified\"): added PNPM_VERSION=10 and `with: version: ${{ env.PNPM_VERSION }}` to pnpm/action-setup (the repo root has no packageManager field for the action to infer from).\n- Deleted .goreleaser.yaml and install.sh.\n- Docs: removed the curl|sh \"Install the binary\" section from installation.md (kept Desktop app + Build-from-source); rewrote the CHANGELOG entry to desktop-only. README was already clean.\n\nThe Go binary still exists and is used (cairn web/serve, and embedded as the desktop sidecar) — it's just no longer distributed standalone. Building from source (make build) is unchanged.\n\nVerified: release.yml parses and references no goreleaser/install.sh; docs build clean.\n\nTo ship: the earlier v0.1.0 run failed on the pnpm step before any release was created, so no assets exist. Commit these changes and move the tag to the new commit (git tag -f v0.1.0 && git push -f origin v0.1.0) — or use a fresh tag — to retrigger. With the TAURI_SIGNING_PRIVATE_KEY secret set, the run will produce signed/notarized macOS + Linux + Windows installers and latest.json. Windows code signing remains deferred."}
  - {who: 'agent:claude', at: '2026-06-25T14:56:28Z', did: began session ses_04c57c07e05a035288a69c12}
  - {id: n_n19bezj9, who: 'agent:claude', at: '2026-06-25T14:58:32Z', did: note, text: 'Docs now lead with download, not build-from-source (v0.1.0 desktop installers are live). index.md hero: added a primary "Download" button → releases/latest (Get started/Connect/GitHub demoted to alt); updated the "Run it however you like" feature card to mention downloading. installation.md: replaced the placeholder Desktop-app section with "## Desktop app (recommended)" — a Download link + per-OS table (macOS signed dmg by arch; Windows NSIS .exe with the SmartScreen note; Linux AppImage/.deb) + auto-update note; reframed the page intro so downloading needs no toolchain and renamed the build path "## Build from source" with its Go/Node/Rust requirements folded in (dropped the standalone Requirements section). README: added a bold Download link to the header row and changed the run-it-your-way table''s "installers coming to Releases" to a live Download link. Docs build clean.'}
  - {who: 'human:shaho', at: '2026-06-25T19:14:18Z', did: attested, text: check 2 pass}
  - {who: 'human:shaho', at: '2026-06-25T19:14:24Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:14:24Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_04c57c07e05a035288a69c12
---
User wants distribution = the desktop app only (which embeds the Go binary as a sidecar). No standalone binary archives / curl|sh installer.

## Do
- `.github/workflows/release.yml` — remove the GoReleaser `release` job; keep only the Tauri desktop matrix (tauri-action, includeUpdaterJson, Apple signing, TAURI_SIGNING). The matrix job creates the release itself (drop `needs: release`). **Pin pnpm**: add `version: 10` to each `pnpm/action-setup@v4` (failure: "No pnpm version is specified").
- Delete `.goreleaser.yaml` and `install.sh`.
- Docs: remove the "Install the binary" curl|sh section in `installation.md` (keep Desktop app + Build-from-source); fix the CHANGELOG entry to desktop-only. README check.

## Acceptance
- release.yml valid, references no goreleaser/install.sh; building from source still works (`make build`); docs build clean.