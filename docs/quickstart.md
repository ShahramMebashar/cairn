---
title: Quickstart
---

# Quickstart

From an empty repo to your first agent session in about ten minutes.

## 1. Build and open

```sh
make build
cairn web --repo .
```

Open the printed URL. In a fresh project you'll see an **Initialize** screen with the prefix
filled in from the folder name. Click **Initialize** and you land on the board.

## 2. Create a task

Use **New task** in the UI, or create one over the API or MCP. A task needs only a title, but a
good one has a short body and at least one check:

```json
{
  "title": "Add idempotency keys to the webhook",
  "body": "Dedupe duplicate gateway deliveries.",
  "checks": [
    { "desc": "tests pass", "cmd": "go test ./internal/payments" }
  ]
}
```

The engine assigns the id (`PREFIX-<base32>`) and starts the task in your `initial` state
(`backlog` by default). See [Task files and config](/guides/task-files) for the full shape.

## 3. Connect an agent

Open the **Connect** page (sidebar, then Connect). cairn lists the agents installed on your
machine. Click **Connect** next to one, say Claude Code, and it writes that agent's config for
this project, pointed at the running binary, under the name `agent:claude`.

cairn doesn't detect your agent? Pick any agent and copy its **Manual setup** snippet, or see
[Connect to Claude](/agents/claude).

## 4. Run the agent loop

Point your agent at the task. A session-aware client runs:

```
identity → list(ready=true) → begin → heartbeat* → run_checks → finish
```

- `begin` claims the task, moves it to the working state, and starts a session.
- `heartbeat` reports progress. The UI shows it live (Active or Stalled).
- `finish` ends the session with a review summary and moves the task to review. It does **not**
  close the task.

Watch it on the board: the task shows a live **Active** signal, the agent's heartbeats, and
then **Awaiting review**. See [the agent loop](/guides/agent-loop).

## 5. Review and close

When you're happy, move the task to a closed state (`done`). Closing re-runs the `cmd` checks
and refuses if any fail, so a stale pass can't slip through. A `manual` check has to be attested
first. That's the [checks gate](/guides/checks-and-gates).

::: tip
Everything you just did is files under `.cairn/`. Run `git diff` to see the task, its history,
and the session record, then commit them like any other change.
:::
