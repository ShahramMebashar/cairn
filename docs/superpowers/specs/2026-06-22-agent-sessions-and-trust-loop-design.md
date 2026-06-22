# Agent sessions + evidence-gated completion

Date: 2026-06-22
Status: proposed for implementation

## Goal

Make Cairn the agent-neutral control plane for software work:

1. **Sessions** make each bounded agent attempt observable and recoverable.
2. **Evidence** connects an attempt to concrete repository artifacts.
3. **Fresh verification** proves checks apply to the current workspace snapshot.
4. **Completion gates** prevent `done` from meaning only "the agent said so."

The core does **not** launch, host, or steer model runtimes. Claude, Codex, Cursor, and
other workers execute independently and report through the same MCP contract.

## Scope and product bet

This is a deliberate product pivot, not a small task-tracker feature. The complete design
adds a second durable entity, live supervision state, snapshot verification, new protocol
verbs, and new UI. That cost is justified only if Cairn is becoming an agent control plane.

The minimum-friction north star still applies:

- ship the work as vertical slices with visible user value after the first slice;
- keep the MVP to one session store plus one ephemeral live-state directory;
- do not persist raw session logs in the MVP;
- do not build a custom lease database, stale-lock algorithm, or generic repair framework;
- stop after the observable-session slice and validate usage before enforcing trust gates.

The decision checkpoint is empirical: if agents do not reliably call `begin`/`finish`, or
humans do not use active/stalled/review visibility, do not continue into the completion-gate
slice merely because the architecture can support it.

## Product invariants

- A task may have many historical sessions but at most one live session.
- A session cannot begin unless the caller's expected identity matches the MCP connection's
  bound actor exactly.
- Every session belongs to exactly one attempt; handoff is the only way to continue an
  existing attempt in another session.
- Session health is derived from its lifecycle state and heartbeat age, never manually
  copied onto the task.
- Agent-reported metadata is useful context but is not verified evidence.
- Only Cairn may mark evidence or checks as verified.
- A passing check is effective only for the workspace snapshot it actually tested.
- Finishing a session requests review; it does not close the task.
- Entering a configured completion state requires fresh checks and sufficient evidence.
- `canceled` may remain closed for dependency purposes without pretending work was verified.
- Gate logic remains pure in `internal/task`; MCP and HTTP remain thin adapters.
- Existing v0 task files continue to load. New fields are additive.

## Concepts

### Task

The durable intent and workflow object already stored in `.cairn/tasks/`.

### Session

A bounded attempt by one actor to advance one task. A session owns lifecycle, execution
metadata, final summary, and references to evidence. It does not own workflow status.

Session states:

- `active` — worker is expected to make progress and heartbeat.
- `blocked` — worker explicitly needs input or an external condition.
- `handed_off` — terminal; context was prepared for another worker.
- `finished` — terminal; the attempt ended and may be reviewable.
- `canceled` — terminal; the attempt was intentionally abandoned.

`active` and `blocked` are live states. The other states release live ownership.

### Evidence

A typed claim attached to a session. Evidence carries both the agent's declaration and
Cairn's verification result.

Initial evidence types:

- `commit` — a local commit object exists; Cairn records its SHA and tree.
- `workspace` — Cairn records the current workspace snapshot digest.
- `pull_request` — URL/number is recorded; unverified until an adapter confirms it.
- `artifact` — a repo-relative file exists and its SHA-256 is recorded.
- `check_run` — generated only by Cairn from a check execution.
- `no_change` — explicit rationale; requires a manual check/attestation to count.

A narrative summary is mandatory for `finish`, but is not evidence by itself.

### Workspace snapshot

The trust boundary is the exact content tested, not merely `HEAD`. Snapshot version `ws1`
is a canonical manifest produced without Git diff output:

```text
sha256(
  format-version + NUL +
  for each sorted tracked-or-untracked, non-ignored path:
    path-length + path-bytes + file-kind + executable-bit + content-length + sha256(contents)
)
```

Paths come from `git ls-files --cached --others --exclude-standard -z`; Cairn reads and
hashes workspace bytes itself. Symlinks hash their target, and submodules hash the checked-out
commit. Ignored files and Cairn's own `.cairn/` control-plane files do not participate; task,
session, provenance, and check-result writes must not invalidate the source snapshot they
describe. For a clean repository the response also includes
the Git tree OID, but the digest remains the universal comparison key. Snapshot computation
is read-only and must not modify the worktree, index, or refs.

Unlike `git diff --binary`, the manifest format does not vary with Git's presentation
format. Two machines with the same path bytes, modes, and file bytes produce the same `ws1`
digest. Line-ending conversion or mode differences intentionally produce a different digest
because the check ran against different workspace bytes. The format version is always stored
with the digest; different versions are never compared as equal.

If an untracked file exceeds the configured hashing limit, snapshot creation fails closed
with a useful error. Silently omitting content would produce false freshness.

## Storage

```text
.cairn/
  tasks/                       # existing, git-tracked
  sessions/                    # durable, git-tracked
    <session-id>.yaml
  runs/                        # existing check logs, gitignored
  live/                        # ephemeral supervision state, gitignored
    <session-id>.json          # heartbeat, progress, local worktree, live usage
  write.lock                   # persistent file used for an OS advisory lock, gitignored
```

`init` adds `live/` and `write.lock` to `.cairn/.gitignore`. Existing `runs/` keeps its
meaning as check-run output; session execution is never overloaded onto that directory.
The MVP keeps only the latest progress string in `live/`; an append-only session-log store is
deferred until real usage proves it necessary.

### Durable session file

```yaml
id: ses_7f42c91e8a32
task: PROJ-037
attempt: att_82a7de615cf1
previous_session: ses_1a2b3c4d5e6f
actor: agent:codex-1
client: codex
model: gpt-5.1-codex
status: finished
started_at: 2026-06-22T18:00:00Z
ended_at: 2026-06-22T18:42:11Z
branch: codex/session-trust
head_started: 98c1d6a
head_finished: c4f0a12
summary: Added snapshot-bound checks and completion evidence.
usage:
  input_tokens: 42000
  output_tokens: 7800
  cached_tokens: 19000
  tool_calls: 67
evidence:
  - id: ev_01
    type: commit
    value: c4f0a12
    verified: true
    verified_at: 2026-06-22T18:42:10Z
    snapshot: ws1:8de4...
  - id: ev_02
    type: check_run
    value: make check
    verified: true
    verified_at: 2026-06-22T18:41:55Z
    snapshot: ws1:8de4...
handoff:
  requested_actor: agent:claude
  next_steps: Review the migration behavior.
```

The file is machine-owned YAML, atomically replaced, and forward-compatible: unknown keys
are preserved. Absolute worktree paths and heartbeat timestamps stay in `live/` so durable
files remain portable and do not create heartbeat churn in Git.

### Session IDs

Use `ses_` plus 96 bits from `crypto/rand`, encoded as lowercase hex. IDs are safe across
processes, branches, and cloned worktrees without a shared counter.

Attempt IDs use the same scheme with `att_`.

### Attempt lineage

An **attempt** is the chain of sessions pursuing one candidate result for a task. It prevents
old sessions and evidence from accidentally satisfying completion after a retry or reopen.

- A normal `begin` creates a new `attempt` and sets `previous_session` empty.
- `handoff` terminally closes its session and issues a random, single-use continuation token
  stored only in the live handoff response and as a hash in the durable session.
- The successor calls `begin(resume_from, continuation_token, ...)`. Cairn verifies the token,
  copies the predecessor's `attempt`, sets `previous_session`, and marks the predecessor
  `continued_by` the new session.
- A handed-off session can have at most one successor. A session chain is therefore linear,
  matching the one-live-session invariant.
- Beginning without a valid handoff always creates a new attempt, even when the previous
  session finished, stalled, or was canceled.
- Reopening a completed task does not reuse its old attempt; the next normal `begin` creates
  a new one.

The task stores the additive machine-owned `active_attempt` ID when a session begins. It is
not cleared on terminal session state, so review and completion know which attempt is current.
Starting a new attempt replaces it. Completion facts consider only sessions and evidence
whose `attempt` equals the task's `active_attempt`.

Evidence eligibility is revalidated at read/transition time:

- `workspace` and `check_run` count only when their `wsN:` digest equals the current snapshot;
- `commit` counts only when the commit still exists and is reachable from current `HEAD`;
- `artifact` counts only while its repo-relative path still has the recorded content hash;
- remote evidence counts only when its provider verification is still valid;
- narrative summaries, canceled attempts, and prior attempts never count.

## Live state and derived supervision

The heartbeat interval defaults to 30 seconds; `session_stale_after` defaults to 3 minutes.
Agents send cumulative usage totals, not deltas, so retries are idempotent.

Task execution state is derived with this precedence:

1. latest live session is `blocked` → `blocked`
2. latest live session is `active` and heartbeat expired → `stalled`
3. latest live session is `active` and heartbeat fresh → `active`
4. latest session is `finished`, task is open, and review prerequisites pass → `awaiting_review`
5. latest session is `finished` but review prerequisites fail → `verification_needed`
6. otherwise → no execution state

The task's configured workflow status and derived execution state are separate dimensions.
For example, a task may be `in_progress / blocked` or `in_review / awaiting_review`.

The web list computes the repository snapshot once per request, not once per task.

## Lifecycle operations

All writes stamp actor and time. Mutations are idempotent where a client retry is likely.

### `begin`

```json
{
  "id": "PROJ-037",
  "expected_actor": "agent:codex",
  "client": "codex",
  "model": "gpt-5.1-codex",
  "worktree": "/path/to/worktree",
  "resume_from": "",
  "continuation_token": "",
  "force": false
}
```

Behavior:

1. Compare `expected_actor` with the server-bound actor and reject before any write when they
   differ. This prevents a Codex worker from accidentally writing through a connection
   configured as `agent:claude`.
2. Validate task exists, is not closed, and can leave the initial state.
3. While holding the repository write lock, scan durable sessions for this task and refuse if
   another `active` or `blocked` session exists.
4. Refuse a stalled active session unless `force=true`; forced takeover cancels the old
   session with
   provenance explaining who took it over.
5. Resolve attempt lineage: validate and consume a handoff continuation, or mint a new
   attempt ID.
6. In one task-file write, claim for the calling actor and, when initial, move to the
   configured first working state; reject a conflicting assignee.
7. Create the durable session and its ephemeral live-state file.

This intentionally combines today's `claim` + transition dance for session-aware clients.
The existing verbs remain available for compatibility.

The MCP server also exposes a read-only `identity` tool returning its bound `actor`, declared
`client`, and version. Session-aware setup configures both `--actor` and `--client`; tool
descriptions state the bound actor so identity is visible before the first mutation. The
explicit `expected_actor` comparison remains the hard guard because labels and descriptions
are advisory to an LLM.

### `heartbeat`

```json
{
  "session": "ses_7f42c91e8a32",
  "progress": "Implementing snapshot comparison tests",
  "usage": {
    "input_tokens": 42000,
    "output_tokens": 7800,
    "cached_tokens": 19000,
    "tool_calls": 67
  }
}
```

Atomically replaces the session's ephemeral live-state file. Only the latest progress string
is retained in the MVP. Usage fields are cumulative and monotonic; a lower retry value is
ignored. Raw chain-of-thought, prompts,
secrets, and arbitrary tool payloads are explicitly outside the protocol.

### `block`

```json
{
  "session": "ses_7f42c91e8a32",
  "reason": "Need the completion-state migration decision",
  "needs": "Choose legacy compatibility or automatic migration"
}
```

Moves `active → blocked`, persists the reason durably, and emits a high-signal event for the
human inbox. Live ownership remains with the session. Repeating the same request is a no-op.

### `resume`

Moves `blocked → active`, clears the current live blocker, records who resumed it, and
requires the same actor unless `force=true`.

### `record_evidence`

```json
{
  "session": "ses_7f42c91e8a32",
  "type": "commit",
  "value": "c4f0a12"
}
```

Cairn validates the declaration against the repository and stores its normalized value,
verification result, timestamp, and current snapshot. Agents cannot set `verified`.

### `handoff`

```json
{
  "session": "ses_7f42c91e8a32",
  "summary": "Snapshot hashing is implemented; gate wiring remains.",
  "next_steps": "Wire CompletionFacts into CanTransition.",
  "requested_actor": "agent:claude"
}
```

Moves the session to `handed_off`, removes its live-state file, releases the task assignee,
and records an advisory target. It returns the single-use continuation token described in
**Attempt lineage**. The next actor must call `begin`; handoff never impersonates or
force-claims on another actor's behalf.

### `finish`

```json
{
  "session": "ses_7f42c91e8a32",
  "summary": "Implemented snapshot-bound checks with tests.",
  "usage": { "input_tokens": 42000, "output_tokens": 7800 }
}
```

Moves the session to `finished`, copies final cumulative usage into the durable file, records
the ending HEAD/snapshot, and removes the live-state file. It does **not** move the task to a
completion state.

Before completion trust is enabled (the observable-session slice), a non-empty summary is
enough for Cairn to move the task to the configured review state; existing v0 completion
gates remain unchanged. Once completion trust is enabled, automatic checks must be fresh and
verified evidence must be present before that review transition. Pending manual checks are
allowed because human review is exactly what should happen next. Otherwise the task remains
open with derived state `verification_needed`, and a new session may address it.

### `cancel`

Moves a live session to `canceled`, requires a reason, and removes its live-state file. It does not
cancel the task. Task cancellation remains an explicit workflow transition.

### Reads

- `get_session(session)` — durable fields plus current live state and derived health.
- `list_sessions(task?, actor?, status?, health?)` — newest first.
- Existing `get(task)` adds recent sessions, completion facts, and execution state.
- Existing `list(...)` accepts optional `execution` and returns execution state/counts.

## Check freshness

Each check gains additive machine-owned fields:

```yaml
checks:
  - desc: repository checks pass
    cmd: make check
    workspace_mutation: allow
    result: pass
    snapshot: ws1:8de4...
    checked_at: 2026-06-22T18:41:55Z
```

`RunChecks` behavior:

1. Compute the workspace snapshot before running a command.
2. Run the command and capture its existing log.
3. Compute the snapshot again.
4. A zero exit is `pass` and binds to the **after** snapshot. The configured check command is
   responsible for validating any changes it makes before returning zero.
5. When `workspace_mutation: forbid`, a changed before/after snapshot instead fails with a
   precise mutation error. The default is `allow` to support formatters, generators, coverage
   manifests, and build steps that intentionally write non-ignored files.
6. Store the after `snapshot`, `checked_at`, and whether the command changed the workspace;
   attach Cairn-generated `check_run` evidence to the
   current session when one exists.

Projects should still gitignore disposable coverage/build output. Tracked generated code is
allowed when the check command also validates its generated result (for example,
`go generate ./... && git diff --check && go test ./...`). Cairn cannot infer command ordering;
the check author owns that contract.

If a later check changes the workspace, an earlier passing check's stored snapshot no longer
matches the final workspace and is correctly derived as stale. The close response identifies
the later mutating check and asks the caller to rerun or reorder the suite. Cairn does not
silently rerun an expensive suite until it stabilizes.

Freshness is derived:

```text
effective_pass = result == pass && check.snapshot == current_snapshot
```

The stored result is retained for history and displayed as `stale`, not rewritten to
`pending` whenever files change. Legacy passing checks without `snapshot` are stale when the
completion trust gate is enabled.

Manual attestation also records the current snapshot. A later code change makes the review
stale and requires re-attestation.

## Completion gate

New optional configuration:

```yaml
working_state: in_progress
review_state: in_review
completion:
  states: [done]
  require_finished_session: true
  require_verified_evidence: true
  require_fresh_checks: true
session_heartbeat_interval: 30
session_stale_after: 180
snapshot_untracked_file_limit: 104857600
```

Compatibility rule: if `completion` is absent, Cairn preserves the v0 checks behavior.
Newly initialized workspaces receive the stricter defaults above when their configured
states contain `in_progress`, `in_review`, and `done`.

For transitions into `completion.states`, `internal/task.CanTransition` receives pure
`CompletionFacts` prepared by the service layer:

```go
type CompletionFacts struct {
    AttemptID string
    TerminalStatus string
    HasTerminalSummary bool
    HasVerifiedEvidence bool
    ChecksFresh bool
    HasLiveSession bool
}
```

The transition is allowed only when:

- no session is `active` or `blocked`;
- the terminal session in `task.active_attempt` is `finished` (not merely handed off);
- that terminal session's summary is non-empty;
- the current attempt has at least one still-valid Cairn-verified evidence item, or a
  snapshot-bound manual attestation
  explicitly approves `no_change`;
- every command and manual check passes for the current snapshot.

`CanTransition` remains pure: it evaluates supplied facts and config-derived rules. Git,
files, clocks, and session storage stay outside `internal/task`.

Canceled sessions, handed-off sessions, PR URLs that have not been adapter-verified, plain
notes, and agent-authored summaries do not satisfy verified evidence.

## Service and package boundaries

```text
internal/task       pure task/deps/check/completion gates
internal/session    pure session state transitions + validation
internal/snapshot   read-only Git/workspace fingerprinting
internal/store      task store + session store + ephemeral live-state files + repo lock
internal/check      command execution; receives/returns snapshot metadata
internal/mcp        orchestration and MCP schemas
internal/server     HTTP/SSE adapters only
web                 supervision and review UI
```

`internal/session` is a leaf package. Cross-entity completion facts are assembled in the
service and evaluated in `internal/task`, preserving the repo rule that adapters never own
gate logic.

## Cross-process correctness

The web server and one or more MCP servers are separate processes. Session work therefore
uses one persistent `.cairn/write.lock` file with a POSIX advisory exclusive lock (`flock`).

Exact protocol:

1. Open/create the file with mode `0600`; never delete it.
2. Acquire `LOCK_EX | LOCK_NB`, retrying with bounded jitter until the request context's
   five-second lock timeout expires.
3. After acquisition, truncate and write diagnostic JSON (`pid`, `actor`, `operation`,
   `acquired_at`). The JSON is observability only; it never decides ownership.
4. Perform short file mutations, each through the existing temp-file + rename path.
5. Release with `LOCK_UN` and close the descriptor in `defer`.

The kernel releases the lock when a process crashes or closes its descriptor. There is no
stale-lock deletion, lock heartbeat, owner clock comparison, or stale-lock repair pass. This
removes the highest-risk clock-skew/crashed-holder logic. The file is advisory, so **every**
Cairn task/config/session mutation must acquire it. Native network-filesystem and multi-host
coordination are out of scope; Cairn v1 supports separate processes on one POSIX host.

Long operations never hold the lock. `RunChecks` and snapshot hashing execute unlocked, then
acquire the lock, verify the task/check definition version is unchanged and the current
snapshot still equals the recorded after-snapshot, and only then persist the result. A
conflict discards the result and asks the caller to retry.

There is no separate lease store. Under the lock, `begin` scans durable sessions for an
`active` or `blocked` session on the task before creating another one. Session counts are
small and read-fresh behavior matches the existing engine.

Multi-file lifecycle writes use retryable ordering rather than pretending to be atomic:

- `begin`: update task claim/working status, then create the durable session, then write live
  state. A crash before session creation leaves an ordinary claimed/in-progress task; the
  same actor can retry `begin`, while a different actor must explicitly take over.
- `finish`: mark the session terminal first, remove live state, then move a reviewable task
  to the review state. A crash leaves a finished session that is still derived as awaiting
  review; retry completes the task transition.
- `handoff`/`cancel`: mark the session terminal first, remove live state, then release the
  assignee where applicable.

All lifecycle writes accept an idempotency key. Retrying the same operation completes missing
later steps and returns the original session; it never creates a second session. Contention,
crash-between-step, and idempotent-retry tests are part of the first shipping slice, not a
later hardening phase.

## HTTP and event surface

HTTP mirrors MCP:

```text
POST /api/tasks/{id}/sessions/begin
GET  /api/tasks/{id}/sessions
GET  /api/sessions/{session}
POST /api/sessions/{session}/heartbeat
POST /api/sessions/{session}/block
POST /api/sessions/{session}/resume
POST /api/sessions/{session}/evidence
POST /api/sessions/{session}/handoff
POST /api/sessions/{session}/finish
POST /api/sessions/{session}/cancel
```

Filesystem watching expands to `.cairn/sessions/` and `.cairn/live/`. SSE remains a signal
channel:

```json
{"type":"session-changed","task":"PROJ-037","session":"ses_7f42c91e8a32"}
```

The client invalidates task, session, and list queries and refetches canonical DTOs.

## UI

### Navigation

Add high-signal derived views, not new workflow states:

- Active
- Blocked
- Stalled
- Awaiting review

Each shows a count and uses the existing task rows. The inbox emits durable attention events
for blocked, stalled, takeover, failed verification, and awaiting review.

### Task detail

The activity column gains a session timeline:

- actor/client/model and live health;
- current progress and last heartbeat;
- elapsed time and usage when reported;
- branch, commits, PR, and artifact evidence;
- check freshness with the tested/current snapshot short IDs;
- final summary, blocker, or handoff packet;
- actions appropriate to state: resume, cancel, review evidence, attest, complete.

Raw local logs remain expandable and visually secondary. The primary view answers: who is
working, what changed, what is needed, and what proves the result.

### Completion interaction

When completion is blocked, the UI presents specific unmet facts:

- `2 checks are stale — rerun against ws1:8de4`
- `No finished session with a summary`
- `PR recorded but not verified`
- `Manual review was attested before the latest change`

No generic "gate failed" toast.

## Trust boundary

Cairn can prove that a local artifact exists and that configured checks passed against a
specific snapshot. It cannot prove that:

- the configured tests are sufficient;
- an agent-reported model or token count is truthful;
- a remote PR is merged without a provider adapter;
- an agent did not weaken tests before running them.

Those risks are addressed through immutable task check definitions, Git review, optional
human attestation, and future provider adapters—not claims of perfect autonomous safety.

## Failure handling

- Duplicate mutating requests return the current session when their idempotency key matches.
- A heartbeat for a terminal session returns a conflict and does not resurrect it.
- `finish`, `handoff`, and `cancel` are terminal and mutually exclusive.
- A missing worktree makes snapshot-dependent operations fail closed.
- Non-Git repositories may record artifacts and manual checks, but cannot produce `commit`
  or `workspace` evidence until a future pluggable snapshot provider exists.
- Live-state write failure does not lose durable lifecycle state; it adds a warning and the
  session eventually derives as stalled.
- Durable session/task write failure keeps the operation open and retryable.

## Vertical delivery slices

### Slice 1 — observable sessions (ship and evaluate)

- Add the repository `flock`, `internal/session`, one durable session store, and ephemeral
  live-state files.
- Add only `identity`, `begin`, `heartbeat`, `finish`, `cancel`, `get_session`, and
  `list_sessions`.
- Add actor handshake, idempotency, and contention/crash-window tests now.
- Extend SSE/DTOs and ship task-detail session status plus Active, Stalled, and Awaiting
  review views in the same slice.
- Add minimal Codex/Claude/Cursor instructions so real agents can exercise the protocol.
- In this slice, `finish` requires a summary and may move to `review_state`, but it does not
  alter or strengthen the existing v0 close gate.

Checkpoint after dogfooding: measure session adoption, missing finishes, time-to-detect stalled
work, and whether humans open the review surface. Continue only if the visibility loop is
actually useful. No completion behavior changes in this slice.

### Slice 2 — recovery and handoff

- Add `block`, `resume`, and `handoff` with explicit attempt lineage and single-use
  continuation tokens.
- Add Blocked attention events, handoff packets, forced stalled-session takeover, and the
  corresponding UI actions.
- Add lineage, takeover, and idempotent-retry tests.

### Slice 3 — snapshot-bound checks

- Add canonical `internal/snapshot` manifest hashing and portability fixtures.
- Persist post-run check snapshot/time, mutation metadata, and derive `stale`.
- Add `workspace_mutation: allow|forbid` and bind manual attestations to snapshots.
- Surface fresh/stale checks and specific rerun guidance without changing completion gates.

### Slice 4 — evidence-gated completion

- Add evidence verification and `record_evidence`.
- Assemble current-attempt `CompletionFacts` and enforce the new pure completion gate.
- Move reviewable finished work to `review_state` without auto-closing it.
- Add actionable completion-gate explanations.

### Slice 5 — provider evidence and scale

- Optional GitHub/Linear adapters for verified PR evidence.
- Signed CI evidence, indexed read cache if session volume justifies it, and high-volume tests.

## Rejected alternatives

### Make sessions task provenance entries

Rejected. Provenance is a useful audit trail but cannot model live state, cumulative usage,
multiple attempts, typed evidence, or large local logs without turning task files into hot,
conflict-prone append logs.

### Store heartbeats in tracked session files

Rejected. Thirty-second writes would create Git noise and cross-process conflicts. Heartbeat
loss is operationally useful but not durable product history.

### Let `finish` transition directly to `done`

Rejected. It collapses claim, evidence, verification, and approval into an agent assertion—the
exact trust failure this design exists to remove.

### Treat `HEAD` as the tested state

Rejected. Agents commonly run checks before committing. `HEAD` alone would report stale or
false freshness whenever staged, unstaged, or untracked source files participate.

### Require identical snapshots before and after every check

Rejected. Formatters, generators, coverage manifests, and build steps legitimately mutate
non-ignored files. Successful checks bind to their post-run snapshot; strict checks opt into
`workspace_mutation: forbid`. Freshness still prevents an earlier result from validating
later changes.

### Implement ownership with an `O_EXCL` lockfile and stale-lock recovery

Rejected. PID reuse, crashes, clock skew, and competing stale-lock breakers create a small
distributed-systems problem on every write. POSIX `flock` is already crash-released by the
kernel and matches Cairn's documented local macOS/Linux/WSL platform boundary.

### Build an agent launcher into core

Rejected for this phase. It couples Cairn to runtime-specific process, auth, sandbox, and
steering semantics. The protocol remains useful across every agent; launch adapters can be
added outside the rule engine later.

## What to revisit as Cairn grows

- Replace local advisory locking with transactional coordination for multi-host/remote
  workspaces.
- Move large session history to an indexed cache while retaining portable exports.
- Add signed evidence from CI and code-host providers.
- Add policy tiers by task risk, path, or label.
- Add budget enforcement once usage is supplied by trusted runtime adapters.
