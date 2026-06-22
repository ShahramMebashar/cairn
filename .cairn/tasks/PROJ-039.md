---
id: PROJ-039
title: Refine session design from review
status: done
priority: urgent
labels: [design, sessions, trust]
checks:
  - desc: repository checks pass
    cmd: make check
    result: pass
  - desc: design diff is clean
    cmd: git diff --check
    result: pass
provenance:
  - {who: 'agent:codex', at: '2026-06-22T18:14:19Z', did: created}
  - {who: 'agent:codex', at: '2026-06-22T18:15:20Z', did: claimed}
  - {who: 'agent:codex', at: '2026-06-22T18:15:25Z', did: transitioned to in_progress}
  - {who: 'agent:codex', at: '2026-06-22T18:18:41Z', did: note, text: 'Applied review by reducing the MVP to one durable session store plus ephemeral live state, removing custom leases/session logs, and making the first delivery slice visibly useful in the UI. Chose kernel-released POSIX flock for all writes, explicit attempt IDs with single-use handoff continuation, canonical manifest snapshots, and post-run check binding with optional workspace_mutation: forbid.'}
  - {who: 'agent:codex', at: '2026-06-22T18:19:32Z', did: ran checks}
  - {who: 'agent:codex', at: '2026-06-22T18:19:39Z', did: note, text: 'Refined the session/trust design from Claude review: declared the control-plane pivot and dogfood checkpoint; reduced MVP storage/verbs; replaced custom lease recovery with crash-released POSIX flock; defined attempt IDs, active_attempt, single-use handoff continuation, and evidence eligibility; switched to canonical portable manifests that exclude .cairn; bound passing checks to post-run snapshots with optional mutation forbidding; and recut delivery into early-value vertical slices. Verified with make check and git diff --check.'}
  - {who: 'agent:codex', at: '2026-06-22T18:19:44Z', did: transitioned to done}
assignee: agent:codex
---
Resolve the implementation risks identified in review of the agent sessions and trust-loop design.

## Scope
- Remove mutation-equality friction from check freshness semantics.
- Define attempt/session lineage and evidence eligibility precisely.
- Specify repository write locking and crash behavior.
- Recut implementation into early-value vertical slices.

## Acceptance
- The design is implementation-ready on all four points without weakening completion trust.
