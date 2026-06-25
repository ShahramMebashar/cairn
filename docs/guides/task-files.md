---
title: Task files & config
---

# Task files & config

## File layout

```
.cairn/
  config.yaml           # engine config
  tasks/
    PROJ-001.md         # filename = task id; lookups are a direct file open
  runs/                 # gitignored check-run logs
```

## Task file format

A task is Markdown: **YAML frontmatter** (machine fields) + **Markdown body** (human
intent). The engine never edits the body after creation.

### Minimal valid task

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
assignee: agent:claude-1
deps: [PROJ-001]
context:                       # opaque to the engine; for humans/agents
  files: [internal/payments/webhook.go]
checks:
  - desc: duplicate webhook returns 200 no-op
    cmd: go test ./internal/payments -run TestIdempotent
    timeout: 300               # optional seconds; else check_timeout_default
    cwd: .                     # optional; relative to repo root
    result: pending            # pending | pass | fail  (engine-managed)
  - desc: reviewed by a human
    type: manual               # no cmd ⇒ manual; result set by attestation
    result: pending
provenance:                    # append-only audit log (engine-managed)
  - {who: human:shah, at: 2026-06-21T10:00:00Z, did: created}
---

Prose intent and constraints go here.
```

### Field ownership

| Field | Owner | Notes |
|---|---|---|
| `id` | engine | assigned at create: `prefix` + time-ordered base32 token; sorts by creation time; never reused |
| `title` | caller | free text |
| `status` | engine | one of `config.states` |
| `assignee` | engine | set by `claim` (`human:<name>` / `agent:<name>`) |
| `deps` | caller | task ids that must be closed first |
| `context` | caller | opaque map; engine never interprets it |
| `checks` | caller | see [gates](#gates) |
| `provenance` | engine | append-only; one entry per write |
| **any other key** | — | preserved verbatim, ignored by the engine |

**Unknown-key preservation is guaranteed.** Add `priority: high` and it survives — with
ordering and comments intact — across engine writes. (Writes edit the YAML node surgically,
never a struct round-trip.)

## config.yaml

```yaml
prefix: PROJ                 # id prefix
counter: 2                   # deprecated, unused; retained so existing configs still parse
states: [backlog, in_progress, in_review, done, canceled]
closed: [done, canceled]     # subset of states considered "closed"
initial: backlog             # state new tasks start in
check_timeout_default: 120   # seconds, when a check omits timeout
```

- `states` are free strings you define — there is no hardcoded status enum.
- `closed` drives deps-readiness and the checks gate.
- Ids are minted at create as `prefix` + a time-ordered, collision-resistant base32 token, so
  concurrent creators in separate clones never collide — no shared counter, no merge conflict.
  `counter` is retained only for backward-compatible parsing and is no longer incremented.

## Dependencies

- A task is **ready** when every id in `deps` is in a `closed` state. `ready` is derived on
  read, never stored.
- **Gate point is START:** a task can't leave `initial` until its deps are closed. Deps do
  not gate closing.
- A dep id not present in `tasks/` (**dangling**) or a **cycle** is a hard error on load —
  loud failure beats a silently-stuck task.

## Gates

Transitions are free — any state to any state — except two gates:

1. **Deps gate** — can't leave `initial` unless all deps are closed.
2. **Checks gate** — can't enter a `closed` state unless all `checks` pass.
   - Zero checks ⇒ passes vacuously.
   - On closing, if checks aren't already all `pass`, the engine **auto-runs** the `cmd`
     checks, then closes on all-pass or **refuses** on any fail.

Reopening a closed task is allowed; check results are **not** reset on reopen — they keep
their last value, so a re-close reuses them.

::: tip
For the full gates + checks model — including manual checks, exit codes, and run logs — see
[Checks & gates](/guides/checks-and-gates).
:::

## Checks

- A check **with** a `cmd` is executed via `sh -c "<cmd>"` — any shell line works
  (`go test ./...`, `pytest -q && ruff check .`, `./scripts/verify.sh`).
- A check **without** a `cmd` is **manual** — its `result` is set by attestation, not
  execution. A pending manual check blocks closing until resolved.
- Exit code `0` = `pass`, non-zero = `fail`. Timeout ⇒ the process (and its group) is
  killed, `result: fail`.
- Output (stdout+stderr, ~8KB tail) goes to `.cairn/runs/<id>-<timestamp>.log`; the task
  file stores only `result:` for clean diffs.

## Provenance

Every write appends one entry `{who, at, did[, text]}`, stamped with the server's
`--actor`. It is the task's append-only audit trail. `note` adds an entry with `text` and no
state change.
