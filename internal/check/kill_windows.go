//go:build windows

package check

import "os/exec"

// configureKill arranges cancellation on Windows, which has no POSIX process groups. Killing
// the process directly is enough for cairn's checks; any stragglers are bounded by WaitDelay.
// (Checks still run via `sh -c`, so Windows users need Git Bash or WSL on PATH — see the
// package doc.)
func configureKill(cmd *exec.Cmd) {
	cmd.Cancel = func() error { return cmd.Process.Kill() }
}
