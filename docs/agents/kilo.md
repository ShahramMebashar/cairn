---
title: Kilo Code
---

# Kilo Code

[Kilo Code](https://kilocode.ai) is an open-source VS Code AI agent. Its project MCP config lives in `./.kilocode/mcp.json`.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Kilo Code. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:kilo`, then verifies the connection.

## Manual setup

Paste this into `./.kilocode/mcp.json`:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:kilo", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

Kilo Code connects as `agent:kilo` by default. Edit the identity on the card to run multiple instances (e.g. `agent:kilo-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Kilo creates `.kilocode/mcp.json` if missing, and project config takes precedence over its global VS Code extension settings.
- You can also edit it via Kilo's MCP panel ("Edit Project MCP").

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:kilo` (the port may differ — cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
