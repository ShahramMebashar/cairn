---
id: PROJ-032
title: 'Frontend: editable human identity in web UI'
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: identity works (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T16:10:24Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T16:13:37Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-22T16:19:17Z', did: note, text: asd}
  - {who: 'human:shahot', at: '2026-06-22T16:19:25Z', did: note, text: asd}
  - {who: 'human:shaho', at: '2026-06-22T16:19:30Z', did: attested, text: check 1 pass}
  - {who: 'human:shaho', at: '2026-06-22T16:19:34Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T16:21:47Z', did: note, text: 'lib/identity.ts: currentActor/setActor/displayName/toActor + useIdentity(suggested) (useSyncExternalStore, seeds localStorage cairn-actor from suggestedActor via effect, syncs across tabs). api.ts: req() sends X-Cairn-Actor=encodeURIComponent(currentActor()); Status.suggestedActor added. AppSidebar: IdentityChip in footer (Assignee glyph + name) opens a shadcn Popover to rename → setActor(toActor(name)); NotificationBell now gets the live identity actor. Added shadcn popover component. Verified E2E via Playwright on a temp repo: chip seeded from OS user, renamed to ''shahram'' (shows SH + human glyph), created a task → provenance who=human:shahram. pnpm build + make check green.'}
  - {who: 'agent:claude', at: '2026-06-22T16:22:04Z', did: transitioned to done}
---
Surface and edit the current human identity; send it on writes.

- lib/identity.ts: currentActor()/setActor()/displayName()/toActor() + useIdentity(suggested) hook (effective = stored ?? suggested; seed localStorage cairn-actor when suggested arrives; re-render on change).
- api.ts: req() sends X-Cairn-Actor=encodeURIComponent(currentActor()); Status.suggestedActor.
- App.tsx: seed identity from status.suggestedActor on load when unset.
- AppSidebar footer: "You · <name>" chip (Assignee glyph) opens a Popover (add shadcn popover if missing) to edit name -> setActor(toActor(name)).
- NotificationBell gets the live identity actor (assigned-to-me).</body>
<parameter name="labels">["frontend", "identity"]