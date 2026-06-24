---
id: PROJ-01kvxj6cfxj513n2
title: Shorten task IDs to 10 chars (6 time + 4 random)
status: done
priority: low
labels: [backend, ids]
checks:
  - desc: Go tests pass
    cmd: go test ./...
    timeout: 180
    result: pass
  - desc: Web typecheck/build
    cmd: pnpm -C web build
    timeout: 240
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T19:37:28Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T19:37:37Z', did: began session ses_2224a25f2c5f288e7161a64c}
  - {who: 'agent:claude', at: '2026-06-24T19:39:39Z', did: note, text: 'Done. id.go: idTimeChars 10→6, idRandChars 6→4, idEpoch=1704067200 (2024-01-01Z); time prefix now encodes at.Unix()-idEpoch (seconds), ~34yr range to ~2058. New ids are PREFIX + 10 chars, e.g. PROJ-k3m9x7q2vw. filter.ts effectiveRank: discriminates by suffix length (≤10 = current sec-since-2024, else legacy 16-char ms-since-1970) and normalizes both to absolute ms so all three eras (legacy numeric, old time, new time) sort together. Tests: four {16}→{10} regexes (id_test, store_test, service_test, server_test/WEB-); monotonic check now uses a 1-second gap and compares the 6-char prefix (seconds resolution). Tradeoff accepted: no sub-second creation-order tie-break (display order is manual rank anyway). go test ./... + pnpm build green; rebuilt bin/cairn (needs restart for the running MCP/web server to mint short ids).'}
  - {who: 'agent:claude', at: '2026-06-24T19:39:50Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T19:40:00Z', did: finished session ses_2224a25f2c5f288e7161a64c, text: "Task IDs shortened from 16 to 10 chars (6 time + 4 random).\n\n- internal/store/id.go: idTimeChars 10→6, idRandChars 6→4, new idEpoch=1704067200 (2024-01-01Z). Time prefix encodes seconds since that epoch (~34yr range to ~2058). New ids look like PROJ-k3m9x7q2vw.\n- web/src/lib/filter.ts: effectiveRank now handles three eras — legacy numeric (sorts first), legacy 16-char (ms-since-1970), current 10-char (sec-since-2024) — discriminating by suffix length and normalizing both time formats to absolute ms so they interleave correctly.\n- Tests: four {16}→{10} regexes; monotonic test uses a 1-second gap at seconds resolution.\n\nTradeoff (accepted by user): seconds precision drops sub-second creation-order tie-breaking; display order is governed by manual rank regardless. Existing 16-char ids are untouched and still sort correctly alongside new ones.\n\nVerified: go test ./... and pnpm build green. Rebuilt bin/cairn — running MCP/web server must restart to start minting short ids. Review focus: confirm mixed old/new ordering on a real board looks right after restart."}
  - {who: 'human:shaho', at: '2026-06-24T20:21:00Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T20:21:11Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T20:21:11Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_2224a25f2c5f288e7161a64c
rank: 1.5
---
Task IDs are 16-char suffixes (10 ms-since-1970 + 6 random) — longer than needed. Shorten to 10 chars: 6 time (seconds since a 2024 epoch, ~34yr range) + 4 random (20-bit tie-break).

## Scope
- `internal/store/id.go` — idTimeChars 10→6, idRandChars 6→4, add idEpochSec (2024-01-01), encode `at.Unix()-epoch`.
- `web/src/lib/filter.ts` — effectiveRank: decode both old (16-char, ms-since-1970) and new (10-char, sec-since-2024) to absolute ms so eras sort together; length discriminates.
- Tests: regex `{16}`→`{10}` in id_test.go, store_test.go, service_test.go, server_test.go (WEB-); monotonic check uses a 1-second gap (seconds precision).

## Notes
- Existing 16-char IDs are untouched; sorting across old/new/legacy-numeric preserved.
- Tradeoff: loses sub-second creation-order tie-breaking (display order governed by manual rank anyway).

## Acceptance
- New IDs match `^PREFIX-[0-9a-z]{10}$`; go test + pnpm build green.