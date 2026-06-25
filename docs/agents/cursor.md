---
title: Cursor
---

# Cursor

[Cursor](https://cursor.com) is an AI code editor. Its project MCP config lives in `./.cursor/mcp.json`.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Cursor. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:cursor`, then verifies the connection.

## Manual setup

Paste this into `./.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:cursor", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

Cursor connects as `agent:cursor` by default. Edit the identity on the card to run multiple instances (e.g. `agent:cursor-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Cursor also supports a global `~/.cursor/mcp.json`; cairn writes the project one so it stays scoped to this repo.
- Enable and verify the server under Cursor Settings → MCP.
- Project config takes precedence over the global one.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:cursor` (the port may differ — cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
