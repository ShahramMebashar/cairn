# Friction log (cairn dogfooding)

Feedback for the tool itself, gathered by managing cairn's own development with cairn.
Each observation becomes a tool-improvement task. (The generic working agreement lives in
[WORKFLOW.md](WORKFLOW.md); this file is cairn-repo-specific.)

- **#1 — `claim` doesn't start the task (FIXED for session clients in PROJ-041).** It only
  sets the assignee; status stays `backlog`. The additive `begin` verb now atomically claims,
  enters the working state, and creates an observable attempt; legacy `claim` stays compatible.
- **#2 — closing re-runs already-passing checks.** `run_checks` then `transition→done`
  executes the checks twice. A "skip if passed since last file change" optimization matters
  once suites get slow.
- **#3 — actor attribution.** A running dev/web server can record provenance as
  `human:web` against a task an agent is driving, interleaving actors. Worth clarifying who
  the auto-run-on-close attributes to.
- **#4 — manual checks had no attestation verb (FIXED in PROJ-008).** SPEC said manual
  checks' result is "set by attestation," but nothing set it, so any task with a manual
  check could never reach a closed state (`done`/`canceled` are both gated). Added an
  `attest` verb. Lesson: a task carrying a manual check parks in `in_review` until a human
  attests it, then closes. Note: attesting does NOT auto-close — you still transition to done.
- **#5 — concurrent writers clobber state (FIXED in PROJ-010).** When the web server
  (`human:web`) and an MCP agent (`agent:claude`) act on the same task within a moment, the
  doc was load-mutate-saved wholesale with no optimistic lock, so one overwrote the other —
  observed PROJ-007's `in_progress` reverting to `backlog` after a near-simultaneous
  `ran checks`. Fixed: `Get` records a content hash; `Save` rejects with `ErrConflict`
  (HTTP 409) if the file changed underneath, so the caller re-reads instead of clobbering.
  A tiny TOCTOU window remains (not a hard cross-process lock), but the real seconds-apart
  race is caught.
- **#6 — MCP connection identity can misattribute a different active agent (FIXED in
  PROJ-042).** The available
  Cairn connection was configured as `agent:claude` while Codex was doing the work, so its
  first task writes were stamped as Claude. The actor is fixed at server startup and the
  old tool surface gave the caller no identity warning. Added `identity`; `begin` now
  requires an exact `expected_actor` match before any write.
- **#7 — a session finish can refresh one projection but leave another stale (FIXED in
  PROJ-043).** Dogfooding showed the sidebar move from Active to Awaiting review while the
  open task still displayed the old live heartbeat. Task and session SSE events now
  invalidate both task-detail and session-history queries.
