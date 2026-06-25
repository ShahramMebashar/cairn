# CLAUDE.md

Project guidance lives in **[AGENTS.md](AGENTS.md)** — read it.

We manage this repo's work **with cairn itself**. Before starting a task, read the working
agreement in **[.cairn/WORKFLOW.md](.cairn/WORKFLOW.md)** — the task lifecycle, the agent
loop, and the note discipline (add concise provenance notes as you work).

Hard rule worth repeating: **all UI uses shadcn/ui. Check the shadcn registry and add
components with `pnpm dlx shadcn@latest add <component>` before hand-rolling anything.**
Style only with the design tokens in `web/src/style.css`. Aesthetic target: Linear —
clean, neutral, one subtle brand accent. Living reference: `web/src/pages/DesignSystem.tsx`.

<!-- cairn:agent-loop:start -->
## Agent loop — required

All work in this repo is tracked in **cairn** (the task graph under `.cairn/`). Drive every
non-trivial change through a task using cairn's MCP tools — don't edit task files by hand:

1. **identity** — confirm your bound actor.
2. **find work** — list ready tasks in the initial state.
3. **begin** — claim a task and open a session (`expected_actor` + a unique `idempotency_key`).
4. **build + heartbeat** — make the change; report concise progress.
5. **note** — add a short provenance note at each meaningful decision.
6. **run_checks** — run the task's checks before handoff.
7. **finish** — end the session into review with a summary.
8. **close** — transition to a closed state once reviewed (re-runs checks).

Full lifecycle, gates, and note discipline: [.cairn/WORKFLOW.md](.cairn/WORKFLOW.md).
<!-- cairn:agent-loop:end -->
