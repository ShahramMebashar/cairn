---
title: Agents
---

# Connecting agents

cairn is an MCP server, so any MCP-capable agent can use its task and session tools. The
**Connect** page in the web UI wires one up in a single click: cairn detects the agents
installed on your machine and writes each one's MCP config for you.

## How auto-connect works

The cairn process runs locally with your permissions, so it writes the config files
itself, by the same code path whether you're in the desktop app or `cairn web` in a browser. For
each agent, Connect:

- **Writes the agent's native config** in its own shape and location: `mcpServers` JSON for
  most, `[mcp_servers]` TOML for Codex, OpenCode's `mcp` block for OpenCode.
- **Merges, never clobbers.** Your other MCP servers and unrelated keys are preserved. The file
  is written atomically and the previous version is backed up to `<file>.bak`.
- **Uses the absolute binary path.** The config launches cairn by its real path (in the desktop
  app, the bundled sidecar), never a bare `cairn` that may not be on `PATH`.
- **Connects over stdio.** The written config is self-contained and works even when the cairn
  app isn't running.
- **Verifies the write**, then flips the card to **Connected**.

A **Disconnect** (the unplug button on a connected card) removes only cairn's entry and leaves
the rest of the file intact.

## Identity

Every agent connects under its **own** identity (`agent:claude`, `agent:cursor`,
`agent:codex`, and so on) so its writes are attributed to it in each task's provenance log,
never to you. Edit the identity on a card to run more than one instance of the same agent
(e.g. `agent:cursor-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Supported agents

| Agent | One-click | Config it writes | Scope |
| --- | --- | --- | --- |
| [Claude Code](/agents/claude) | Yes | `./.mcp.json` | Project |
| [Cursor](/agents/cursor) | Yes | `./.cursor/mcp.json` | Project |
| [Codex](/agents/codex) | Yes | `./.codex/config.toml` | Project |
| [Windsurf](/agents/windsurf) | Yes | `~/.codeium/windsurf/mcp_config.json` | Global |
| [OpenCode](/agents/opencode) | Yes | `./opencode.json` | Project |
| [Kilo Code](/agents/kilo) | Yes | `./.kilocode/mcp.json` | Project |
| [Pi](/agents/pi) | Yes | `./.mcp.json` | Project |
| [Antigravity](/agents/antigravity) | Manual | `~/.gemini/config/mcp_config.json` | Global |

Project-scoped configs point at *this* repo (`--repo` is set to the project path), so they're
safe to commit and share. Windsurf and Antigravity have no project-level config, so cairn
writes them to your home directory.

::: tip Don't see your agent?
Any MCP client works. These are just the ones cairn can configure automatically. Point your
client at `cairn serve --actor agent:<name> --repo <path>` (stdio), or at the running app's
`/mcp?repo=<path>&actor=agent:<name>` endpoint (HTTP). The per-agent pages show both.
:::
