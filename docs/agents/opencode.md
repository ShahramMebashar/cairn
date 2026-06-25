---
title: OpenCode
---

# OpenCode

[OpenCode](https://opencode.ai) is an open-source terminal coding agent. Its project MCP config lives in `./opencode.json` at the repo root.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to OpenCode. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:opencode`, then verifies the connection.

## Manual setup

Paste this into `./opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "cairn": {
      "type": "local",
      "command": ["/absolute/path/to/bin/cairn", "serve", "--actor", "agent:opencode", "--repo", "/absolute/path/to/project"],
      "enabled": true
    }
  }
}
```

The first element of `command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

OpenCode connects as `agent:opencode` by default. Edit the identity on the card to run multiple instances (e.g. `agent:opencode-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- OpenCode uses a different shape from the others: a top-level `mcp` object, `type: "local"`, and `command` is a single array (binary + args), not separate command/args.
- cairn seeds `$schema` for fresh files.
- OpenCode also reads a global `~/.config/opencode/opencode.json`.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:opencode` (the port may differ; cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
