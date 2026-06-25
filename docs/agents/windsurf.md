---
title: Windsurf
---

# Windsurf

[Windsurf](https://windsurf.com) is Codeium's AI editor. Its MCP config lives in the GLOBAL file `~/.codeium/windsurf/mcp_config.json`.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Windsurf. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:windsurf`, then verifies the connection.

## Manual setup

Paste this into `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:windsurf", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

Windsurf connects as `agent:windsurf` by default. Edit the identity on the card to run multiple instances (e.g. `agent:windsurf-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Windsurf has NO project-level config — this file is global, shared across all your projects, so the `--repo` arg pins this entry to one project.
- Re-Connect from another project to repoint it, or add a second entry by hand.
- After writing, refresh MCP servers in Windsurf (Cascade → MCP) or restart.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:windsurf` (the port may differ — cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
