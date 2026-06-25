---
title: Introduction
---

# Introduction

cairn is repo-native task management. The task graph lives as Markdown files in your
repository; a single Go binary (`cairn`) serves it to AI agents over MCP and to humans over
a web UI. One source of truth — the files. One rule-set — `internal/task`. No database.

If you have ever wanted your coding agents and your team to work the same backlog, with a
real audit trail of who did what, cairn is that: a task is a file, a change is a commit, and
every write is stamped with the actor that made it.

```sh
make build                                  # -> bin/cairn
cairn web --repo .                          # open the board in your browser
# or wire up an agent:
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude-1 --repo "$(pwd)"
```

## Why files

A task is Markdown — **YAML frontmatter** the engine owns (id, status, checks, provenance)
plus a **prose body** the engine never touches after creation. That choice buys a lot:

- **Versioned by git.** Tasks branch, merge, and review like code. A task's history *is* its
  git history.
- **No drift.** The web UI and the MCP server are thin adapters over the same engine, so an
  agent and a human always see the same rules — the same gates, the same readiness.
- **Portable & collision-free.** Ids are minted as `PREFIX` + a time-ordered base32 token, so
  two people creating tasks in separate clones never collide and never merge-conflict.

## How agents fit

cairn speaks [MCP](https://modelcontextprotocol.io) (the Model Context Protocol). An agent
connects, then drives a short, observable loop:

1. **identity** — confirm the actor this connection writes as.
2. **list** ready work — `list(ready=true, status=backlog)` answers "what can I start now".
3. **begin** a session — claims the task, enters the working state, starts heartbeating.
4. **build + heartbeat** — make the change, report concise progress.
5. **note** decisions — short provenance breadcrumbs as you go.
6. **run_checks** — the task's checks must pass before handoff.
7. **finish** — end the session with a review summary (this requests review; it does not
   close the task).

See [The agent loop](/guides/agent-loop) for the full lifecycle and [Sessions](/guides/sessions)
for how work is made observable.

## What you get

- **A task graph** with dependencies, readiness, and a [two-gate](/guides/checks-and-gates)
  transition model (deps must be closed to start; checks must pass to close).
- **Observable agent sessions** — heartbeats, stall detection, and review handoff, so you can
  supervise agents instead of trusting a self-report.
- **One-click agent integrations** — the [Connect page](/agents/) detects installed agents and
  writes their MCP config for you, each under its own identity.
- **A desktop app** — the same binary embeds the UI and ships as a Tauri app with a live tray.

::: tip First time?
Start with [Installation](/installation), then the [Quickstart](/quickstart) walks you from an
empty repo to your first agent session in about ten minutes.
:::
