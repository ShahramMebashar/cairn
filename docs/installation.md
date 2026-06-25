---
title: Installation
---

# Installation

Just want to use cairn? **[Download the desktop app](#desktop-app-recommended)** — no toolchain
needed. The rest of this page covers building from source (for the CLI, a headless server, or
developing cairn itself).

## Desktop app (recommended)

The easiest way to run cairn is the desktop app — a small native window with a live menu-bar
tray that bundles the web UI and the `cairn` engine together. It's the same thing `cairn web`
serves, just native, and it configures agents, fires notifications, and stays in your tray.

**[Download the latest release →](https://github.com/ShahramMebashar/cairn/releases/latest)**
and grab the installer for your OS:

| OS | Download | Notes |
| --- | --- | --- |
| **macOS** | `Cairn_<ver>_aarch64.dmg` (Apple Silicon) or `_x64.dmg` (Intel) | Signed & notarized — opens normally. |
| **Windows** | `Cairn_<ver>_x64-setup.exe` (NSIS) | Not code-signed yet — SmartScreen shows "unknown publisher"; click **More info → Run anyway**. |
| **Linux** | `Cairn_<ver>_amd64.AppImage` or `_amd64.deb` | AppImage: `chmod +x` then run. `.deb`: `sudo dpkg -i`. |

The app **auto-updates** in place — once installed, new releases are picked up automatically.

::: tip
Everything below (building from source, the CLI, headless MCP) is optional — only needed if
you're developing cairn itself or running it on a server. To just use cairn, download above.
:::

## Build from source

Prefer the command line or a headless server, or developing cairn itself? Build the binary —
the desktop app bundles this same binary as a sidecar.

**You'll need:** **Go 1.25+** (the binary); **Node 18+** (the embedded web UI); and a **Rust
toolchain** (only for `make desktop`). Task checks run via a POSIX shell — see
[Checks & gates → The shell](/guides/checks-and-gates#the-shell) for the Windows note.

```sh
make build        # -> bin/cairn
```

Or directly:

```sh
go build -o bin/cairn ./cmd/cairn
```

That single binary is everything: the MCP server (`cairn serve`), the web server
(`cairn web`), and the workspace scaffolder (`cairn init`). To build the desktop app yourself:

```sh
make desktop       # -> native installer (.dmg / .exe / .AppImage / .deb)
make desktop-dev   # native window against a dev server (run `make web` alongside)
```

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
