---
title: CLI commands
---

# CLI commands

The single `cairn` binary serves one file-based task graph (`.cairn/`) to two front-ends
over the same rule-set: AI agents over MCP (stdio) and humans over HTTP. There are three
subcommands.

| Command | Purpose |
|---|---|
| [`cairn init`](#cairn-init) | scaffold `.cairn/` in a project |
| [`cairn serve`](#cairn-serve) | MCP server over stdio for an agent |
| [`cairn web`](#cairn-web) | HTTP server for the web UI (and MCP over HTTP) |

## cairn init

Scaffolds the `.cairn/` workspace in a project. Idempotent — running it on an
already-initialized repo is a no-op. When `--prefix` is omitted it is derived from the
project folder name.

| Flag | Default | Meaning |
|---|---|---|
| `--prefix` | derived from folder name | id prefix for task ids, e.g. `PROJ` |
| `--repo` | `.` | repo root to initialize |

```sh
cairn init --prefix PROJ --repo .
```

## cairn serve

Runs the MCP server over stdio, the way an agent connects. `--actor` is **required** and
fixes the write identity for the whole process — every write is stamped with it as
provenance (it is never a per-call argument). The workspace is auto-initialized on start,
so a freshly opened project just works.

| Flag | Default | Meaning |
|---|---|---|
| `--actor` | **required** | write identity: `agent:<name>` or `human:<name>` |
| `--client` | "" | agent client identity, e.g. `codex` or `claude` |
| `--repo` | `.` | repo root containing `.cairn/` |

```sh
cairn serve --actor agent:claude-1 --client claude --repo .
```

A client closing the pipe (EOF) or `Ctrl-C` is a normal shutdown, not a failure. See
[Connecting agents](/agents/) for wiring this into an agent, and the
[MCP tools](/reference/mcp-tools) reference for the verbs it exposes.

## cairn web

Runs the HTTP server for the web UI. It prints one machine-parseable line on stdout —
`CAIRN_WEB_URL=<url>` — for a desktop shell to read; the human-readable line goes to
stderr. It prefers `127.0.0.1:7777` (via the default `--addr`); if that port is busy it
scans the next 20 ports and finally falls back to an OS-assigned free port, so the URL it
prints may differ from the requested address.

The same server also exposes MCP over Streamable HTTP at
`/mcp?repo=<path>&actor=<actor>`, so an agent can connect to the running app by URL with
the same rule-set as `cairn serve`.

| Flag | Default | Meaning |
|---|---|---|
| `--addr` | `:8080` | address to listen on (busy port falls back to the next free one) |
| `--actor` | `human:web` | identity stamped on web-driven writes |
| `--repo` | `.` | default repo root (the UI can open other folders) |

```sh
cairn web --addr 127.0.0.1:7777 --repo .
# CAIRN_WEB_URL=http://127.0.0.1:7777
```

See the [HTTP API](/reference/http-api) reference for the endpoints this server exposes.

## Environment variables

| Variable | Default | Meaning |
|---|---|---|
| `CAIRN_SHELL` | `sh` | Shell used to run a task's `cmd` checks (`<shell> -c "<cmd>"`). Must be on `PATH`; set it on Windows when `sh` isn't available. See [Checks & gates → The shell](/guides/checks-and-gates#the-shell). |
