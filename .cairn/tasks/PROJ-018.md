---
id: PROJ-018
title: Command palette (Cmd-K) + keyboard navigation
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: feels pro (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T07:32:53Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T07:45:08Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T07:45:08Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T07:57:27Z', did: note, text: 'CommandPalette (cmdk) on Cmd/Ctrl-K: create, jump-to-task (fuzzy), switch view, graph, theme, switch folder. Board keyboard nav: j/k select (ring highlight + scrollIntoView), enter/o open, c new, / focus search. Fixed: shadcn radix-nova CommandDialog omits the <Command> root, so wrapped content in <Command> (was crashing on open: undefined subscribe). Verified Ctrl-K + j/Enter.'}
  - {who: 'agent:claude', at: '2026-06-22T07:57:27Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T07:57:30Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T07:57:30Z', did: transitioned to done}
  - {who: 'agent:claude', at: '2026-06-22T08:06:30Z', did: updated}
assignee: agent:claude
priority: low
---
shadcn command; CommandPalette (create, jump to task, switch view/folder, theme, graph). Board keys: j/k move, enter/o open, c create, / search, g graph. Stretch: x multi-select + bulk bar.