---
id: PROJ-01kvxgfk5q23yahc
title: Stop reading the whole board on single-task status updates
status: done
priority: high
labels: [performance, backend]
checks:
  - desc: Go tests pass
    cmd: go test ./internal/...
    timeout: 180
    result: pass
  - desc: Web typecheck/build
    cmd: pnpm -C web build
    timeout: 240
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T19:07:32Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T19:07:38Z', did: began session ses_360973f54bf19b1ff27931b4}
  - {who: 'agent:claude', at: '2026-06-24T19:10:21Z', did: note, text: 'Implemented. Added task.DepResolver + ReadyFunc/CanTransitionFunc (task.go); Ready/CanTransition now wrap them. Service.depResolver() does one store.Get per listed dep. Transition (service.go) and ReadyOf (tools.go) dropped store.List() — a status change now touches only the task, its deps, and config, not the whole board. Chose on-demand dep resolution over an in-memory board cache: smaller, no cache-invalidation surface, and ReadyOf is on every single-task response path so it fixes note/update/get latency too. Frontend useTransition (queries.ts) made optimistic (mirrors useReorder) for instant UI. Also closed the SSE realtime gap in events.go buildEvent. Verified: go test ./... + go vet clean, pnpm build clean, new TestReadyFuncResolvesOnlyListedDeps asserts only listed deps are queried.'}
  - {who: 'agent:claude', at: '2026-06-24T19:10:29Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T19:10:40Z', did: finished session ses_360973f54bf19b1ff27931b4, text: "Status updates no longer re-read the whole board. Root cause: single-task paths called store.List() (full disk scan + whole-graph re-validation) when a status move only needs the task, its listed deps, and config — notes were fast because they skip that.\n\nChanges:\n- task.go: added DepResolver, ReadyFunc, CanTransitionFunc; Ready/CanTransition now wrap them.\n- service.go: Service.depResolver() (one store.Get per dep); Transition dropped store.List().\n- tools.go: ReadyOf resolves only the task's deps — also speeds note/update/get since it's on every single-task response.\n- queries.ts: useTransition is now optimistic (instant UI, rollback on error).\n- events.go: buildEvent keeps task-changed when one task + a coincident session write land in the same window (fixes the original \"task page not realtime\" report).\n\nTrade-off: removed per-transition whole-graph integrity validation; still enforced on list/read paths. Gated review/closed transitions still run command checks synchronously (intentional; not addressed). Verified: go test ./... and go vet clean, pnpm build clean, added TestReadyFuncResolvesOnlyListedDeps. Review focus: confirm dropping graph validation from the transition path is acceptable."}
  - {who: 'agent:claude', at: '2026-06-24T19:17:50Z', did: note, text: 'Measured the real cost — List() was never the bottleneck. Server timings against the running binary: list GET 5.6ms, single-task GET 2ms, non-gated transition (→in_progress) 2.5ms, but gated transition (→in_review/done) 3.2s. PROJ-048''s check 0 is `pnpm build`; gated states run runCmdChecks synchronously inside the request (service.go:311-318), so the whole build blocks the response. That is the user''s "900ms-1s", not the board scan. Per user choice (keep blocking, show progress): exposed config Review() as status.review (server.go), and in TaskDetail the status Select now shows "Running checks…" (vs "Updating…") while a gated move is pending; useTransition skips the optimistic flip for gated targets (isGatedStatus in queries.ts) since checks can refuse. Non-gated moves stay optimistic/instant. Rebuilt bin/cairn (needs server restart to serve status.review). go build/test/vet + pnpm build all clean.'}
  - {who: 'human:shaho', at: '2026-06-24T19:18:17Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:18:17Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T19:18:22Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T19:18:22Z', did: transitioned to in_review}
  - {who: 'agent:claude', at: '2026-06-24T19:19:51Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T19:19:51Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_360973f54bf19b1ff27931b4
---
Status transitions feel ~1s on localhost while notes are instant. Cause: single-task paths re-read and re-parse every task file via `store.List()`, though a status move only needs the task, its direct deps, and config.

## Scope
- `internal/task/task.go` — add `DepResolver`, `ReadyFunc`, `CanTransitionFunc`; `Ready`/`CanTransition` become thin wrappers over them.
- `internal/mcp/service.go` — add `Service.depResolver()`; `Transition` drops `store.List()`, uses `CanTransitionFunc`.
- `internal/mcp/tools.go` — `ReadyOf` resolves only the task's deps (was a full `List()`), fixing note/update/get response latency too.
- `web/src/lib/queries.ts` — optimistic `useTransition` (mirror existing `useReorder`).

## Notes
- Also fixed an SSE realtime gap in `internal/server/events.go` `buildEvent`: one task + a coincident session write now stays `task-changed` instead of downgrading to a list-only refresh.
- Gated review/closed transitions still run command checks synchronously by design — out of scope.

## Acceptance
- `go test ./internal/...` green (deps gate, checks gate, dangling-dep rejection preserved).
- A non-gated transition no longer calls `store.List()`.