---
id: PROJ-2apbf857vy
title: Complete cross-OS desktop bundling (macOS, Windows, Linux)
status: done
priority: high
checks:
  - desc: Sidecar cross-compiles to the requested triple's arch (amd64 on arm host)
    cmd: TARGET_TRIPLE=x86_64-apple-darwin node scripts/build-sidecar.mjs && file desktop/src-tauri/binaries/cairn-x86_64-apple-darwin | grep -qi 'x86_64' && echo OK
    timeout: 180
    result: pass
  - desc: tauri.conf.json is valid JSON with all-OS bundle targets
    cmd: 'node -e "const c=require(''./desktop/src-tauri/tauri.conf.json'');const t=c.bundle.targets;if(![''nsis'',''deb'',''appimage''].every(x=>t.includes(x)))throw new Error(''missing OS targets: ''+JSON.stringify(t));console.log(''targets ok'',t)"'
    timeout: 30
    result: pass
  - desc: 'Manual: a tagged release produces installers for macOS, Windows, and Linux'
    type: manual
    result: pass
provenance:
  - {who: 'agent:claude', at: '2026-06-25T13:34:32Z', did: created}
  - {who: 'agent:claude', at: '2026-06-25T13:34:39Z', did: began session ses_52784d175e94376a01882ea4}
  - {who: 'agent:claude', at: '2026-06-25T13:40:04Z', did: ran checks}
  - {id: n_t7cd6ytd, who: 'agent:claude', at: '2026-06-25T13:40:47Z', did: note, text: 'Cross-OS bundling done — and it required a real Windows port (the matrix had Windows but the Go sidecar didn''t compile there). Three config fixes: (1) tauri.conf.json bundle.targets ["dmg","app"] → ["app","dmg","nsis","deb","appimage"] so Windows/Linux actually bundle; (2) scripts/build-sidecar.mjs now derives GOOS/GOARCH (+CGO_ENABLED=0) from the Tauri target triple — previously it named the binary per-triple but built for the host arch, so x86_64-mac on an arm64 runner would embed an arm64 binary; (3) release.yml gained workflow_dispatch (build-only; tagName/version stamping guarded to tag pushes). Windows port: split Unix-only syscalls behind build tags — internal/store/{flock_unix.go,flock_windows.go} (flock vs LockFileEx, shared errWouldBlock) and internal/check/{kill_unix.go,kill_windows.go} (Setpgid+SIGKILL group vs Process.Kill). lock.go/check.go now call portable helpers. go mod tidy promoted go-toml to a direct dep. Verified: GOOS in {windows,linux,darwin} × amd64 all `go build ./...` clean; make check green; sidecar cross-compiles produce correct-arch binaries (Mach-O x86_64, ELF x86-64, PE32+ x86-64). Checks run via `sh -c` still — Windows needs Git Bash/WSL (already documented). Code signing (Apple notarization / Windows cert) remains TODO; updater artifacts are signed.'}
  - {who: 'agent:claude', at: '2026-06-25T13:41:04Z', did: finished session ses_52784d175e94376a01882ea4, text: "Completed cross-OS desktop bundling for macOS, Windows, and Linux. The release matrix already had all three OSes, but two things meant it couldn't actually ship installers everywhere — and one was a real Windows port.\n\nConfig fixes:\n- tauri.conf.json: bundle.targets [\"dmg\",\"app\"] (macOS-only) → [\"app\",\"dmg\",\"nsis\",\"deb\",\"appimage\"], so Windows (NSIS) and Linux (deb + AppImage) bundle.\n- scripts/build-sidecar.mjs: now derives GOOS/GOARCH (+ CGO_ENABLED=0) from the Tauri target triple. It previously named the Go binary per-triple but compiled for the host arch — so the x86_64-apple-darwin bundle built on an arm64 runner would embed an arm64 binary.\n- release.yml: added workflow_dispatch (build-only smoke test); version stamping + release publish are now guarded to tag pushes (empty tagName ⇒ tauri-action builds without publishing).\n\nWindows port (cairn's Go code didn't compile on Windows):\n- internal/store: split flock behind build tags — flock_unix.go (unix.Flock) and flock_windows.go (LockFileEx/UnlockFileEx), sharing a portable errWouldBlock; lock.go calls lockExclusiveNB/unlock.\n- internal/check: split the timeout killer — kill_unix.go (Setpgid + SIGKILL to the process group) and kill_windows.go (Process.Kill); check.go calls configureKill.\n- go mod tidy promoted go-toml/v2 to a direct dependency.\n\nVerification: `go build ./...` clean for GOOS=windows/linux/darwin (amd64); `make check` green; the sidecar cross-compiles to correct-arch binaries (Mach-O x86_64, ELF x86-64, PE32+ x86-64 — the Windows case failed before this change).\n\nFollow-ups / notes:\n- Checks still run via `sh -c`, so Windows users need Git Bash or WSL on PATH (already stated in the check package doc and getting-started/installation). The Windows timeout kill is process-only (no group), bounded by WaitDelay — acceptable for v1.\n- Code signing (Apple notarization, Windows cert) is still TODO; only the updater artifacts are signed. Unsigned installers will warn on first launch.\n- To smoke-test the full matrix now: run the \"release\" workflow via workflow_dispatch (no tag needed). To publish: push a tag like v0.1.0. Requires repo secrets TAURI_SIGNING_PRIVATE_KEY (+ password), already referenced.\n- The manual check (a tagged release yields installers for all three OSes) is best closed after the first dispatch/tag run is observed green in CI."}
  - {who: 'human:shaho', at: '2026-06-25T19:15:28Z', did: attested, text: check 2 pass}
  - {who: 'human:shaho', at: '2026-06-25T19:15:32Z', did: ran checks}
  - {who: 'human:shaho', at: '2026-06-25T19:15:32Z', did: transitioned to done}
assignee: agent:claude
active_attempt: att_52784d175e94376a01882ea4
---
The release pipeline (tauri-action + signed updater feed) exists but doesn't actually produce installers for every OS.

## Fixes
- `desktop/src-tauri/tauri.conf.json` — `bundle.targets` is `["dmg","app"]` (macOS only). Widen to cover all platforms so Windows (nsis) and Linux (deb, appimage) actually bundle.
- `scripts/build-sidecar.mjs` — names the Go sidecar with the Tauri target triple but never sets `GOOS`/`GOARCH`, so cross-target builds (notably x86_64 mac on an arm64 runner) embed a wrong-arch binary. Derive `GOOS`/`GOARCH` (+ `CGO_ENABLED=0`) from the triple.
- `.github/workflows/release.yml` — add `workflow_dispatch` (build-only, no publish) so the matrix can be smoke-tested without cutting a tag.

## Acceptance
- `node scripts/build-sidecar.mjs` with `TARGET_TRIPLE=x86_64-apple-darwin` produces an amd64 binary (verified with `file`).
- tauri.conf.json valid; bundle targets cover dmg/app (mac), nsis (win), deb/appimage (linux).