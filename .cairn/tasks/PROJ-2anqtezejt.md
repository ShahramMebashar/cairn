---
id: PROJ-2anqtezejt
title: 'Desktop feature wave: hub, notifications, quick-capture, auto-update'
status: in_review
priority: high
labels: [desktop, tauri]
provenance:
  - {who: 'agent:claude', at: '2026-06-25T07:59:10Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T07:59:18Z', did: began session ses_81e639e02eb81065d4206a30}
  - {who: 'agent:claude', at: '2026-06-25T08:14:33Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-25T08:14:48Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T08:15:07Z', did: ran checks}
  - {id: n_yfkcmntr, who: 'agent:claude', at: '2026-06-25T08:15:20Z', did: note, text: 'All four milestones implemented + compiling. Verified: go vet/test green, frontend tsc+build green, Rust cargo check green, full pnpm tauri:build produced Cairn.app, launched it → all 9 plugins init without panic, server/UI/MCP work, Quit kills sidecar. Two fixes during build: process permission is process:allow-restart (not allow-relaunch); local bundle with createUpdaterArtifacts needs TAURI_SIGNING_PRIVATE_KEY env (used the generated dev key). CI switched to tauri-apps/tauri-action (matrix) — robustly builds/signs/creates release + generates latest.json; needs repo secrets TAURI_SIGNING_PRIVATE_KEY + _PASSWORD. Updater keypair generated to scratchpad (NOT committed); pubkey in tauri.conf. DMG packaging still headless-only locally. Interactive behaviors (tray/close-to-tray/shortcut/menu/notifications/autostart/update) need GUI + a published release for full confirmation — left as the manual check.'}
  - {who: 'agent:claude', at: '2026-06-25T08:15:36Z', did: finished session ses_81e639e02eb81065d4206a30, text: "Shipped the desktop feature wave (4 milestones) on branch feat/desktop-tauri (uncommitted).\n\nM1 Always-on hub: single-instance (focus existing, no duplicate sidecar), system tray (Open/New Task/Settings/Quit + left-click show), close-to-tray (X hides, server+MCP stay up), opt-in Launch-at-login (autostart plugin), set_tray_badge command driven by a useTrayBadge hook (awaiting-review count).\nM2 Native notifications: notifications.ts mirrors ready/failed/awaiting_review to OS notifications when the window is unfocused (new \"review\" kind + ScanEye icon); isTauri-guarded; in-app bell unchanged.\nM3 Quick-capture + menu + window-state: global Cmd/Ctrl+Shift+K opens a capture window (#capture route → CaptureView, last project + switcher, Enter creates, Esc closes); native menu bar with Edit submenu (macOS copy/paste) + accelerators emitting menu:* events consumed by useDesktopMenu; window geometry persists.\nM4 Auto-updater: tauri-plugin-updater + process; tauri.conf createUpdaterArtifacts + plugins.updater (GitHub Releases latest.json endpoint + pubkey); useUpdater checks on launch + Settings button → install+relaunch. release.yml rewritten to tauri-apps/tauri-action (matrix mac arm/x64, linux, windows) which builds/signs/creates the release + generates latest.json.\n\nNew: web/src/lib/desktop.ts (+ desktop-hooks.ts), SettingsDialog, CaptureView, switch component; edits to lib.rs, Cargo.toml, tauri.conf.json, capabilities, notifications.ts, NotificationBell, AppSidebar, App.tsx, AGENTS.md.\n\nVerified: go vet+test, frontend tsc+build, cargo check all green; full tauri build → Cairn.app launches, all 9 plugins init, server/UI/MCP work, Quit kills sidecar; browser/dev build unaffected (all guarded).\n\nOpen items: NOT committed. Repo secrets TAURI_SIGNING_PRIVATE_KEY + _PASSWORD must be added for releases (dev keypair generated to scratchpad, pubkey in tauri.conf — private key NOT committed and must be saved by the user). Apple/Windows code signing still deferred. Interactive behaviors + real auto-update need GUI/a published release to fully confirm (manual check left pending). Local DMG packaging needs a GUI runner (CI handles it)."}
assignee: agent:claude
active_attempt: att_81e639e02eb81065d4206a30
checks:
  - desc: Go vet + tests pass
    cmd: go vet ./... && go test ./...
    result: pass
  - desc: Frontend typechecks + builds (all new desktop components, isTauri-guarded)
    cmd: cd web && pnpm build
    result: pass
  - desc: Desktop Rust shell compiles with all plugins (tray/menu/shortcut/window-state/autostart/notification/updater/process)
    cmd: cd desktop/src-tauri && cargo check
    result: pass
  - desc: Interactive desktop behaviors (tray click, close-to-tray, Cmd+Shift+K capture, native menu, OS notifications, autostart toggle, update install) — app launches + all plugins init verified; full confirmation needs GUI interaction + a published release
    type: manual
    result: pending
---
Make the desktop form factor pay off: turn cairn into an always-on local hub for humans + agents.

Plan: /Users/shaho/.claude/plans/inherited-nibbling-blossom.md

Four sequenced milestones:
1. **Always-on hub** — tray + close-to-tray + single-instance + opt-in autostart.
2. **Native OS notifications** — fire on ready/failed/awaiting_review even when hidden; click focuses task.
3. **Quick-capture + native menu + window-state** — global CmdOrCtrl+Shift+K capture window (last project + switcher), menu bar w/ accelerators, persisted window geometry.
4. **Auto-updater** — GitHub Releases feed + Tauri update key; latest.json job in release.yml.

Decisions: close→hide-to-tray; quick-add→last project+switcher; updater→GitHub Releases + we hold the key; autostart opt-in. All desktop features isTauri()-guarded so browser/dev build is unaffected.