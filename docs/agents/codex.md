---
title: Codex
---

# Codex

[Codex](https://openai.com/codex) is OpenAI's coding agent. cairn writes a project-scoped config to `./.codex/config.toml`.

## Connect (one-click)

Open cairn's web UI, go to the sidebar **Connect**, and click **Connect** next to Codex. cairn writes the config shown below, launching the cairn binary over stdio with the absolute binary path and `--repo` set to this project, under the identity `agent:codex`, then verifies the connection.

## Manual setup

Paste this into `./.codex/config.toml`:

```toml
[mcp_servers]
  [mcp_servers.cairn]
    args = ['serve', '--actor', 'agent:codex', '--repo', '/absolute/path/to/project']
    command = '/absolute/path/to/bin/cairn'
```

`command` is the absolute path to your built `cairn` binary (the Connect page fills this in automatically; if writing by hand, use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

## Identity

Codex connects as `agent:codex` by default. Edit the identity on the card to run multiple instances (e.g. `agent:codex-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Codex's primary config is the GLOBAL `~/.codex/config.toml`; cairn writes a PROJECT-scoped `./.codex/config.toml`.
- If your Codex build only reads the global file, copy the `[mcp_servers.cairn]` block into `~/.codex/config.toml`.
- The empty `[mcp_servers]` header above is valid TOML.

## Alternative: HTTP transport

While the cairn app is running it also exposes an MCP endpoint at `http://127.0.0.1:7777/mcp?repo=<project>&actor=agent:codex` (the port may differ; cairn prints `CAIRN_WEB_URL=` on startup). URL-based MCP clients can point at that instead of the stdio binary. The tradeoff: the app must be running for the endpoint to respond.
