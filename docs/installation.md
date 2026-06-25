---
title: Installation
---

# Installation

## Requirements

- **Go 1.25+** to build the binary.
- **POSIX shell (`sh`)** to run task checks. macOS and Linux work out of the box; on Windows
  use WSL or Git Bash, or set [`CAIRN_SHELL`](/guides/checks-and-gates#the-shell) to a shell on
  your `PATH`.
- **Node 18+** only if you want to build the web UI, the desktop app, or these docs from source.
- **Rust toolchain** only to build the desktop app (Tauri).

## Desktop app

The easiest way to run cairn is the desktop app — a small cross-platform window with a live
menu-bar tray that bundles the web UI and the `cairn` binary together. It's the same thing
`cairn web` serves, just native, and it can configure agents, fire notifications, and stay in
your tray.

::: warning Prebuilt installers — placeholder
Signed installers aren't published yet. Build from source for now (below). Once the first
version is tagged, installers for macOS, Linux, and Windows will appear on the
[Releases page](https://github.com/ShahramMebashar/cairn/releases) and update themselves in
place.
:::

Build the installer for your OS:

```sh
make desktop       # -> native installer (.dmg / .AppImage / .msi)
```

This packages the app with [Tauri](https://tauri.app): it builds the web UI, compiles the
`cairn` binary as a bundled **sidecar**, and wraps both in a native shell. To iterate without
packaging:

```sh
make desktop-dev   # native window against a dev server (run `make web` alongside)
```

Prefer the command line or a headless server? Build the `cairn` binary from source and run it
directly (the desktop app bundles this same binary as a sidecar).

## Build the binary

```sh
make build        # -> bin/cairn
```

Or directly:

```sh
go build -o bin/cairn ./cmd/cairn
```

That single binary is everything: the MCP server (`cairn serve`), the web server
(`cairn web`), and the workspace scaffolder (`cairn init`).

## Initialize a repo

cairn stores everything under `.cairn/` at your repo root:

```
.cairn/
  config.yaml     # prefix, states, gates — see Task files & config
  tasks/          # one Markdown file per task, filename = id
  runs/           # check-run logs (gitignored)
```

You don't create these by hand. `init` scaffolds them, and it's idempotent — an existing
`config.yaml` is left untouched; it only fills in missing dirs and the `.gitignore` entry.
There are three ways in, all calling the same code so they can't drift:

```sh
cairn init --repo /path/to/project              # prefix derived from the folder name
cairn init --repo /path/to/project --prefix MYP # explicit prefix
```

- **Auto:** `cairn serve` initializes the workspace on first run if it's missing.
- **Web:** open `cairn web` in an uninitialized project and click **Initialize** (the prefix is
  pre-filled from the folder name, uppercased: `web-app` → `WEBAPP`).

## Run the web UI

The browser front-end and the desktop app are both `cairn web`:

```sh
cairn web --repo .
```

It prints the URL it bound to (`CAIRN_WEB_URL=http://127.0.0.1:7777`, or the next free port)
and serves the task board, the graph, and the **Connect** page. For development with live
reload:

```sh
make web          # Go HTTP server on :8080 (serves /api)
make web-dev      # Vite dev server; proxies /api -> :8080
```

## Connect an agent

`cairn serve` is the MCP server; it's normally launched *by* an MCP client, not run by hand.
The easiest path is the [Connect page](/agents/) in the web UI — one click writes the agent's
config. To do it by hand, see the per-agent pages, e.g. [Claude Code](/agents/claude):

```sh
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude-1 --repo "$(pwd)"
```

## Test

```sh
make check        # gofmt + go vet + go test ./...
```

## Quick manual smoke

The best interactive tool is the MCP Inspector — it launches the binary and holds the
connection open:

```sh
npx @modelcontextprotocol/inspector ./bin/cairn serve --actor agent:dev --repo .
```

Then call `identity`, `list` (try `ready: true`), `begin`, `heartbeat`, and `finish` from the
GUI.

::: warning Don't pipe one-shot JSON into the binary
`echo … | cairn serve` closes stdin instantly and races the MCP handshake. Always use a real
client (a configured agent, or the Inspector).
:::
