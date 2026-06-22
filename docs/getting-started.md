# Getting started

## Requirements

- **Go 1.25+** to build.
- **POSIX shell (`sh`)** to run checks. macOS/Linux work out of the box; on Windows use
  WSL or Git Bash.

## Build

```sh
make build        # -> bin/cairn
```

Or directly:

```sh
go build -o bin/cairn ./cmd/cairn
```

## Initialize a repo

cairn stores everything under `.cairn/` at your repo root:

```
.cairn/
  config.yaml     # prefix, states, gates (see docs/task-files.md)
  tasks/          # one Markdown file per task, filename = id
  runs/           # check-run logs (gitignored)
```

You don't create these by hand — `init` scaffolds them. There are three ways, all calling
the same code, so they can't drift:

```sh
cairn init --repo /path/to/myproject        # CLI; prefix derived from the folder name
cairn init --repo /path/to/myproject --prefix MYP   # explicit prefix
```

- **Auto:** `cairn serve` initializes the workspace on first run if it's missing.
- **Web:** open the UI (`make web` + `make web-dev`) in an uninitialized project and click
  **Initialize**.

`init` is idempotent: an existing `config.yaml` is left untouched; it only fills in missing
dirs and the `.gitignore` entry. The prefix defaults to the project folder name, uppercased
(`web-app` → `WEBAPP`).

## Run the server

`cairn` speaks MCP over **stdio**. It is normally launched by an MCP client (see
[connect-claude.md](connect-claude.md)), not run by hand. Identity is fixed at startup
with `--actor` and stamped onto every write:

```sh
cairn serve --actor agent:claude-1 --repo .
```

| Flag | Default | Meaning |
|---|---|---|
| `--actor` | *(required)* | `agent:<name>` or `human:<name>`; recorded in provenance |
| `--repo` | `.` | repo root containing `.cairn/` |

It then waits on stdin for MCP messages. A client closing the connection (or `Ctrl-C`)
exits cleanly.

## Live reload (development)

```sh
make dev          # rebuilds bin/cairn on save via air
```

Note: a rebuild restarts the process, which drops any live MCP client connection —
reconnect the client after a reload.

## Web UI

The browser front-end talks to an HTTP server (the same binary, `cairn web`):

```sh
make web          # Go HTTP server on :8080 (serves /api)
make web-dev      # Vite dev server; proxies /api -> :8080
```

Open the Vite URL. In an uninitialized project the UI shows an **Initialize** screen
(prefix pre-filled from the folder name); after that it's the task board. `cairn web` and
`cairn serve` are two front-ends over one rule-set — agents use MCP, humans use the web.

## Test

```sh
make check        # gofmt + go vet + go test ./...
```

## Quick manual smoke

The best interactive tool is the MCP Inspector — it launches the binary and keeps the
connection open:

```sh
npx @modelcontextprotocol/inspector ./bin/cairn serve --actor agent:dev --repo .
```

Then call `list` (try `ready: true`), `create`, `transition`, etc. from the GUI.

> Don't pipe one-shot JSON into the binary (`echo … | cairn`): closing stdin instantly
> races the MCP handshake. Use a real client.
