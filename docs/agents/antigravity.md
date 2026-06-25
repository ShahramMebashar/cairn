---
title: Antigravity
---

# Antigravity

[Antigravity](https://antigravity.google) is Google's agentic IDE. Its MCP config lives in the GLOBAL file `~/.gemini/config/mcp_config.json`, shared with the Gemini CLI.

## Manual setup only

cairn does NOT auto-write Antigravity's config. Its exact config path and local/stdio schema aren't yet confirmed from official docs, so cairn shows this guide instead of writing the file. On the Connect page, Antigravity carries a **Manual** badge and has no button.

Add the `cairn` entry yourself:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/absolute/path/to/bin/cairn",
      "args": ["serve", "--actor", "agent:antigravity", "--repo", "/absolute/path/to/project"]
    }
  }
}
```

`command` is the absolute path to your built `cairn` binary (use your `bin/cairn` absolute path), and `--repo` is this project's absolute path.

Open the file via Antigravity → Manage MCP Servers → View raw config, paste the `cairn` entry, and save. Antigravity reloads automatically.

## Identity

Antigravity connects as `agent:antigravity` by default. Edit the `--actor` value to run multiple instances (e.g. `agent:antigravity-2`). See [Sessions](/guides/sessions) for how identity is enforced.

## Gotchas

- Antigravity shares config with the Gemini CLI; the `mcpServers` shape is confirmed, but the exact file path (`~/.gemini/config/mcp_config.json`) comes from a single source — verify against your install.
- Antigravity uses `serverUrl` for REMOTE servers, but cairn is a LOCAL stdio server, so it uses `command`/`args` as shown.
