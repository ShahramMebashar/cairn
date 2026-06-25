---
id: PROJ-2apepkeznc
title: 'Wire Tauri updater: regenerate keypair + generate latest.json feed'
status: in_review
priority: high
checks:
  - desc: release.yml + tauri.conf.json valid (yaml/json parse)
    cmd: ruby -ryaml -e "YAML.load_file('.github/workflows/release.yml')" && node -e "require('./desktop/src-tauri/tauri.conf.json')" && echo OK
    timeout: 60
    result: pass
  - desc: 'Manual: tagged release produces installers + latest.json; in-app update works'
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-25T14:29:39Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T14:29:47Z', did: began session ses_39ca85c3d215adba8faf8196}
  - {who: 'agent:claude', at: '2026-06-25T14:31:27Z', did: ran checks}
  - {id: n_09r65318, who: 'agent:claude', at: '2026-06-25T14:31:46Z', did: note, text: 'Updater wired. Regenerated the Tauri updater keypair (empty password) via `tauri signer generate` into the scratchpad; updated desktop/src-tauri/tauri.conf.json plugins.updater.pubkey to the new public key (verified MATCH vs the generated .pub). Private key stays in scratchpad (outside the repo) for the user to load into the secret. Rewrote .github/workflows/release.yml: kept the GoReleaser `release` job (Go binary + install.sh) and replaced the 3 raw `tauri build` desktop jobs with a single tauri-action matrix (macos-14 aarch64 + x86_64, ubuntu-24.04, windows-latest), needs: release, includeUpdaterJson: true → generates + merges latest.json across the matrix and uploads to the release the updater endpoint points at. Apple signing env mirrors the user''s already-set secrets (APPLE_CERTIFICATE/_PASSWORD, APPLE_SIGNING_IDENTITY, APPLE_TEAM_ID, APPLE_API_KEY=key id, APPLE_API_ISSUER, APPLE_API_KEY_PATH; the AuthKey .p8 is written from APPLE_API_KEY secret content, guarded on macOS+cert present). TAURI_SIGNING_* passed for .sig signing. release.yml + tauri.conf.json parse clean. User still needs: set TAURI_SIGNING_PRIVATE_KEY (from scratchpad/cairn-updater.key) + TAURI_SIGNING_PRIVATE_KEY_PASSWORD (empty). Windows code signing deferred.'}
  - {who: 'agent:claude', at: '2026-06-25T14:32:11Z', did: finished session ses_39ca85c3d215adba8faf8196, text: "Closed the auto-update gap. Regenerated the Tauri updater keypair and switched the desktop release to generate the latest.json feed.\n\n- Generated a fresh updater keypair (empty password) and updated desktop/src-tauri/tauri.conf.json plugins.updater.pubkey to the new public key (verified it matches the generated .pub). The private key is in the session scratchpad (outside the repo) for the user to load into a secret.\n- Rewrote .github/workflows/release.yml: kept the GoReleaser `release` job (binary + install.sh) and replaced the three raw `tauri build` jobs with one tauri-action matrix (macOS arm64+x86_64, Linux, Windows), needs: release, includeUpdaterJson: true. tauri-action now generates AND merges latest.json across the matrix and uploads it to the release the updater endpoint points at — which the per-OS approach couldn't. Apple signing env mirrors the user's already-set secrets and is guarded so unsigned builds still work; TAURI_SIGNING_* passed for .sig signing.\n\nVerified: tauri.conf.json pubkey == generated pubkey; release.yml + tauri.conf.json parse clean.\n\nUser action remaining (only they can): set two repo secrets —\n  gh secret set TAURI_SIGNING_PRIVATE_KEY < <scratchpad>/cairn-updater.key\n  gh secret set TAURI_SIGNING_PRIVATE_KEY_PASSWORD --body \"\"\nThen push a tag (e.g. v0.1.0): GoReleaser publishes the binary; the tauri-action matrix attaches signed+notarized macOS installers, Linux/Windows installers, and latest.json → in-app auto-update works.\n\nDeferred: Windows code signing (separate cert; until then Windows shows a SmartScreen warning). The empty updater-key password keeps it to one secret; a password can be added later by regenerating."}
assignee: agent:claude
active_attempt: att_39ca85c3d215adba8faf8196
---
Close the auto-update gap. Apple secrets are set by the user; Windows signing deferred.

## Do
- Regenerate the Tauri updater keypair (the repo has a pubkey but the private key is lost); update `desktop/src-tauri/tauri.conf.json` `plugins.updater.pubkey`. User sets `TAURI_SIGNING_PRIVATE_KEY` (+ password) secrets from generated files.
- Rewrite `.github/workflows/release.yml` desktop part: replace the 3 raw-`tauri build` jobs with a single **tauri-action matrix** (macOS arm64+x86_64, Linux, Windows) using `includeUpdaterJson: true` so `latest.json` is generated + merged and uploaded to the release. Keep the GoReleaser `release` job for the Go binary. Pass Apple signing env (guarded) + TAURI_SIGNING_*.

## Acceptance
- New pubkey in tauri.conf; release.yml valid; `make desktop` (or a dispatch) still bundles; updater feed produced on tag.