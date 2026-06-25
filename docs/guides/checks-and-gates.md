---
title: Checks & gates
---

# Checks & gates

Transitions in cairn are free — any state to any state — except for two gates. Everything
else about moving a task through its states is unconstrained.

## The two gates

1. **Deps gate** — a task can't leave the `initial` state until every id in its `deps` is in
   a `closed` state. Deps do not gate closing, only starting. The `ready` flag reflects this
   and is derived on read, never stored.
2. **Checks gate** — a task can't enter a `closed` state unless all its `checks` pass.
   - Zero checks ⇒ passes vacuously.
   - On closing, if checks aren't already all `pass`, the engine **auto-runs** the `cmd`
     checks, then closes on all-pass or **refuses** on any fail.

::: warning
Reopening a closed task is allowed, but check results are **not** reset on reopen — they keep
their last value, so a re-close reuses them. Closing re-runs `cmd` checks fresh, so a stale
`pass` from an earlier attempt still can't slip a broken close through.
:::

See [Task files & config → Gates](/guides/task-files#gates) for where checks live in the task
file.

## `cmd` vs `manual` checks

| Kind | Has `cmd`? | How `result` is set |
|---|---|---|
| `cmd` | yes | executed; exit code decides pass/fail |
| `manual` | no | set by attestation, not execution |

- A check **with** a `cmd` is executed via `sh -c "<cmd>"` — any shell line works
  (`go test ./...`, `pytest -q && ruff check .`, `./scripts/verify.sh`).
- A check **without** a `cmd` is **manual** — its `result` is set by attestation, not
  execution. A pending manual check blocks closing until it's resolved.

## Exit codes, timeout, and output

- Exit code `0` = `pass`, non-zero = `fail`.
- A check may set an optional `timeout` (seconds); otherwise `check_timeout_default` from
  `config.yaml` applies. On timeout the process (and its process group) is killed and the
  result is `fail`.
- An optional `cwd` (relative to the repo root) sets the working directory for the command.
- Output (stdout+stderr, ~8KB tail) goes to `.cairn/runs/<id>-<timestamp>.log`. The task file
  stores only `result:` so diffs stay clean. The `runs/` directory is gitignored.
