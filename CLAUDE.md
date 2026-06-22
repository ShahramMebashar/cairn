# CLAUDE.md

Project guidance lives in **[AGENTS.md](AGENTS.md)** — read it.

We manage this repo's work **with cairn itself**. Before starting a task, read the working
agreement in **[.cairn/WORKFLOW.md](.cairn/WORKFLOW.md)** — the task lifecycle, the agent
loop, and the note discipline (add concise provenance notes as you work).

Hard rule worth repeating: **all UI uses shadcn/ui. Check the shadcn registry and add
components with `pnpm dlx shadcn@latest add <component>` before hand-rolling anything.**
Style only with the design tokens in `web/src/style.css`. Aesthetic target: Linear —
clean, neutral, one subtle brand accent. Living reference: `web/src/pages/DesignSystem.tsx`.
