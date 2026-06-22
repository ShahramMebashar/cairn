# AGENTS.md — cairn

Repo-native task management. Go binary (`cairn`) serves a file-based task graph to agents
(MCP/stdio) and humans (web). One rule-set in `internal/task`; thin adapters everywhere
else. See `SPEC.md` for the frozen contract and `docs/` for guides.

**We dogfood: this repo's own work is tracked in cairn.** Read
[.cairn/WORKFLOW.md](.cairn/WORKFLOW.md) before starting a task — it defines the lifecycle,
the agent loop (claim → in_progress → build → note → run_checks → done), the note
discipline, and the **task body style** (concise, structured Markdown — short `##` sections,
inline code for identifiers — that reads well for humans and agents). Add concise provenance
notes (`note`) as you make decisions; log tool friction in that file's friction log.

## Backend (Go)

- Gate logic lives ONLY in `internal/task` (pure). MCP verbs and the web server call it —
  never reimplement rules in an adapter.
- Run `make check` (gofmt + vet + test) before claiming work done.
- Workspace dir is `.cairn/` (config, tasks, runs).

## UI / Design system — RULES (must follow)

The web UI must be world-class: clean, professional, consistent, and easy to follow —
**Linear-like**. To get there without churn:

1. **Use shadcn/ui. Do not reinvent the wheel.** Before building ANY UI element, check if
   shadcn provides it. If it does, add it — don't hand-roll.
   - Browse: https://ui.shadcn.com/docs/components
   - Add: `cd web && pnpm dlx shadcn@latest add <component>`
   - Components live in `web/src/components/ui/`. Compose them; don't duplicate them.
2. **Hand-roll only as a last resort** — when shadcn has no equivalent. When you do, match
   shadcn's API shape, use `cn()` + the design tokens, and put it under `components/`.
3. **Use design tokens, never raw colors.** Style with the semantic Tailwind classes backed
   by CSS variables in `web/src/style.css` (`bg-background`, `text-muted-foreground`,
   `border`, `bg-primary`, `text-brand`, `ring-ring`, `rounded-lg`, …). Never hardcode hex
   or arbitrary `oklch(...)` in components.
4. **Aesthetic bar (Linear):** monochrome neutral base, one restrained brand accent (indigo)
   used sparingly for focus/selection/links; tight spacing; subtle 1px borders over heavy
   shadows; fast, quiet transitions; keyboard-friendly. Prefer clarity over decoration.
5. **Consistency:** reuse spacing/radius/type scale already defined. New patterns go through
   the design system, not one-off styles. View the living reference at `/#design`
   (`web/src/pages/DesignSystem.tsx`).

If a change needs a component shadcn offers, adding it via the CLI is REQUIRED, not optional.
