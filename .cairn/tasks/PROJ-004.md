---
id: PROJ-004
title: 'Runs endpoint: GET /api/tasks/{id}/runs'
status: done
checks:
  - desc: endpoint parses a seeded run log into JSON, newest-first
    cmd: go test ./internal/server
    cwd: .
    result: pass
  - desc: gofmt + vet + test pass
    cmd: make check
    cwd: .
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-21T19:47:42Z', did: created}
  - {who: 'agent:claude', at: '2026-06-21T19:50:38Z', did: claimed}
  - {who: 'agent:claude', at: '2026-06-21T19:50:44Z', did: transitioned to in_progress}
  - {who: 'agent:claude', at: '2026-06-21T19:52:56Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T19:53:04Z', did: transitioned to done}
  - {who: 'human:web', at: '2026-06-21T19:53:18Z', did: ran checks}
  - {who: 'agent:claude', at: '2026-06-21T19:55:32Z', did: note, text: 'Done in internal/server/runs.go + route in server.go. Chose filename-glob over a schema change (honors frozen SPEC §149): GET /api/tasks/{id}/runs globs .cairn/runs/<id>-*.log, parses the writeLog header, sorts newest-first lexically. `at` derived from the filename stamp. Unparseable header degrades to raw body. Tests: TestRunsEndpointParsesLogsNewestFirst, TestRunsEndpointEmptyWhenNoRuns.'}
  - {who: 'human:shaho', at: '2026-06-24T18:45:05Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:45:05Z', did: transitioned to in_review}
  - {who: 'human:shaho', at: '2026-06-24T18:45:08Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-24T18:45:08Z', did: transitioned to done}
assignee: agent:claude
rank: !!float 2
---
Expose captured check output without changing the task-file schema (SPEC §149 frozen).

- `GET /api/tasks/{id}/runs` lists `.cairn/runs/<id>-*.log` newest-first.
- Parse each log header (cmd, cwd, exit, timedout, duration) + body output.
- Return `{ runs: [{ file, at, cmd, cwd, exit, timedout, duration, output }] }`.
- Missing runs dir → empty list (not error); unparseable log → skip header, keep raw body.
- Wire route in server.Handler().

See spec §B.