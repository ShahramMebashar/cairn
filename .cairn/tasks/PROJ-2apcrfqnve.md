---
id: PROJ-2apcrfqnve
title: Make check shell resolvable (CAIRN_SHELL) with a clear error; document it
status: done
priority: medium
checks:
  - desc: Go build/vet/tests pass incl. the shell-resolution test
    cmd: make check
    timeout: 300
    result: pass
  - desc: Docs build clean
    cmd: npm --prefix docs run build
    timeout: 300
    result: pass
  - desc: Web build clean (HelpDialog change)
    cmd: pnpm --dir web build
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-25T13:56:31Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T13:56:38Z', did: began session ses_a84b2a6564ee0d25c65eac28}
  - {who: 'agent:claude', at: '2026-06-25T13:59:09Z', did: ran checks}
  - {id: n_w65n0erw, who: 'agent:claude', at: '2026-06-25T13:59:20Z', did: note, text: 'Checks are now shell-resolvable with a clear error, documented in all three surfaces. Backend (internal/check/check.go): new resolveShell() reads CAIRN_SHELL (default `sh`), preflights exec.LookPath, and returns one actionable error ("shell %q not found on PATH — install Git Bash/WSL on Windows or set CAIRN_SHELL") instead of a cryptic per-check exec failure; RunContext uses the resolved shell; package doc updated. Tests: TestRunMissingShellIsClearError + TestRunHonorsCairnShell. Default behavior unchanged (sh), so SPEC §6 holds. In-app guide (web/src/components/HelpDialog.tsx): added a "Checks run a shell" section (Terminal icon) — POSIX sh, Windows needs Git Bash/WSL or CAIRN_SHELL. Docs: guides/checks-and-gates.md gained a "## The shell" section (per-OS + CAIRN_SHELL + the exact error text + portability tip); installation.md requirement now links it; reference/cli.md gained an Environment variables table with CAIRN_SHELL. make check + docs build + web build all green.'}
  - {who: 'agent:claude', at: '2026-06-25T13:59:33Z', did: finished session ses_a84b2a6564ee0d25c65eac28, text: "Made the check shell resolvable with a clear error, and documented it in code, the docs site, and the in-app guide.\n\nBackend (internal/check/check.go): new resolveShell() reads CAIRN_SHELL (default `sh`), preflights it with exec.LookPath, and returns a single actionable error — `check: shell \"sh\" not found on PATH — install a POSIX shell (Git Bash or WSL on Windows) or set CAIRN_SHELL to one` — instead of a cryptic per-check \"exec: sh not found\". RunContext now runs `<resolved-shell> -c`. Default stays `sh`, so the SPEC §6 contract is unchanged. Tests added: TestRunMissingShellIsClearError, TestRunHonorsCairnShell.\n\nIn-app guide (web/src/components/HelpDialog.tsx): new \"Checks run a shell\" section (Terminal icon) explaining command checks run in a POSIX shell and that Windows needs Git Bash/WSL or CAIRN_SHELL.\n\nDocs: guides/checks-and-gates.md got a \"## The shell\" section (per-OS guidance, CAIRN_SHELL, the exact error text, and a write-portable-commands tip); installation.md's POSIX-shell requirement now links it; reference/cli.md gained an Environment variables table documenting CAIRN_SHELL.\n\nVerification: make check green (incl. the two new tests), docs build clean, web build clean.\n\nNet: checks compile and run on macOS/Linux/Windows; on Windows they need a POSIX sh (Git Bash/WSL) or CAIRN_SHELL — now with a clear failure and three places that say so. POSIX-only-by-design is preserved (no cmd/PowerShell fork that would fragment check semantics)."}
  - {who: 'human:shaho', at: '2026-06-25T19:15:04Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:15:04Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_a84b2a6564ee0d25c65eac28
---
Checks run via hardcoded `sh -c` (`internal/check/check.go`). Cross-platform at the code level, but on Windows a bare `cmd`/PowerShell box has no `sh`, giving a cryptic per-check `exec: "sh" not found`. Keep `sh` the default (SPEC §6) but make it resolvable and fail clearly.

## Changes
- `internal/check/check.go` — resolve shell from `CAIRN_SHELL` (default `sh`); preflight `exec.LookPath` and return one clear, actionable error if it's missing ("install Git Bash/WSL on Windows, or set CAIRN_SHELL"). Update the package doc. Add a test.
- Docs: `guides/checks-and-gates.md` (shell note + CAIRN_SHELL + Windows), a mention in `installation.md`, and `reference/cli.md` (env var).
- In-app guide: `web/src/components/HelpDialog.tsx` — one line that command checks run in a POSIX shell; Windows needs Git Bash/WSL or `CAIRN_SHELL`.

## Acceptance
- `make check` green incl. new shell test; default still `sh`; docs + web build clean.