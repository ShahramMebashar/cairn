---
id: PROJ-048
title: Make the task checks panel legible and scannable
status: done
priority: medium
labels: [web, design, ui]
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: 'design review: checks panel reads clearly, matches Linear aesthetic + design tokens'
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-24T17:47:27Z', did: created}
  - {who: 'agent:claude', at: '2026-06-24T17:47:35Z', did: began session ses_6527d717c6c494c0d45cee83}
  - {who: 'agent:claude', at: '2026-06-24T17:48:55Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-24T17:49:04Z', did: note, text: 'Redesigned the checks panel in web/src/pages/TaskDetail.tsx. Checks now render in a bordered, divided list (matching the Sub-tasks pattern) so they read as one distinct, scannable group. Each row leads with a status icon (CircleCheck/success, CircleX/destructive, hollow Circle/muted pending, Loader2 while a run is in flight) and ends with a soft status pill (bg-success/10 text-success etc.) instead of bare colored text. Header gained an n/n passing summary (turns success-green when all pass) and the Run button is now variant=outline so it reads as an action. Collapsible log + manual attest behavior preserved; expandable command rows get a hover state and the log panel sits in a bordered bg-muted/30 area. All styling via existing design tokens/shadcn. Left the manual "design review" check pending for human visual attest. pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-24T17:49:13Z', did: finished session ses_6527d717c6c494c0d45cee83, text: 'Reworked the task checks panel (web/src/pages/TaskDetail.tsx) from bare text into a scannable, Linear-style group. Bordered/divided list; each row leads with a status icon (✓ success / ✗ destructive / hollow pending / spinner running) and ends with a soft status pill; header shows an n/n passing summary and an outline Run button. Collapsible logs + manual attest preserved, all via existing design tokens. pnpm build green. The manual "design review" check is intentionally left pending — it''s yours to attest after a visual look (open a task with checks, e.g. PROJ-047, in the web UI). Once you''re happy, attest it and close to done.'}
  - {who: 'human:shaho', at: '2026-06-24T18:44:15Z', did: attested, text: check 1 pass}
  - {who: 'human:shaho', at: '2026-06-24T18:44:26Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:44:33Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:44:33Z', did: transitioned to done}
  - {who: 'human:shaho', at: '2026-06-24T18:44:41Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:44:41Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T18:44:47Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:44:47Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-24T19:11:55Z', did: transitioned to in_progress}
  - {who: 'human:web', at: '2026-06-24T19:12:41Z', did: ran checks}
  - {who: 'human:web', at: '2026-06-24T19:12:41Z', did: transitioned to in_review}
  - {who: 'human:web', at: '2026-06-24T19:13:10Z', did: ran checks}
  - {who: 'human:web', at: '2026-06-24T19:13:10Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_6527d717c6c494c0d45cee83
---
The checks UI (`web/src/pages/TaskDetail.tsx`) is bare text — a new user may not realize what it is or that it gates closing. Give it visual hierarchy, matching the Linear-style system (shadcn + `style.css` tokens).

## Scope
- Wrap checks in a bordered, divided list (reuse the Sub-tasks `divide-y rounded-lg border` pattern).
- Per-row status icon: `CircleCheck`/success, `CircleX`/destructive, `Circle`/muted pending, `Loader2` while running.
- Replace bare colored "pass" text with a soft status pill (`bg-success/10 text-success`, etc.).
- Header: add `n/n` passing summary; make Run a bordered (`outline`) button.
- Keep existing collapsible log + manual attest behaviour.

## Acceptance
- Checks read as a distinct, scannable group with obvious status; web builds.