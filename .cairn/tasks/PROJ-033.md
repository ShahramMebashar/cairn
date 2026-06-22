---
id: PROJ-033
title: Render task bodies like Linear (prose polish)
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: reads like Linear (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T16:32:08Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T16:32:42Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T16:32:47Z', did: transitioned to in_progress}
  - {who: 'human:shaho', at: '2026-06-22T16:34:05Z', did: attested, text: check 1 pass}
  - {who: 'human:shaho', at: '2026-06-22T16:34:09Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T16:37:19Z', did: note, text: 'Added a `.prose` block to web/src/style.css: removes @tailwindcss/typography''s backtick pseudo-content on inline code, styles inline code as a subtle bordered chip (bg-muted + border, ~0.82em, mono), and tunes h1/h2/h3 hierarchy + list/paragraph rhythm — all on design tokens, light + dark. Also `.cairn-md-inline code` for the note inline variant. Simplified Markdown.tsx InlineCode to a bare <code> (styling now shared in CSS); Shiki CodeBlock unchanged. Verified via Playwright on a rich demo body (headings + inline code + go fenced block) in light and dark: no literal backticks, clear hierarchy, Linear-like chips, contained highlighted code block. pnpm build green.'}
  - {who: 'agent:claude', at: '2026-06-22T16:38:59Z', did: note, text: 'Follow-up: fixed light-mode code blocks. Shiki runs with defaultColor:false, so it sets no color itself — style.css only had the .dark rules, leaving light-mode code faint/uncolored. Added the light-mode base (.shiki + .shiki span → var(--shiki-light)/(--shiki-light-bg) + font-style/weight/decoration). Verified: light-mode go block now has proper github-light highlighting (keywords/functions colored, comment greyed) on a light bg, matching dark mode.'}
  - {who: 'agent:claude', at: '2026-06-22T16:39:05Z', did: transitioned to done}
assignee: agent:claude
---
Inline code shows literal backticks and headings lack hierarchy — make bodies/notes read like Linear.

## Cause
`@tailwindcss/typography` adds `code::before/::after { content: "\`" }` to every `code` inside `.prose`; `style.css` never overrides it, so the chip in `Markdown.tsx` still gets wrapped in backticks.

## Scope
- `web/src/style.css`: add a `.prose` block (light + dark) — drop the backtick pseudo-content, style inline `code` as a subtle chip (`bg-muted` + `border`, rounded, ~`0.82em`), tune `h1`/`h2`/`h3` hierarchy + list/paragraph rhythm. Tokens only.
- `web/src/components/Markdown.tsx`: lean `InlineCode` on the shared CSS; keep Shiki `CodeBlock`.

## Acceptance
- No literal backticks; clear heading hierarchy; Linear-like inline chips; contained code blocks (light + dark).</body>
<parameter name="labels">["frontend", "ux"]