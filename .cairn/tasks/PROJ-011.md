---
id: PROJ-011
title: Markdown rendering + Shiki code highlight (body, notes, composers)
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: matches Linear aesthetic (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T20:45:37Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T20:46:07Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T20:46:07Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T20:47:54Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:50:03Z', did: note, text: 'Markdown via react-markdown + remark-gfm; Shiki highlights each fenced block in an async child component (react-markdown v9 is sync, so an async rehype plugin would not run). Dual github light/dark themes via CSS vars. Body + notes (inline) + Write/Preview tabs in create & note composers. `pnpm build` green.'}
  - {who: 'agent:claude', at: '2026-06-21T20:57:21Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-21T20:57:29Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T20:58:50Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T21:01:16Z', did: note, text: Verified rich markdown (headings/lists/bold/inline code) + Shiki code blocks in body and inline markdown in notes; Write/Preview composers. Light+dark confirmed via screenshots.}
  - {who: 'agent:claude', at: '2026-06-21T21:01:16Z', did: transitioned to done}
assignee: agent:claude
---
Render task content as rich markdown with VS Code-grade highlighting.

- `Markdown.tsx`: react-markdown + remark-gfm + @shikijs/rehype (dual light/dark)
- Use in task **body** and provenance **notes**
- Write/Preview tabs in create + note composers
- Tune prose to the neutral/Linear aesthetic