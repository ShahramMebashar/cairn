---
layout: home

hero:
  name: cairn
  text: Task management your agents and you share
  tagline: A repo-native task graph as plain Markdown files, served to AI agents over MCP and to humans over a clean web UI — one source of truth, one rule-set, no database.
  actions:
    - theme: brand
      text: Get started
      link: /introduction
    - theme: alt
      text: Connect an agent
      link: /agents/
    - theme: alt
      text: View on GitHub
      link: https://github.com/ShahramMebashar/cairn
features:
  - title: Lives in your repo
    details: Tasks are Markdown files under .cairn/ — YAML frontmatter for the machine, a prose body for humans. Diff them, branch them, review them in a PR. No external service.
  - title: Built for agents
    details: A single Go binary serves the graph over MCP (stdio or HTTP). Agents list ready work, claim it, run an observable session with heartbeats, and hand off for review.
  - title: One rule-set, two front-ends
    details: The web UI and the MCP server are thin adapters over the same engine in internal/task — the rules can't drift between what an agent sees and what you see.
  - title: Connect any agent in one click
    details: The Connect page detects installed agents (Claude, Cursor, Codex, Windsurf, OpenCode, Kilo, Pi…) and writes their MCP config for you — each under its own identity.
---

## Supported agents

cairn writes each agent's MCP config for you from the **Connect** page (or you can copy the
snippet and do it by hand). Every agent connects under its own identity (`agent:<id>`) so its
work is attributed correctly in provenance.

| Agent | One-click | Config it writes | Default identity |
| --- | --- | --- | --- |
| [Claude Code](/agents/claude) | Yes | `./.mcp.json` (project) | `agent:claude` |
| [Cursor](/agents/cursor) | Yes | `./.cursor/mcp.json` (project) | `agent:cursor` |
| [Codex](/agents/codex) | Yes | `./.codex/config.toml` (project) | `agent:codex` |
| [Windsurf](/agents/windsurf) | Yes | `~/.codeium/windsurf/mcp_config.json` (global) | `agent:windsurf` |
| [OpenCode](/agents/opencode) | Yes | `./opencode.json` (project) | `agent:opencode` |
| [Kilo Code](/agents/kilo) | Yes | `./.kilocode/mcp.json` (project) | `agent:kilo` |
| [Pi](/agents/pi) | Yes | `./.mcp.json` (project) | `agent:pi` |
| [Antigravity](/agents/antigravity) | Manual | `~/.gemini/config/mcp_config.json` (global) | `agent:antigravity` |

Any MCP client works — these are just the ones cairn can wire up automatically. See the
[Agents overview](/agents/) for how auto-connect works, or [Connect to Claude](/agents/claude)
to start.
