---
id: PROJ-2apddeedrp
title: Add a Check shell setting (config.yaml check_shell + Settings UI)
status: done
priority: medium
checks:
  - desc: Go build/vet/tests pass (shell precedence + config round-trip)
    cmd: make check
    timeout: 300
    result: pass
  - desc: Web build clean (Settings input)
    cmd: pnpm --dir web build
    timeout: 300
    result: pass
  - desc: 'Manual: Settings shows the Check shell field and saving persists to config.yaml'
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-25T14:07:42Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T14:07:48Z', did: began session ses_7481354ae28f33505dd4e89d}
  - {who: 'agent:claude', at: '2026-06-25T14:13:05Z', did: ran checks}
  - {id: n_nee2wvg4, who: 'agent:claude', at: '2026-06-25T14:13:51Z', did: note, text: 'Check shell is now a real per-repo setting. Backend: config.Config gained CheckShell (yaml check_shell,omitempty); check.Runner gained a Shell field and resolveShell is now a method with precedence CAIRN_SHELL env → Runner.Shell (config) → sh; service.go runCmdChecks passes Shell: cfg.CheckShell; store.SaveConfig(cfg) (validates then config.Save); server adds checkShell to statusResp + a POST /api/config { checkShell } handler that load-modify-saves (pointer field → omitted leaves unchanged, "" clears; trims ends + strips newlines, preserves internal spaces for paths). Frontend: api.Status.checkShell + setCheckShell(path,shell); useSetCheckShell hook (updates status cache + toast); SettingsDialog now shows an always-visible "Check shell" input (Input + Save button, dirty-gated) above the desktop-only toggles, wired path + checkShell from App.tsx. Tests: TestRunUsesConfiguredShellWhenEnvUnset + TestEnvShellOverridesConfiguredShell. make check + web build green. Live-verified via cairn web: POST checkShell=bash → status.checkShell=bash and config.yaml has `check_shell: bash`; clearing removes the key (omitempty).'}
  - {who: 'agent:claude', at: '2026-06-25T14:14:06Z', did: finished session ses_7481354ae28f33505dd4e89d, text: "Added a real \"Check shell\" setting in the UI, backed by config.yaml — so the shell is no longer env-only.\n\nBackend:\n- config.Config gains CheckShell (`yaml:\"check_shell,omitempty\"`).\n- check.Runner gains a Shell field; resolveShell is now a method with precedence CAIRN_SHELL env → Runner.Shell (config) → `sh`. service.go's runCmdChecks passes Shell: cfg.CheckShell.\n- store.SaveConfig(cfg) validates then writes config.yaml.\n- server: statusResp gains checkShell; new POST /api/config { checkShell } load-modify-saves config (pointer field — omitted leaves it unchanged, \"\" clears back to sh; trims ends and strips newlines but preserves internal spaces so shell paths like \"C:\\\\Program Files\\\\Git\\\\bin\\\\sh.exe\" survive).\n\nFrontend:\n- api.Status.checkShell + setCheckShell(path, shell); useSetCheckShell hook updates the status cache and toasts.\n- SettingsDialog now shows an always-visible \"Check shell\" input (text field + dirty-gated Save) above the desktop-only toggles; description notes empty=sh, Windows uses Git Bash/WSL, and CAIRN_SHELL overrides. App.tsx passes path + status.checkShell.\n\nTests: TestRunUsesConfiguredShellWhenEnvUnset, TestEnvShellOverridesConfiguredShell. make check + pnpm web build green. Live `cairn web`: POST checkShell=bash → status.checkShell=bash and config.yaml shows `check_shell: bash`; clearing removes the key.\n\nWhere it is now: workspace dropdown → Settings… → \"Check shell\". Precedence at run time: CAIRN_SHELL env (one-off) > this setting (persisted per repo) > sh. The Help dialog's \"Checks run a shell\" card still explains the concept. Manual check (Settings shows + persists) verified via the API end-to-end; a human can confirm the dialog visually after a rebuild."}
  - {who: 'human:shaho', at: '2026-06-25T19:14:40Z', did: attested, text: check 2 pass}
  - {who: 'human:shaho', at: '2026-06-25T19:14:50Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:14:50Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_7481354ae28f33505dd4e89d
---
Surface the check shell as a persistent, per-repo setting (today it's only the CAIRN_SHELL env var + a Help-dialog note).

## Backend
- `internal/config/config.go` — add `CheckShell string` (`yaml:"check_shell,omitempty"`).
- `internal/check/check.go` — Runner gains a `Shell` field; precedence in resolveShell: `CAIRN_SHELL` env → `Runner.Shell` (config) → `sh`.
- `internal/mcp/service.go` (runCmdChecks) — pass `Shell: cfg.CheckShell` when building the runner.
- `internal/store/store.go` — add `SaveConfig(config.Config)`.
- `internal/server` — add `checkShell` to status; `POST /api/config { checkShell }` to set it.

## Frontend
- `api.ts` — Status.checkShell + `setCheckShell(path, shell)`.
- `SettingsDialog.tsx` — a "Check shell" text input (placeholder `sh`); wire path + status from `App.tsx`; invalidate status on save.

## Acceptance
- env > config > default precedence (test); config round-trips check_shell; make check + web build clean; Settings shows/saves the value.