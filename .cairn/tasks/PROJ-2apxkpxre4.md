---
id: PROJ-2apxkpxre4
title: Derive Git context for task sessions
status: done
priority: high
labels: [backend, git, api]
parent: PROJ-2apxk5hz23
deps: [PROJ-2apxkgd2my]
checks:
  - desc: Go checks pass
    cmd: make check
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-25T18:44:06Z', did: created}
  - {id: n_xhykw5xp, who: 'agent:codex', at: '2026-06-25T18:54:57Z', did: note, text: 'Backend API slice is being implemented in the same pass: HTTP-only Git context endpoints derive evidence from local Git and use latest check-run HEAD for stale-check warnings. No MCP tool is added for MVP.'}
  - {who: 'agent:codex', at: '2026-06-25T19:08:55Z', did: began session ses_283016e38bca2572c433e59e}
  - {id: n_zvsgpz06, who: 'agent:codex', at: '2026-06-25T19:09:03Z', did: note, text: 'Implemented via `internal/gitctx` and `internal/server/git_context.go`: task-level and session-level HTTP endpoints derive files, commits, dirty state, and warnings from local Git, using latest run-log `head` for stale-check detection.'}
  - {who: 'agent:codex', at: '2026-06-25T19:09:04Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:09:11Z', did: finished session ses_283016e38bca2572c433e59e, text: 'Implemented derived Git context: added `internal/gitctx` for direct `git -C` reads, parsed NUL-delimited name/status output, exposed task/session HTTP context endpoints, and derived dirty/no-change/stale-check warnings from session anchors and latest run-log head. Verified with `make check`.'}
  - {who: 'agent:codex', at: '2026-06-25T19:09:18Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:09:18Z', did: transitioned to done}
assignee: agent:codex
active_attempt: att_283016e38bca2572c433e59e
---
Add the backend evidence layer that derives files, commits, dirty state, and warnings from local Git.

## Scope
- Add an internal Git context package using `exec.CommandContext` and `git -C`.
- Parse machine-readable output (`-z`/stable formats) for changed files and status.
- Expose session-level and task-level HTTP endpoints for Git context.
- Derive warnings for stale checks, missing refs, no commits, and dirty finished work.

## Acceptance
- Finished sessions return `headStarted..headFinished` files and commits.
- Active sessions return current diff/status since `headStarted`.
- Windows paths and filenames with spaces are handled by argument-based exec and NUL parsing.
- Errors are typed enough for the UI to show useful unavailable/degraded states.