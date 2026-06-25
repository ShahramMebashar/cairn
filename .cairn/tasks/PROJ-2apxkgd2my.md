---
id: PROJ-2apxkgd2my
title: Capture Git anchors for sessions and check runs
status: done
priority: high
labels: [backend, git, sessions]
parent: PROJ-2apxk5hz23
checks:
  - desc: Go checks pass
    cmd: make check
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-25T18:44:00Z', did: created}
  - {who: 'agent:codex', at: '2026-06-25T18:48:00Z', did: began session ses_821edc849d2bbae823278a40}
  - {id: n_mv35yygg, who: 'agent:codex', at: '2026-06-25T18:48:37Z', did: note, text: 'Implementation direction: add a small internal Git context package and reuse it from session begin/finish/check logging. Session fields already exist, so changed files stay derived and only anchors/check HEAD metadata are persisted.'}
  - {who: 'agent:codex', at: '2026-06-25T18:57:36Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:08:38Z', did: finished session ses_821edc849d2bbae823278a40, text: 'Implemented Git anchor capture: added `internal/gitctx`, auto-filled session begin/finish refs when omitted, recorded check-run `head:` metadata, parsed it in the runs API, and covered the parser/log behavior with tests. Verified with `make check` and `cd web && pnpm build`.'}
  - {who: 'agent:codex', at: '2026-06-25T19:08:49Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:08:49Z', did: transitioned to done}
assignee: agent:codex
active_attempt: att_821edc849d2bbae823278a40
---
Record the Git commits needed to prove what each session and check run observed.

## Scope
- On `begin`, persist branch and `HEAD` when available.
- On `finish`, persist final `HEAD`.
- Add current `HEAD` metadata to command check run logs.
- Use direct `git` exec with timeouts; no shell wrapping.

## Acceptance
- Session records contain stable `branch`, `headStarted`, and `headFinished` values when Git is available.
- Check logs include the commit used for the run.
- Missing Git or non-repo paths do not break session/check workflows.