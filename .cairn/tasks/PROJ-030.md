---
id: PROJ-030
title: How-it-works explainer + agent/human visibility
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: explainer + glyphs look good (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T15:42:44Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T15:47:13Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T15:47:17Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T15:51:47Z', did: note, text: 'Added actorKind() to utils and a shared Assignee component (avatar + Bot badge for agent:* / User badge for human:*, tooltip "AI agent · <actor>"). Wired into TaskRow, BoardView card, and TaskDetail properties. HelpDialog (shadcn Dialog) explains Tasks & states, the two gates, You + your agent, and Provenance; opened from a new Help (?) button in the sidebar footer. Verified via Playwright on the real repo: help dialog reads well, agent Bot glyph shows on the assignee avatar. make check + pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T15:51:52Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T15:52:02Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T15:52:02Z', did: transitioned to done}
assignee: agent:claude
---
Make the model discoverable and the AI pairing visible.

- HelpDialog (web/src/components/HelpDialog.tsx): shadcn Dialog opened from a Help (?) entry in AppSidebar footer. Scannable sections: Tasks & states, Two gates (deps + checks), You + your agent (MCP/Claude Code, claim/handoff), Provenance. Reuse the connect snippet.
- Agent-vs-human glyph: actorKind(assignee) -> agent|human|null in lib/utils.ts; small Bot (agent:*) / User (human:*) glyph wherever assignee shows — TaskRow, BoardView card, TaskDetail properties. Reuse Avatar/initials.</body>
<parameter name="labels">["frontend", "ux"]