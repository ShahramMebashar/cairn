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

## Desktop app (Tauri)

cairn ships as a light cross-platform desktop app. The Go binary embeds the built UI
(`web/embed.go`, `//go:embed all:dist`) and runs as a **Tauri sidecar**; the Rust shell
in `desktop/src-tauri/` (a top-level sibling of `web/`, with its own `desktop/package.json`
holding just the Tauri CLI) spawns it, reads the `CAIRN_WEB_URL=` line it prints, waits for
`/healthz`, then points the webview at it. The server also exposes MCP over Streamable
HTTP at `/mcp?repo=<path>&actor=<actor>` so agents can connect to the running app.

- **Dev:** `make desktop-up` runs everything together — Go server + Vite + the native window
  (or split across shells with `make web` + `make desktop-dev`).
- **Build (this OS):** `make desktop` → installer under `desktop/src-tauri/target/release/bundle/`.
- **Release:** push a `v*` tag → `.github/workflows/release.yml` builds mac/linux/windows
  installers (unsigned MVP) and attaches them to the GitHub Release. PRs get a no-bundle
  desktop build check via `.github/workflows/ci.yml`.
- The sidecar dies with the app: `cairn web --parent-watch` shuts down on stdin EOF.
- Port: prefers `127.0.0.1:7777`, falls back to the next free port (the real URL is what
  the shell reads from stdout and what the "Connect an agent" panel shows).

### Desktop-native features

The app is an always-on hub: closing the window **hides to the tray** (server + MCP stay up;
Quit from the tray/menu), `single-instance` focuses the running window instead of spawning a
second sidecar, and there's an opt-in **Launch at login**. Other native bits, all wired in
`desktop/src-tauri/src/lib.rs` + guarded on the JS side via `web/src/lib/desktop.ts`
(`isTauri()` — browser/dev builds no-op):

- **OS notifications** — `notifications.ts` mirrors ready / check-failed / awaiting-review to
  native notifications when the window is unfocused.
- **Quick-capture** — global `Cmd/Ctrl+Shift+K` opens a small capture window (`#capture` route,
  `CaptureView`) targeting the last project with a switcher.
- **Native menu + window-state** — menu bar with accelerators (events flow to the UI via
  `useDesktopMenu`); window size/position persists.
- **Auto-update** — checks the GitHub Releases `latest.json` on launch and from Settings;
  installs + relaunches. The release workflow uses `tauri-apps/tauri-action` and **requires repo
  secrets `TAURI_SIGNING_PRIVATE_KEY` + `TAURI_SIGNING_PRIVATE_KEY_PASSWORD`** (the Tauri *update*
  key — separate from OS code signing; pubkey lives in `tauri.conf.json`). Without them, tagged
  releases fail to sign the updater artifacts.

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
