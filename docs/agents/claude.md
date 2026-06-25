---
title: Claude Code
---

# Claude Code

[Claude Code](https://claude.com/claude-code) is Anthropic's official CLI. Its project MCP config lives in `./.mcp.json` at the repo root, project-scoped and shareable.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Claude Code. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:claude`, then verifies the connection.

## Manual setup

Paste this into `./.mcp.json` at the repo root:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:claude", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

The Claude Code CLI also offers a shortcut that writes the same entry:

```sh
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude --repo "$(pwd)"
```

Claude Desktop uses a different file with the same `mcpServers` shape but ABSOLUTE paths:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

## Identity

Claude Code connects as `agent:claude` by default. Edit the identity on the card to run multiple instances (e.g. `agent:claude-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- `.mcp.json` is project-scoped and Claude Code resolves it relative to the repo, so it's safe to commit and share.
- Per-user setups still prefer absolute paths.
- Remove the entry with `claude mcp remove cairn`.
- Pi reads the same `./.mcp.json` file (see [/agents/pi](/agents/pi)), so connecting both writes one shared `cairn` entry.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:claude` (the port may differ; cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
