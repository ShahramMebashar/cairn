# Security Policy

**Report vulnerabilities privately via GitHub Security Advisories**. On this repository, open
the *Security* tab → *Report a vulnerability*. This routes to the maintainer with audit-logged
access and never appears in public issues. Please don't open a public issue for a security bug.

- Supported versions: cairn is **pre-1.0**; only the **latest release** is supported.
- Public disclosure window: **90 days** from acknowledgement.

## Trust model

cairn is a **local, single-user tool**, and its security model reflects that:

- **No authentication, by design.** `cairn web` and `cairn serve` bind to `127.0.0.1` and have
  no auth. The trust model is like a git author. Anyone who can reach the port can call the API.
  **Do not expose the port to a network you don't control.**
- **Local file access is intentional.** Endpoints accept a `?path=`/`?repo=` for the project to
  operate on; this is deliberate for a local tool. Paths must resolve to an existing directory.

> [!WARNING]
> **Checks run arbitrary shell. Trust your repos.** A task's `cmd` check is executed via
> `sh -c`, and **closing a task auto-runs its checks**. A task file is just Markdown in a repo,
> so a malicious `.cairn/tasks/*.md` can run arbitrary commands the moment you run checks or
> close that task. Only open and operate cairn on repositories whose task files you trust, the
> same caution you'd apply to running a repo's `Makefile` or test suite.

## Guarantees

- **Identity integrity.** Every write is stamped with the server's `--actor`. `begin` requires
  `expected_actor` to match the connection's actor exactly and **refuses before writing** on a
  mismatch. A misconfigured client can't silently record work under the wrong identity.
- **Safe config writes.** When cairn writes an agent's MCP config (the Connect feature) it
  merges **only** its own `cairn` entry, preserving every other server and key; writes are
  atomic (temp file + rename), the prior file is backed up to `<file>.bak`, and the result is
  re-read and verified. Disconnect removes only the `cairn` entry.
- **Cross-process safety.** `.cairn/write.lock` is an OS **advisory** lock that serializes writes
  across processes; file existence is never treated as ownership.
- **No silent identity fallback for agents.** Connecting an agent writes its own identity
  (`agent:<id>`); the human request actor is never substituted into an agent's config.

## What to report

- Path traversal or writes/reads **outside** the intended workspace or config-file paths.
- Identity bypass: recording writes under an actor the caller didn't authorize.
- Command execution **beyond** the documented check mechanism (e.g. injection via task fields
  that runs without you invoking checks or closing a task).
- The HTTP/MCP server being reachable or abusable **beyond localhost** in a default config.

## Out of scope

By design (and with no version commitment pre-1.0):

- The deliberate **no-auth localhost** model.
- **Arbitrary shell from checks you authored** in your own repo. You control your task files.
  (Untrusted repos are the risk; see the warning above.)
- Exposing the port to an untrusted network yourself.
