---
id: PROJ-029
title: Onboarding + unified empty states
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: onboarding + empty states look good (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T15:42:35Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T15:42:49Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T15:42:53Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T15:46:51Z', did: note, text: 'Shared EmptyState component (icon/title/message/action) generalized from Board. Onboarding component shows on an empty workspace: welcome + 3 one-glance points + "Create your first task" CTA + collapsible "Connect your AI agent" with the claude mcp add snippet (built from status.root) + copy button. Board uses Onboarding when tasks==0 and EmptyState for filter-empty; BoardView and Graph now show EmptyState instead of blank. Verified on a fresh temp workspace via Playwright: onboarding, expanded connect snippet, empty board all render cleanly. pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T15:46:59Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T15:47:07Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T15:47:07Z', did: transitioned to done}
assignee: agent:claude
---
Make a fresh workspace welcoming and unify empty states.

- New Onboarding component (web/src/components/Onboarding.tsx): one-glance what-cairn-is, "Create your first task" CTA (opens CreateTaskDialog via onNewTask), and a collapsible "Connect your AI agent" with the `claude mcp add cairn -- "<root>/bin/cairn" serve --actor agent:claude-1 --repo "<root>"` snippet + copy button (root from status). Shown in Board when tasks.length===0.
- Generalize EmptyState (from Board.tsx) into web/src/components/EmptyState.tsx (icon+title+message+optional action).
- Use it in Board (filter-empty), BoardView (currently blank), Graph (currently blank). Copy nudges the agent angle.</body>
<parameter name="labels">["frontend", "ux"]