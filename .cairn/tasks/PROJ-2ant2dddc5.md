---
id: PROJ-2ant2dddc5
title: 'Live tray hub: status counts, task/agent lists, quick actions, project switcher + DND'
status: in_review
priority: high
labels: [desktop, tauri, tray]
provenance:
  - {who: 'agent:claude', at: '2026-06-25T08:37:33Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T08:37:40Z', did: began session ses_0275164806eb8ffcd2ce900a}
  - {who: 'agent:claude', at: '2026-06-25T08:43:55Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-25T08:44:11Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T08:44:25Z', did: finished session ses_0275164806eb8ffcd2ce900a, text: "Live tray hub implemented on branch feat/desktop-tauri (uncommitted).\n\nRust (lib.rs): replaced set_tray_badge with a generic update_tray(TrayMenu{tooltip,title,sections:[[TrayItem{id,label,checked?,enabled?}]]}) command that rebuilds the tray menu (CheckMenuItem for toggles, disabled rows for headers) + sets tooltip and macOS menubar title. handle_menu now emits a single tray:menu event with the item id for all dynamic items (showing the window except for toggle:*); tray_open/tray_quit handled natively.\n\nFrontend: api.listSessions + useSessions(path) (project-wide sessions, live via existing SSE session-changed invalidation). desktop.ts: updateTray, onTrayEvent, DND (dndEnabled/setDnd; notify() now respects DND). useTrayMenu(path, handlers) builds the model from useTasks (counts: awaiting_review/active/stalled/ready; awaiting-review task list) + useSessions (active/stalled agent rows with live progress + relative heartbeat, stalled flagged) + project switcher (when >1 workspace) + DND toggle + Open/Settings/Quit; debounced 250ms and diffed against last JSON; dispatches tray:menu → openTask/openFilter/switchProject/newTask/openSettings/toggle-DND. Wired in App.tsx Workspace (replaced useTrayBadge).\n\nVerified: go vet+test, frontend tsc+build, cargo check all green; full tauri build → Cairn.app launches and runs (server/MCP healthy). Manual check pending: the live tray content + clicks + project switcher + DND need a project open in a GUI session to confirm (screenshots blocked in this env; app left running for the user to inspect). DMG packaging still headless-only locally."}
assignee: agent:claude
active_attempt: att_0275164806eb8ffcd2ce900a
checks:
  - desc: Go vet + tests pass
    cmd: go vet ./... && go test ./...
    result: pass
  - desc: Frontend typechecks + builds
    cmd: cd web && pnpm build
    result: pass
  - desc: Desktop Rust compiles (update_tray renderer)
    cmd: cd desktop/src-tauri && cargo check
    result: pass
  - desc: 'Live tray verified in GUI: counts/tasks/agents populate + update via SSE, clicks navigate, project switcher + DND work (needs a project open + GUI)'
    type: manual
    result: pending
---
Make the tray a real-time hub. Plan: /Users/shaho/.claude/plans/inherited-nibbling-blossom.md

Rust becomes a generic tray renderer (update_tray command takes a menu model; clicks emit tray:menu with the item id). Frontend owns content: useTrayMenu builds live model from useTasks (counts + awaiting-review/active task lists) + useSessions (live agent progress), debounced/diffed, and dispatches tray:menu events (open task, open filter, switch project, quick-add, DND toggle). DND + attention badge included. All isTauri-guarded.