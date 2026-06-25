---
title: Introduction
---

# Introduction

cairn tracks tasks as Markdown files in your repository. A single Go binary serves those files
to AI agents over MCP and to people over a web UI. The files are the source of truth. There is
no database.

If you want your coding agents and your team working the same backlog, with a clear record of
who did what, that is cairn. A task is a file. A change is a commit. Every write is signed with
the actor that made it.

```sh
cairn web --repo .        # open the board in your browser
# or wire up an agent:
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude --repo "$(pwd)"
```

## Why files

A task is Markdown. The YAML frontmatter belongs to the engine (id, status, checks, history).
The prose body belongs to you, and the engine never touches it after you create the task. That
buys a few things:

- **Git owns the history.** Tasks branch, merge, and review like code. A task's history is its
  git history.
- **No drift.** The web UI and the agent API run on the same engine, so an agent and a person
  see the same rules and the same readiness.
- **No collisions.** Ids are a prefix plus a time-ordered token, so two people creating tasks in
  separate clones never clash and never merge-conflict.

## How agents fit

cairn speaks [MCP](https://modelcontextprotocol.io), the Model Context Protocol. An agent
connects and runs a short loop you can watch:

1. **identity**: confirm the name this connection writes as.
2. **list** ready work: `list(ready=true, status=backlog)` answers "what can I start now".
3. **begin** a session: claim the task, move it to the working state, start a heartbeat.
4. **build and heartbeat**: do the work, report progress in a line or two.
5. **note** decisions: leave short breadcrumbs as you go.
6. **run_checks**: the task's checks have to pass before handoff.
7. **finish**: end the session with a review summary. This asks for review. It does not close
   the task.

See [the agent loop](/guides/agent-loop) for the full lifecycle and [sessions](/guides/sessions)
for how the work stays observable.

## What you get

- A task graph with dependencies, readiness, and [two gates](/guides/checks-and-gates): deps
  must be closed to start, checks must pass to close.
- Observable agent sessions, with heartbeats, stall detection, and a review handoff, so you can
  supervise an agent instead of trusting its word.
- One-click agent setup from the [Connect page](/agents/).
- A desktop app that bundles all of this with a tray.

::: tip First time?
Start with [installation](/installation), then the [quickstart](/quickstart) takes you from an
empty repo to your first agent session in about ten minutes.
:::
