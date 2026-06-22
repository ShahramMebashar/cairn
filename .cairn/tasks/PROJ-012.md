---
id: PROJ-012
title: 'Log viewer: ANSI colors + copy + wrap toggle'
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: colored logs look clean (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:45:37Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:50:03Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:50:03Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:51:30Z', did: note, text: 'New LogView.tsx renders run output through anser (ANSI->HTML, escaped) with exit/duration header, copy button, and soft-wrap toggle. Replaced the bare <pre> in CheckRow. `pnpm build` green.'}
  - {who: 'agent:claude', at: '2026-06-21T21:01:11Z', did: note, text: 'LogView renders run output via anser (ANSI->HTML, escaped) with copy + wrap toggle; replaced bare <pre> in CheckRow. Verified colored/clean in light+dark.'}
  - {who: 'agent:claude', at: '2026-06-21T21:01:11Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-21T21:01:13Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T21:01:13Z', did: transitioned to done}
assignee: agent:claude
---
Make run logs readable.

- `LogView.tsx`: render output via anser (ANSI->HTML, escaped)
- Copy-to-clipboard + soft-wrap toggle
- Replace the bare <pre> in CheckRow