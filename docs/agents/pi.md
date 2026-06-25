---
title: Pi
---

# Pi

Pi is an MCP-capable coding agent. Its preferred project MCP config lives in `./.mcp.json` at the repo root.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Pi. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:pi`, then verifies the connection.

## Manual setup

Paste this into `./.mcp.json` at the repo root:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:pi", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

Pi connects as `agent:pi` by default. Edit the identity on the card to run multiple instances (e.g. `agent:pi-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Pi's preferred project config is `./.mcp.json` — the SAME file Claude Code uses (see [/agents/claude](/agents/claude)), so connecting both Pi and Claude writes one shared `cairn` entry, and whichever you connect last sets the `--actor`.
- Pi also reads `~/.pi/agent/mcp.json` and `~/.config/mcp/mcp.json`.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:pi` (the port may differ — cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
