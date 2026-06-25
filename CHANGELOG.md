# Changelog

All notable changes to cairn are documented in this file. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and this project follows
[Semantic Versioning](https://semver.org/).

cairn is **pre-1.0**: minor versions may include breaking changes until 1.0. The version is
defined in `internal/mcp/sessions.go` (`ServiceVersion`) and `desktop/src-tauri/tauri.conf.json`.

## [Unreleased]

### Added

- **One-click agent integrations.** A new **Connect** page detects installed agents and writes
  their MCP config for you, each under its own identity (`agent:<id>`). Auto-connect for Claude
  Code, Cursor, Codex, Windsurf, OpenCode, Kilo Code, and Pi; a copy-paste guide for Antigravity
  and any other MCP client.
  - `GET /api/connect`, `POST /api/connect/{agent}`, `DELETE /api/connect/{agent}` (Disconnect),
    and `GET /api/connect/{agent}/manual`, backed by `internal/connect`.
  - Config writes are safe-merge: only the `cairn` entry is added/removed, the file is written
    atomically with a `<file>.bak` backup, and the result is verified.
- **Documentation site**: a VitePress site under `docs/` (guides, per-agent pages, and an
  HTTP/MCP/CLI reference), deployed to GitHub Pages.
- **Cross-platform desktop installers.** The release workflow builds the desktop app for macOS
  (`.dmg`, signed + notarized), Windows (NSIS `.exe`), and Linux (`.deb` + AppImage), and
  publishes them to a GitHub Release with a signed `latest.json` so the in-app updater works.
  The Go binary ships embedded as a sidecar (no separate binary download). cairn now also builds
  on Windows (platform-split file lock and check-timeout kill).
- `SECURITY.md` and this changelog.

## [0.1.0]

Initial release.

### Added

- **File-based task graph** under `.cairn/`: tasks as Markdown (YAML frontmatter + prose body),
  engine-assigned collision-free ids, dependencies, and a two-gate transition model (deps gate to
  start, checks gate to close).
- **MCP server** (`cairn serve`) over stdio, plus MCP over Streamable HTTP from `cairn web`.
- **Web UI** (`cairn web`): task board, dependency graph, and live updates over SSE.
- **Observable agent sessions**: `begin`/`heartbeat`/`finish`/`cancel` with stall detection and
  review handoff.
- **Desktop app**: the same binary embedded in a Tauri shell with a live menu-bar tray.
