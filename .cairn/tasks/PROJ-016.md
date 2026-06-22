---
id: PROJ-016
title: WYSIWYG markdown editor for note + new-task composers
status: done
checks:
  - desc: web builds
    cmd: pnpm build
    cwd: web
    result: pass
  - desc: editor feels good (manual)
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T05:42:48Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T05:42:48Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T05:42:48Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T05:48:29Z', did: note, text: '`MarkdownEditor.tsx` (TipTap v3 StarterKit + Placeholder + tiptap-markdown). WYSIWYG with a compact toolbar and markdown shortcuts; value/onChange are markdown strings (getMarkdown/setContent round-trip). Replaced the Write/Preview textareas in CreateTaskDialog and the TaskDetail note composer. Verified headings/bold/lists/code-block render live.'}
  - {who: 'agent:claude', at: '2026-06-22T05:48:29Z', did: attested, text: check 1 pass}
  - {who: 'agent:claude', at: '2026-06-22T05:48:32Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T05:48:32Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-22T12:54:16Z', did: updated}
  - {who: 'human:web', at: '2026-06-22T12:54:18Z', did: updated}
  - {who: 'human:web', at: '2026-06-22T12:54:19Z', did: updated}
  - {who: 'human:web', at: '2026-06-22T12:54:21Z', did: updated}
assignee: agent:claude
labels: [kk, jj]
---
Replace the Write/Preview textareas with a TipTap WYSIWYG that reads/writes **markdown**.

- `MarkdownEditor.tsx`: TipTap + StarterKit + tiptap-markdown + Link + Placeholder
- Compact toolbar (bold, italic, code, H1/H2, lists, link, code block); markdown shortcuts (#, -, ```)
- value/onChange stay markdown strings (getMarkdown / setContent)
- Use in CreateTaskDialog body and the TaskDetail note composer; WYSIWYG only (no source toggle)