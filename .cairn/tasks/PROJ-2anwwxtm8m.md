---
id: PROJ-2anwwxtm8m
title: Deep links + clickable notifications + copy-link + copy-as-agent-prompt
status: done
priority: high
labels: [desktop, tauri]
provenance:
  - {who: 'agent:claude', at: '2026-06-25T09:25:49Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T09:25:59Z', did: began session ses_8a1711eab78b64a4798eb851}
  - {who: 'agent:claude', at: '2026-06-25T09:32:35Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-25T09:32:51Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T09:33:06Z', did: finished session ses_8a1711eab78b64a4798eb851, text: "Deep-link slice implemented on branch feat/desktop-tauri (uncommitted).\n\nFrontend: new web/src/lib/connect.ts (taskDeepLink, mcpHttpUrl, mcpAddCommand, agentPromptForTask) — ConnectAgentDialog refactored to use it. TaskDetail \"Task actions\" dropdown gained Copy link (cairn://task/ID?repo=) and Copy as agent prompt (task context + MCP connect command + link). desktop.ts: navigateToTask, openDeepLink (parses cairn://task/<id>?repo= and cairn://open?repo=), onDeepLink listener, and notify(title,body,target?) now wires the notification plugin's onAction so clicking an alert jumps to the task (degrades gracefully). notifications.ts passes {path,id} when firing. App.tsx useDeepLinks() at Flow level routes incoming opens.\n\nRust: registered tauri-plugin-deep-link; tauri.conf plugins.deep-link.desktop.schemes=[\"cairn\"]; capability deep-link:default. setup wires app.deep_link().on_open_url → show_main + emit(\"deep-link\", url); register_all() best-effort; single-instance handler scans argv for cairn:// (Win/Linux running-instance case). Build.rs app-manifest unchanged (update_tray already declared).\n\nVerified: go vet+test, frontend tsc+build, cargo check green; full tauri build → Info.plist carries CFBundleURLSchemes=[cairn]; `open cairn://task/...` and `cairn://open?repo=` route to the running app without error/crash. Manual (GUI) check pending: confirm in-app navigation lands on the task, the two copy actions, and notification-click — left for the user (screenshots blocked here). All desktop bits isTauri-guarded; browser build unaffected. Deep-link per-OS registration on Windows/Linux still to verify on installed builds."}
  - {who: 'human:shaho', at: '2026-06-25T12:09:59Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T13:40:52Z', did: attested, text: check 3 pass}
  - {who: 'human:shaho', at: '2026-06-25T13:41:06Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T13:56:22Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:15:46Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:15:59Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:15:59Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_8a1711eab78b64a4798eb851
checks:
  - desc: Go vet + tests pass
    cmd: go vet ./... && go test ./...
    result: pass
  - desc: Frontend typechecks + builds (connect.ts, copy actions, deep-link hooks)
    cmd: cd web && pnpm build
    result: pass
  - desc: Desktop Rust compiles (deep-link plugin + on_open_url)
    cmd: cd desktop/src-tauri && cargo check
    result: pass
  - desc: 'GUI: deep link opens the task, copy link + copy-as-agent-prompt work, clickable notification jumps to task (cairn:// scheme registered + routes confirmed at OS level)'
    type: manual
    result: pass
rank: !!float 1782384812001
---
Low-debt connectivity slice (no GitHub/external deps). Plan: /Users/shaho/.claude/plans/inherited-nibbling-blossom.md

- cairn:// deep links (task/open) via tauri-plugin-deep-link → navigate; Win/Linux cold-start via single-instance argv.
- Clickable OS notifications → jump to the task.
- Copy link (cairn://task/ID) + Copy as agent prompt (task context + MCP connect line) in TaskDetail.
- Shared web/src/lib/connect.ts helpers (reused by ConnectAgentDialog). All isTauri-guarded.