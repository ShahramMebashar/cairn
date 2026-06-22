---
id: PROJ-031
title: 'Backend: per-request human actor + suggestedActor'
status: done
checks:
  - desc: go check passes
    cmd: make check
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-22T16:10:15Z', did: created}
  - {who: 'agent:claude', at: '2026-06-22T16:10:30Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-22T16:10:34Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-22T16:13:21Z', did: note, text: 'repo.DeriveActor() suggests human:<os-username> (os/user, sanitized, Windows DOMAIN\user handled, fallback human:dev). server.service now takes an actor; actorFor(r) reads X-Cairn-Actor (URL-decoded, fallback ?actor=) → sanitizeActor (single-line, trimmed, ≤64) → fallback s.actor; svcFor passes it so all writes are per-actor, reads unaffected. statusResp.SuggestedActor = DeriveActor(). Tests: per-request actor attributes provenance to human:ali, falls back to human:test without a header, suggestedActor is human:*, sanitizeActor strips newlines/empty. make check green.'}
  - {who: 'agent:claude', at: '2026-06-22T16:13:30Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-22T16:13:30Z', did: transitioned to done}
assignee: agent:claude
---
Let each human have their own identity instead of one fixed human:web.

- repo.DeriveActor(): os/user.Current() -> "human:<sanitized username>", fallback human:dev.
- server: actorFor(r) reads X-Cairn-Actor header (URL-decoded; fallback ?actor=), sanitizeActor (trim, cap 64, drop control/newline), fall back to s.actor. service(root,actor); svcFor passes actorFor(r). Writes become per-actor; reads unaffected.
- statusResp.SuggestedActor = repo.DeriveActor().
- Tests: write with X-Cairn-Actor:human:ali records who==human:ali; no header falls back; /api/status suggestedActor starts with human:; sanitizeActor rejects newline.</body>
<parameter name="labels">["backend", "identity"]