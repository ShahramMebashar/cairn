# SPEC.md — Agent-native task tool (v0)

> Repo-native task management where the task graph lives as files in the repo,
> and a single Go binary exposes it equally to humans (UI, later) and AI agents (MCP, now).
> One source of truth (the files), one set of rules (pure functions), two front-ends.
>
> **This document is the frozen v0 contract. Everything here is decided. Do not redesign — implement.**

---

## 0. Principles

- **The files are the source of truth.** No database. Git is sync + audit.
- **One rule-set.** Gate logic lives as pure functions in `internal/task`. MCP and (later) UI are thin adapters that call those functions. They physically cannot diverge.
- **Read-fresh.** No in-memory index, no file watcher in v0. Each operation reads files from disk. At hundreds of tasks on local SSD this is sub-millisecond, and there is no cache to desync against a human editing a file in their editor.
- **Flexible via an open namespace.** Required fields are minimal; unknown frontmatter keys are preserved verbatim and ignored by the engine.
- **Anti-over-engineering.** Single binary, single process, single mutex. No distributed locking, no microservices, no premature abstraction.

---

## 1. File layout

```
.cairn/
  config.yaml
  tasks/
    PROJ-001.md
    PROJ-002.md
  runs/                 # gitignored: check output logs
    PROJ-001-<timestamp>.log
```

- **Filename = task id** (`tasks/PROJ-001.md`). Lookups are a direct file open, never a scan.
- `.cairn/runs/` is gitignored. The scaffold must write a `.gitignore` containing `runs/`.

---

## 2. Task file format

A task is a Markdown file: **YAML frontmatter** (machine fields) + **Markdown body** (human intent).

### Minimal valid task (the floor)

```markdown
---
id: PROJ-001
title: Fix the thing
status: backlog
---
```

Three required fields: `id`, `title`, `status`. Everything else is optional.

### Full-shape task

```markdown
---
id: PROJ-002
title: Add idempotency keys to payment webhook
status: in_progress
assignee: agent:claude-1        # optional; "human:<name>" or "agent:<name>"
deps: [PROJ-001]                # optional; ids that must be CLOSED before this can start
context:                        # optional; opaque to engine, for humans/agents to read
  files: [internal/payments/webhook.go]
  links: [https://...]
checks:                         # optional; gate closing
  - desc: duplicate webhook returns 200 no-op
    cmd: go test ./internal/payments -run TestIdempotent
    timeout: 300                # optional, seconds; falls back to config default
    cwd: .                      # optional; relative to repo root; defaults to repo root
    result: pending             # pending | pass | fail  (engine-managed)
  - desc: reviewed by a human
    type: manual                # a check with no `cmd` is manual; result set by attestation
    result: pending
provenance:                     # optional; append-only audit log, engine-managed
  - { who: human:shah,     at: 2026-06-21T10:00:00Z, did: created }
  - { who: agent:claude-1, at: 2026-06-21T10:05:00Z, did: claimed }
---

Prose intent and constraints go here. The engine never edits the body after creation.
```

### Field rules

| Field | Required | Owner | Notes |
|---|---|---|---|
| `id` | yes | engine | Assigned by engine at create (`prefix` + counter). Never set by a caller. Stable forever; never reused. |
| `title` | yes | caller | Free text. |
| `status` | yes | engine on write | One of the config `states`. |
| `assignee` | no | engine | `human:<name>` or `agent:<name>`. Set by `claim`. |
| `deps` | no | caller | List of task ids. |
| `context` | no | caller | Opaque map. Engine never interprets it. |
| `checks` | no | caller | See §5. |
| `provenance` | no | engine | Append-only. Every write appends one entry. |
| **any other key** | no | preserved | Round-tripped untouched. Engine ignores it. |

**Unknown-key preservation is a hard requirement.** If a user adds `priority: high`, it must still be present, with original ordering and any comments intact, after any engine write. This forbids naive `unmarshal → struct → marshal`. See §6.

---

## 3. config.yaml

```yaml
prefix: PROJ                 # id prefix
counter: 2                   # last assigned number; engine increments on create
states: [backlog, in_progress, in_review, done, canceled]
closed: [done, canceled]     # subset of states considered "closed"
initial: backlog             # the state new tasks start in
check_timeout_default: 120   # seconds, used when a check omits `timeout`
```

- `states` are user-defined free strings. There is **no hardcoded status enum.**
- `closed` lists which states count as closed (for deps-readiness and the checks gate).
- `counter` is the monotonic id source. Engine reads it, assigns `counter+1`, writes it back, all under the process mutex.

---

## 4. Dependencies (`deps`)

- **Meaning:** a task is **ready** iff every id in its `deps` is in a `closed` state. `ready` is a *derived* property, computed on read — never stored. There is no `blocked` field to maintain.
- **Gate point: START.** A task cannot leave the `initial` state until all its deps are closed. (Chosen so an agent scheduling work via `list(ready=true)` skips not-ready tasks. Deps do *not* gate closing.)
- **Dangling dep** (id not present in `tasks/`): **hard error on load.** Loud failure beats a task silently blocked forever by a typo.
- **Cycles** (A→B→A): **rejected at load.** Always. Not configurable.
- **Cross-repo deps:** out of scope for v0. One task graph per repo. A dep id the engine can't resolve in this repo is a dangling-dep error.

---

## 5. Status model & gates

Transitions are **free** — any state to any state — except for exactly **two gates**. No forced ordering (`backlog → in_progress → done` is not enforced; `backlog → done` directly is allowed if both gates pass).

1. **Deps gate** — cannot leave `initial` unless all deps are closed (§4).
2. **Checks gate** — cannot enter a `closed` state unless all `checks` pass.
   - Zero checks ⇒ vacuously passes ⇒ closes freely.
   - On a transition *into* a closed state: if checks are already all `pass`, close. Otherwise **auto-run** the checks first, then close on all-pass or **refuse** on any fail. (This means a close can block up to the checks' timeout. That latency is intentional — closing is exactly when verification belongs.)

Reopening a closed task is allowed (free transition). Check results are **not** auto-reset on reopen; they retain their last value.

---

## 6. Check runner contract

A check with a `cmd` is executed by the engine. A check without a `cmd` is **manual** (result set by attestation, not execution).

- **Execution:** `sh -c "<cmd>"`. `cmd` is therefore a full shell command line — any language/tool, pipes, `&&`, env vars are all valid (`go test ./...`, `npm run test:e2e`, `pytest -q && ruff check .`, `flutter test`, `./scripts/verify.sh`, …). The runner never parses or sequences; one `cmd` = one check. Multi-step ⇒ either multiple checks or the user chains with `&&`.
- **cwd:** repo root by default. Per-check `cwd:` (relative to repo root) overrides — for monorepos.
- **timeout:** per-check `timeout:` seconds; falls back to `check_timeout_default`. On timeout: kill the process, result = `fail`, reason recorded.
- **Result:** process exit code `0` = `pass`, non-zero = `fail`. Universal convention, not configurable.
- **Output:** capture stdout+stderr, store the **tail (~8KB)** to `.cairn/runs/<id>-<timestamp>.log`. The task file stores only `result:` (clean diffs); full output stays in the gitignored log.
- **Platform:** POSIX shell required (`sh`). Document this; native Windows users use WSL/Git Bash. (A future `shell:` config option is out of scope for v0.)

---

## 7. MCP tool surface (the agent API — 7 verbs)

Identity (`--actor`, e.g. `agent:claude-1`) is injected at server startup, **not** passed per call. Every **write** appends a `provenance` entry stamped with the actor, timestamp, and action. Reads never mutate.

| Verb | R/W | Signature | Behavior |
|---|---|---|---|
| `list` | R | `list(status?, assignee?, ready?)` | Returns tasks, filterable. `ready` is the derived deps-satisfied flag. `list(ready=true, status=backlog)` = "what can I start now." |
| `get` | R | `get(id)` | Full task: body, checks (+results), provenance. |
| `create` | W | `create(title, body?, deps?, checks?)` | Engine assigns `id` (`prefix`+counter), sets `status=initial`, first provenance entry `{who:<actor>, did:created}`. |
| `claim` | W | `claim(id)` | Sets `assignee=<actor>`. **Fails if already claimed** by someone else (claiming own is a no-op/ok). |
| `transition` | W | `transition(id, to)` | Applies gates (§5). Closing auto-runs checks and refuses on fail. |
| `run_checks` | W | `run_checks(id, only?)` | Runs all `cmd` checks by default; `only` filters by check index/id. Writes results. |
| `note` | W | `note(id, text)` | Appends a provenance entry carrying `text`. |

Gate logic is **not** implemented inside these verbs — they call `internal/task` pure functions (e.g. `task.Ready`, `task.CanTransition`). The verbs are adapters only.

**MCP library:** use the official Go SDK (`github.com/modelcontextprotocol/go-sdk`) for a public-facing tool. Verify current version/API at scaffold time before pinning.

---

## 8. Store rules

- **Read-fresh:** no index, no watcher (v0). `get` opens one file; `list` scans `tasks/`, parses frontmatter.
- **Lossless writes:** edit the frontmatter as a `yaml.Node` (surgical: change the `status` value node, append to the `provenance` sequence). Never round-trip through a plain struct, which would drop unknown keys and reorder fields. Reads may use a convenience struct with an inline catch-all; writes must use the Node.
- **Body is immutable to the engine** after `create`. No verb edits the Markdown body. Byte-for-byte invariant.
- **Atomic writes:** write to a temp file in the same dir, then `rename` over the target, so a concurrent reader never sees a half-written file.
- **Serialization:** all writes guarded by a single in-process `sync.Mutex`. Single process ⇒ sufficient. No distributed locking.
- **Concurrency with external editors:** a human editing a file in their editor while an agent writes = last-write-wins, same as two people editing one file. Inherent, documented, accepted for v0.

---

## 9. Package layout

```
cmd/cairn/main.go            # flags (--actor, serve), wiring
internal/
  config/                   # load/save config.yaml                        [leaf]
  task/                     # Task type + PURE gate logic: Ready, CanTransition  [leaf]
  check/                    # sh -c runner: cwd, timeout, exit-code, output tail [leaf]
  store/                    # parse, yaml.Node surgical write, atomic save, scan  -> task, config
  mcp/                      # the 7 verbs; adapter over store + task + check       -> store, task, check
  server/                   # http + ws for UI (STUB in v0, no real impl)          -> store
go.mod
```

Dependency graph is acyclic. `config`, `task`, `check` are leaves. `task` owns all rules as pure functions so MCP and the future UI share one implementation.

---

## 10. v0 scope

**In v0:**
- File format, parser (lossless round-trip), config loader.
- `task` pure gate logic + table-driven tests.
- `check` runner.
- `store` (read-fresh, surgical/atomic writes, mutex).
- `mcp` server exposing all 7 verbs.
- Flat hierarchy only — `deps` for ordering; no `parent`/epics/tree.

**Explicitly NOT in v0 (non-goals):**
- Web UI (only a `server` stub). No watcher, no WebSocket — read-fresh has no live-push consumer in a headless MCP/CLI v0; add it purely additively in v1.
- Hierarchy (`parent`, epics, tree). Adding it later is a new optional key ⇒ non-breaking by construction.
- Cross-repo deps, Windows-native shell, multi-user/remote, auth.
- Skills/SKILL.md authoring guidance — later.

**Frozen vs changeable:** the file format/field names, deps semantics, and the two gates are *frozen* — they can't change after the first public tag without breaking users. The UI, internal package layout, and any *new optional* fields are freely changeable.

---

## 11. Build order (leaf-first; one package per Claude Code task)

1. **Scaffold** — `go mod init`, package skeleton, `.gitignore` (`runs/`), example `config.yaml` + 2 example tasks.
2. **`internal/task`** — `Task` type, `Ready(t, all)`, `CanTransition(t, to, all, cfg)`. Table-driven tests covering: deps-gate on leaving initial, checks-gate on closing, free intermediate transitions, reopen keeps results, cycle/dangling detection signatures. **Get this right before anything else — it is the heart.**
3. **`internal/check`** — `sh -c` runner: cwd, timeout (kill), exit-code mapping, output tail to `runs/`.
4. **`internal/config`** — load/save, counter increment.
5. **`internal/store`** — parse (lossless), `yaml.Node` surgical writes, atomic temp+rename, dir scan, process mutex, dangling/cycle validation on load.
6. **`internal/mcp`** — wire the 7 verbs to store/task/check; identity from `--actor`; provenance on every write.
7. **Dogfood** — run the server, point a Claude Code / MCP client at it, drive one real Froshly task end-to-end (create → claim → run_checks → transition done) over MCP.

UI (`server`) comes after v0 is proven.
