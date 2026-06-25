---
title: The agent loop
---

# The agent loop

All work in cairn is tracked as a task — the graph lives as files under `.cairn/`. Every
non-trivial change is a task, driven through its lifecycle. Drive tasks with cairn's MCP
tools (as an agent) or the web UI (as a human); both go through the same rules.

States are defined in `.cairn/config.yaml` (default `backlog → in_progress → in_review →
done | canceled`; closed: `done`, `canceled`). Transitions are free except two gates:

- **deps gate** — a task can't leave the initial state until every dependency is closed
  (its `ready` flag reflects this).
- **checks gate** — moving into a closed state auto-runs the task's `cmd` checks and
  **refuses** if any fail. Manual checks (no `cmd`) must be attested first.

Both gates are detailed in [Checks & gates](/guides/checks-and-gates).

## The eight steps

1. **Identity** — call `identity`; use its exact bound actor for session writes.
2. **Find work** — list ready tasks in the initial state ("what can I start now").
3. **Begin** — call `begin` with `expected_actor` and a unique `idempotency_key`. This
   claims the task, enters the working state, and returns the session id.
4. **Build + heartbeat** — make the change and periodically report concise progress with
   `heartbeat` (status, never chain-of-thought).
5. **Note** — add a short provenance note at each meaningful decision (see below).
6. **Run checks** — run the task's checks with `run_checks` before handoff. This is
   **enforced**, not advisory: `finish` refuses if any command check is still pending or
   failing (manual checks are exempt — they're attested during review).
7. **Finish** — call `finish` with a useful review summary. This ends the session and moves
   the task to review; it does **not** claim the work is verified or close the task. Closing
   to a done state re-runs the command checks fresh, so a stale `pass` can't slip through.
8. **Close** — transition to a closed state once reviewed; the checks gate runs and refuses
   on failure.

If work is abandoned, call `cancel` with a reason. This ends the session and releases the
assignment while leaving the task open. Legacy clients may still use `claim` + `transition`,
but their work is not session-observable.

::: tip
The mechanics of `begin`/`heartbeat`/`finish`/`cancel` — supervision states, storage, the
trust boundary — are covered in [Sessions](/guides/sessions). Full tool parameters are in
the [MCP tools reference](/reference/mcp-tools).
:::

## Note discipline

Provenance is the task's memory. Add a **concise** note (a sentence or two) whenever you:

- make or change a design decision (and why the alternative was rejected),
- deviate from the task's stated intent,
- discover a constraint (a frozen spec, an API limitation),
- hand off, block, or pause (what's left, what's blocking),
- finish — a closing note naming the files touched, the key decision, and how you verified.

Don't note the obvious ("started coding", restating the title) or anything the diff already
says plainly. If a reader of only the provenance couldn't follow the story, you under-noted;
if every entry restates the obvious, you over-noted.
