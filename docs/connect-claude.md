# Connect to Claude

`cairn` is an MCP **stdio** server. The MCP client launches the binary and talks to it
over stdin/stdout — so you register the *command*, you don't run a long-lived service
yourself.

Build it first:

```sh
make build        # -> bin/cairn
```

Use **absolute paths** in all configs below (`$(pwd)` expands to your repo root).

---

## Claude Code (CLI)

The quickest path — `claude mcp add`:

```sh
claude mcp add cairn -- "$(pwd)/bin/cairn" serve --actor agent:claude-1 --repo "$(pwd)"
```

Verify and inspect:

```sh
claude mcp list
claude mcp get cairn
```

Then in a session, the 7 verbs appear as tools. Try:

> List the ready cairn tasks.
> Create a task titled "wire up logging" that depends on PROJ-001.
> Transition PROJ-003 to done.

Remove it with `claude mcp remove cairn`.

### Project-scoped (`.mcp.json`)

To share the server with everyone working in the repo, commit a `.mcp.json` at the repo
root instead:

```json
{
  "mcpServers": {
    "cairn": {
      "command": "bin/cairn",
      "args": ["serve", "--actor", "agent:claude-1", "--repo", "."]
    }
  }
}
```

Claude Code resolves `command`/`--repo` relative to the project root, so this stays
portable across clones. (Per-user setups still prefer absolute paths.)

---

## Claude Desktop

Edit the MCP config file:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "cairn": {
      "command": "/ABSOLUTE/PATH/TO/cairn/bin/cairn",
      "args": ["serve", "--actor", "agent:claude-1", "--repo", "/ABSOLUTE/PATH/TO/cairn"]
    }
  }
}
```

Restart Claude Desktop. The cairn tools appear under the 🔌 (tools) menu.

---

## Identity (`--actor`)

`--actor` is the writer recorded in every task's append-only `provenance` log. Use a
distinct actor per agent/human so the audit trail is meaningful:

- `agent:claude-1`, `agent:claude-2` — different agent instances
- `human:shah` — a person driving the tools directly

Reads never write provenance; every write (create, claim, transition, run_checks, note)
appends one entry stamped with this actor and a timestamp.

---

## Troubleshooting

- **Tools don't appear / server fails to start:** the `command` path must be absolute (or
  repo-relative for project `.mcp.json`) and the binary must exist — run `make build`.
- **"--actor is required":** add `--actor agent:<name>` to `args`.
- **Workspace not set up:** `cairn serve` auto-initializes `.cairn/` on first run, so
  this is rarely an issue. To set it up explicitly, run `cairn init --repo <dir>` (see
  [getting-started.md](getting-started.md)).
- **Checks never pass on close:** a `manual` check (no `cmd`) stays `pending` until
  attested; closing is refused until then. See [task-files.md](task-files.md).
