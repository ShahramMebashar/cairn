---
title: Installation
---

# Installation

Just want to use cairn? **[Download the desktop app](#desktop-app-recommended)**. No toolchain
needed. The rest of this page is for building from source: the CLI, a headless server, or
working on cairn itself.

## Desktop app (recommended)

The desktop app is a small native window with a tray. It bundles the web UI and the cairn
engine, so it is the same thing `cairn web` serves, just native.

**[Download the latest release](https://github.com/ShahramMebashar/cairn/releases/latest)** and
pick the installer for your OS:

| OS | Download | Notes |
| --- | --- | --- |
| **macOS** | `Cairn_<ver>_aarch64.dmg` (Apple Silicon) or `_x64.dmg` (Intel) | Signed and notarized. Opens normally. |
| **Windows** | `Cairn_<ver>_x64-setup.exe` (NSIS) | Not signed yet. SmartScreen says "unknown publisher": click **More info**, then **Run anyway**. |
| **Linux** | `Cairn_<ver>_amd64.AppImage` or `_amd64.deb` | AppImage: `chmod +x`, then run. `.deb`: `sudo dpkg -i`. |

Once installed, the app updates itself when a new release ships.

::: tip
Everything below is optional. You only need it to develop cairn or run it on a server. To use
cairn, download it above.
:::

## Build from source

Building the binary gives you the CLI and a headless server. The desktop app bundles this same
binary as a sidecar.

You need **Go 1.25+** for the binary and **Node 18+** for the embedded web UI. `make desktop`
also needs a **Rust toolchain**. Checks run in a POSIX shell; see
[Checks and gates](/guides/checks-and-gates#the-shell) for the Windows note.

```sh
make build        # -> bin/cairn
```

Or:

```sh
go build -o bin/cairn ./cmd/cairn
```

That one binary is everything: the MCP server (`cairn serve`), the web server (`cairn web`),
and the workspace setup (`cairn init`). To build the desktop app yourself:

```sh
make desktop       # native installer (.dmg / .exe / .AppImage / .deb)
make desktop-dev   # native window against a dev server (run `make web` alongside)
```

## Set up a repo

cairn keeps everything under `.cairn/` at your repo root:

```
.cairn/
  config.yaml     # prefix, states, gates. See Task files and config.
  tasks/          # one Markdown file per task, filename is the id
  runs/           # check-run logs (gitignored)
```

You don't create these by hand. `init` sets them up, and it's safe to re-run: it leaves an
existing `config.yaml` alone and only fills in what's missing. There are three ways in, all
calling the same code:

```sh
cairn init --repo /path/to/project              # prefix from the folder name
cairn init --repo /path/to/project --prefix MYP # explicit prefix
```

- **Automatic:** `cairn serve` sets up the workspace on first run if it's missing.
- **Web:** open `cairn web` in a fresh project and click **Initialize**. The prefix is filled in
  from the folder name, uppercased (`web-app` becomes `WEBAPP`).

## Run the web UI

The browser app and the desktop app both run `cairn web`:

```sh
cairn web --repo .
```

It prints the URL it bound to (`CAIRN_WEB_URL=http://127.0.0.1:7777`, or the next free port) and
serves the board, the graph, and the **Connect** page. For local development with live reload:

```sh
make web          # Go HTTP server on :8080 (serves /api)
make web-dev      # Vite dev server, proxies /api to :8080
```

## Connect an agent

`cairn serve` is the MCP server. An MCP client launches it; you don't run it by hand. The easy
path is the [Connect page](/agents/) in the web UI: one click writes the agent's config. To do
it by hand, see the per-agent pages like [Claude Code](/agents/claude):

```sh
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude-1 --repo "$(pwd)"
```

## Test

```sh
make check        # gofmt + go vet + go test ./...
```

## Try it by hand

The MCP Inspector launches the binary and keeps the connection open:

```sh
npx @modelcontextprotocol/inspector ./bin/cairn serve --actor agent:dev --repo .
```

Then call `identity`, `list` (try `ready: true`), `begin`, `heartbeat`, and `finish` from the
GUI.

::: warning Don't pipe one-shot JSON into the binary
`echo … | cairn serve` closes stdin right away and races the MCP handshake. Use a real client:
a configured agent or the Inspector.
:::
