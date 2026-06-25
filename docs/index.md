---
layout: home

hero:
  name: cairn
  text: Shared task tracking for you and your agents
  tagline: Your tasks are Markdown files in your repo. cairn serves them to AI agents over MCP and to you over a web UI. One source of truth, no database.
  image:
    light: /logo.svg
    dark: /logo-dark.svg
    alt: cairn
  actions:
    - theme: brand
      text: Download
      link: https://github.com/ShahramMebashar/cairn/releases/latest
    - theme: alt
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
    details: A task is a Markdown file. YAML on top for the machine, prose below for people. Diff it, branch it, review it in a PR. A task's history is its git history.
  - title: Made for agents
    details: A single Go binary serves the task graph over MCP. Agents find ready work, claim it, report progress with heartbeats, and hand off for review. Every write is signed with who made it.
  - title: One rule-set
    details: The web UI and the agent API are two front-ends over the same engine. You and your agent always see the same state and the same rules.
  - title: Connect an agent in one click
    details: cairn finds the agents you have installed (Claude, Cursor, Codex, Windsurf, and more) and writes their config for you. Each connects as itself.
  - title: Run it your way
    details: A desktop app with a tray, a browser UI, or a headless server. Download it for macOS, Windows, or Linux, or run cairn web.
---

![The cairn task board, dependency graph, and Connect page](/app-screenshot.png)

## Agents

cairn writes each agent's config for you from the **Connect** page, or you copy a snippet and
paste it yourself. Each agent connects under its own name (`agent:cursor`, `agent:codex`, and
so on), so its work shows up as its own in the task history.

| Agent | One-click | Config it writes | Connects as |
| --- | --- | --- | --- |
| [Claude Code](/agents/claude) | Yes | `./.mcp.json` | `agent:claude` |
| [Cursor](/agents/cursor) | Yes | `./.cursor/mcp.json` | `agent:cursor` |
| [Codex](/agents/codex) | Yes | `./.codex/config.toml` | `agent:codex` |
| [Windsurf](/agents/windsurf) | Yes | `~/.codeium/windsurf/mcp_config.json` | `agent:windsurf` |
| [OpenCode](/agents/opencode) | Yes | `./opencode.json` | `agent:opencode` |
| [Kilo Code](/agents/kilo) | Yes | `./.kilocode/mcp.json` | `agent:kilo` |
| [Pi](/agents/pi) | Yes | `./.mcp.json` | `agent:pi` |
| [Antigravity](/agents/antigravity) | Manual | `~/.gemini/config/mcp_config.json` | `agent:antigravity` |

Any MCP client works. These are the ones cairn sets up for you. See the
[agents guide](/agents/) for how that works, or [Claude Code](/agents/claude) to start.
