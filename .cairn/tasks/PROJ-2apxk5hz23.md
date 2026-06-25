---
id: PROJ-2apxk5hz23
title: Add Git-aware task context
status: done
priority: high
labels: [feature, git, sessions, web]
checks:
  - desc: Go checks pass
    cmd: make check
    timeout: 300
    result: pass
  - desc: Web build passes
    cmd: cd web && pnpm build
    timeout: 300
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-25T18:43:49Z', did: created}
  - {id: n_eqr7pyb8, who: 'agent:codex', at: '2026-06-25T18:44:30Z', did: note, text: 'Initial product shape: MVP does not need a new MCP tool. The backend should derive Git context from the local `git` CLI via argument-based exec, store only durable session/check anchors, and expose the evidence to the web UI first.'}
  - {who: 'human:shaho', at: '2026-06-25T18:55:17Z', did: claimed}
  - {id: n_tewfzjfn, who: 'agent:codex', at: '2026-06-25T19:12:03Z', did: note, text: 'Implemented via closed children `PROJ-2apxkgd2my`, `PROJ-2apxkpxre4`, and `PROJ-2apxkyvcqd`: sessions/checks now capture Git anchors, the backend derives Git context over HTTP, and TaskDetail renders a compact Code Context panel. Parent remained assigned to `human:shaho`, so no agent session was claimed on it.'}
  - {who: 'agent:codex', at: '2026-06-25T19:12:08Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:12:19Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-25T19:12:19Z', did: transitioned to done}
assignee: agent:codex
---
Show reviewable Git evidence for agent sessions so humans can see what code changed for a task.

## Scope
- Capture durable Git anchors on sessions and check runs.
- Derive changed files, commits, dirty state, and warnings from local `git`.
- Render a Code Context panel on task/session detail.
- Keep changed files derived at read time; do not store them in task files.

## Acceptance
- A finished session shows branch, start/end commits, changed files, and commit list.
- An active session shows current changes since `headStarted` plus dirty working-tree state.
- Review warnings call out stale checks, missing commits, and uncommitted changes.
- Missing `git` or invalid refs degrade gracefully in the UI.