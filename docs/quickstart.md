---
title: Quickstart
---

# Quickstart

From an empty repo to your first observable agent session in about ten minutes.

## 1. Build and open

```sh
make build
cairn web --repo .
```

Open the printed URL. In an uninitialized project you'll see an **Initialize** screen — the
prefix is pre-filled from the folder name. Click **Initialize** and you land on the task board.

## 2. Create a task

Use **New task** in the UI, or create one over the API / MCP. A task needs only a title, but a
useful one carries a short intent body and at least one check:

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
(`backlog` by default). See [Task files & config](/guides/task-files) for the full shape.

## 3. Connect an agent

Open the **Connect** page (sidebar → Connect). cairn lists the agents installed on your
machine; click **Connect** next to one — say Claude Code — and it writes that agent's MCP
config for this project, pointed at the running binary, under the identity `agent:claude`.

Nothing installed that cairn detects? Pick any agent and copy its **Manual setup** snippet, or
see [Connect to Claude](/agents/claude).

## 4. Run the agent loop

Point your agent at the task. A session-aware client drives:

```
identity → list(ready=true) → begin → heartbeat* → run_checks → finish
```

- `begin` claims the task, moves it into the working state, and starts a session.
- `heartbeat` reports concise progress; the UI shows it live (Active / Stalled).
- `finish` ends the session with a review summary and moves the task to review — it does **not**
  close it.

Watch it happen on the board: the task shows a live **Active** signal, the agent's heartbeats,
and finally **Awaiting review**. See [The agent loop](/guides/agent-loop).

## 5. Review and close

Once you're satisfied, transition the task to a closed state (`done`). Closing re-runs the
task's `cmd` checks fresh and **refuses** if any fail — so a stale pass can't slip through. A
`manual` check must be attested first. That's the [checks gate](/guides/checks-and-gates).

::: tip
Everything you just did is files under `.cairn/` — `git diff` to see the task, its provenance
log, and the session record. Commit them like any other change.
:::
