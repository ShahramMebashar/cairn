---
id: PROJ-2apfw153n2
title: Windows code signing via SignPath Foundation (OSS)
status: backlog
priority: medium
labels: [blocked, release]
provenance:
  - {who: 'agent:claude', at: '2026-06-25T14:49:37Z', did: created}
---
Sign the Windows installer with a free SignPath Foundation cert (cairn is public OSS). Approval-gated, so split into "apply now" + "wire CI on approval".

## Done
- Added `LICENSE` (MIT) — required for SignPath OSS eligibility and fixes the README link.

## Blocked on approval (user)
1. Commit + push the LICENSE.
2. Apply to **SignPath Foundation** (signpath.org → Open Source) with cairn's GitHub URL. Approval ~days.
3. After approval, in SignPath console: create Project `cairn`, an Artifact Configuration for the NSIS `.exe`, and a Signing Policy `release-signing`. Note **Organization ID**, **Project slug**, **Policy slug**.
4. GitHub: add secret `SIGNPATH_API_TOKEN`; add `SIGNPATH_ORGANIZATION_ID` / `SIGNPATH_PROJECT_SLUG` / `SIGNPATH_POLICY_SLUG` (vars or secrets).

## CI to implement (me, once secrets exist)
Windows job becomes: build **unsigned** NSIS → `signpath/github-action-submit-signing-request` to sign the `.exe` → **regenerate the Tauri updater `.sig` over the SIGNED installer** (`tauri signer sign`) → upload signed `.exe` + `.sig` and merge into `latest.json`. Key caveat: the updater signature must be computed AFTER Authenticode signing, so Windows can't use tauri-action's inline updater-json for the signed artifact — it needs this post-sign step. Gate the whole thing on `SIGNPATH_API_TOKEN` so releases keep working unsigned until then.

## Acceptance
- Tagged release produces a SignPath-signed Windows installer whose Tauri updater signature verifies; no SmartScreen "unknown publisher" (reputation builds over time).