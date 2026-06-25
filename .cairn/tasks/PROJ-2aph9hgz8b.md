---
id: PROJ-2aph9hgz8b
title: Humanize docs prose (concise, no em-dashes, no AI filler)
status: in_review
priority: medium
checks:
  - desc: No em-dashes remain in docs prose (excludes node_modules)
    cmd: '! grep -rl ''—'' docs --include=''*.md'' --exclude-dir=node_modules && echo OK'
    timeout: 30
    result: pass
  - desc: Docs build clean
    cmd: npm --prefix docs run build
    timeout: 300
    result: pass
  - desc: 'Manual: prose reads human and concise'
    type: manual
    result: pending
provenance:
  - {who: 'agent:claude', at: '2026-06-25T15:13:53Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T15:14:00Z', did: began session ses_b8c98419824a8ff456fc8838}
  - {who: 'agent:claude', at: '2026-06-25T15:22:27Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T15:23:20Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-25T15:24:30Z', did: updated}
  - {who: 'agent:claude', at: '2026-06-25T15:24:40Z', did: ran checks}
  - {id: n_sqrmptsh, who: 'agent:claude', at: '2026-06-25T15:24:51Z', did: note, text: 'Humanized all docs per the user''s feedback (concise, human, no em-dashes, no AI filler). Saved the preference as a feedback memory. Rewrote the visitor funnel by hand (index, introduction, installation, quickstart) in the new voice. Three parallel subagents swept the rest (guides ×4, agents ×9, reference ×4), removing every em-dash and cutting hype words while preserving frontmatter/code/tables/links/facts. Fixed the source error string in internal/check/check.go to drop its em-dash (": " / ". " instead) and matched the quote in checks-and-gates.md, so the product output is em-dash-free too. Also cleaned README/CHANGELOG/SECURITY em-dashes (AGENTS.md left as internal contributor doc). Verified zero em-dashes across docs + README/CHANGELOG/SECURITY; go test ./internal/check green; docs build clean. Note: the original em-dash check scanned docs/node_modules (33 dependency READMEs) — fixed the check to --exclude-dir=node_modules; cairn''s sh-grep recurses node_modules whereas an interactive ripgrep alias skips it.'}
  - {who: 'agent:claude', at: '2026-06-25T15:25:06Z', did: finished session ses_b8c98419824a8ff456fc8838, text: "Humanized the docs: concise, plain voice, no em-dashes, no AI filler. Saved the preference as a memory so it sticks.\n\n- Rewrote the visitor funnel by hand in the new voice: index, introduction, installation, quickstart.\n- Three parallel subagents swept the rest (4 guides, 9 agent pages, 4 reference pages): removed every em-dash, cut hype words, kept all frontmatter, code, tables, links, and facts intact.\n- Removed the em-dash from cairn's own shell-not-found error string (internal/check/check.go) and matched the quote in checks-and-gates.md, so even the product output is clean.\n- Cleaned em-dashes in README, CHANGELOG, SECURITY. Left AGENTS.md (internal contributor guide) as-is.\n\nVerified: zero em-dashes across docs + README/CHANGELOG/SECURITY; go test ./internal/check green; docs build clean.\n\nSide fix: the acceptance check originally scanned docs/node_modules (33 dependency READMEs contain em-dashes), so it false-failed. Added --exclude-dir=node_modules. Worth knowing: cairn's checks run real sh `grep -r`, which recurses node_modules, whereas an interactive shell's ripgrep alias skips gitignored dirs. Same lesson applies to any future grep-based check.\n\nGoing forward I'll keep prose plain and avoid em-dashes in chat and commits too, per the saved feedback."}
assignee: agent:claude
active_attempt: att_b8c98419824a8ff456fc8838
---
Rewrite the docs in a plain, human voice per the user's feedback. No em-dashes (—). Cut filler/hype words. Short, direct sentences.

Start with the visitor funnel (index, introduction, installation), then sweep guides/agents/reference. README too.

## Acceptance
- No em-dashes in docs/*.md; tighter prose; docs build clean.