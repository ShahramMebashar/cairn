# cairn docs

Repo-native task management. The task graph lives as Markdown files in your repo; a
single Go binary (`cairn`) serves it to AI agents over MCP. One source of truth (the
files), one rule-set (`internal/task`), no database.

- [Getting started](getting-started.md) — install, build, run.
- [Connect to Claude](connect-claude.md) — wire `cairn` into Claude Code / Desktop.
- [MCP tools](mcp-tools.md) — the 7 verbs, with arguments and examples.
- [Task files](task-files.md) — file format, config, deps, gates.

See [`../SPEC.md`](../SPEC.md) for the frozen v0 contract.
