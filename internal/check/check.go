// Package check executes a task's checks as shell commands and records the outcome
// (SPEC §6). It is a leaf package: it knows nothing about tasks, config, or the store.
// Callers resolve a check into a Spec (command, cwd, timeout), Run it, and write the
// Result back. Manual checks (no command) are never passed here.
//
// Contract (SPEC §6): every command runs via `sh -c`, so POSIX shell is required;
// native Windows users use WSL/Git Bash. Exit code 0 is pass, non-zero is fail, a
// timeout is a (killed) fail. Combined stdout+stderr is tailed to a log under LogDir;
// the task file keeps only the result.
package check

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// TailBytes is how much of the combined output is retained — the trailing ~8KB
// (SPEC §6). Older output is discarded as it streams, bounding memory on noisy checks.
const TailBytes = 8 << 10

// Spec is a single resolved check to execute. Cmd is a full shell command line. Cwd is
// relative to the runner's Root ("" or "." = root). Timeout <= 0 means no deadline; the
// caller is expected to have already applied config's check_timeout_default.
type Spec struct {
	Cmd     string
	Cwd     string
	Timeout time.Duration
}

// Result is the outcome of running a Spec. Pass is the only value the task file stores
// (as pass/fail); the rest aids diagnostics and the run log.
type Result struct {
	Pass     bool
	ExitCode int
	TimedOut bool
	Reason   string // non-empty on failure: exit status, timeout, or start error
	Output   string // trailing TailBytes of combined stdout+stderr
	LogPath  string // path of the written run log
	Duration time.Duration
}

// Runner executes checks rooted at a repo and writes logs under LogDir
// (e.g. <root>/.cairn/runs). Now is an injectable clock for log filenames; nil uses
// the wall clock. The zero value is usable when Root/LogDir default to the process cwd.
type Runner struct {
	Root   string
	LogDir string
	Now    func() time.Time
}

func (r Runner) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

// Run executes spec via `sh -c`, captures the tail of its output, writes a run log, and
// reports the result. A non-zero exit or a timeout is a failed Result, not an error; the
// error return is reserved for infrastructure faults (e.g. the log could not be written).
func (r Runner) Run(id string, spec Spec) (Result, error) {
	if spec.Cmd == "" {
		return Result{}, errors.New("check: empty command (manual checks are not run by the runner)")
	}

	dir := r.Root
	if spec.Cwd != "" && spec.Cwd != "." {
		dir = filepath.Join(r.Root, spec.Cwd)
	}

	ctx := context.Background()
	if spec.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, spec.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", spec.Cmd)
	cmd.Dir = dir
	// Run the command in its own process group so a timeout kills children too, not
	// just the sh that spawned them.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second

	out := &tailBuffer{max: TailBytes}
	cmd.Stdout = out
	cmd.Stderr = out

	start := r.now()
	runErr := cmd.Run()
	res := Result{Output: out.String(), Duration: r.now().Sub(start)}

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		res.TimedOut = true
		res.ExitCode = -1
		res.Reason = fmt.Sprintf("timed out after %s", spec.Timeout)
	case runErr == nil:
		res.Pass = true
	default:
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			res.ExitCode = ee.ExitCode()
			res.Reason = fmt.Sprintf("exit status %d", res.ExitCode)
		} else {
			res.ExitCode = -1
			res.Reason = "failed to run: " + runErr.Error()
		}
	}

	logPath, err := r.writeLog(id, spec, dir, res)
	if err != nil {
		return res, err
	}
	res.LogPath = logPath
	return res, nil
}

func (r Runner) writeLog(id string, spec Spec, dir string, res Result) (string, error) {
	if err := os.MkdirAll(r.LogDir, 0o755); err != nil {
		return "", fmt.Errorf("check: create log dir: %w", err)
	}
	stamp := r.now().UTC().Format("20060102-150405.000")
	path := filepath.Join(r.LogDir, fmt.Sprintf("%s-%s.log", id, stamp))
	header := fmt.Sprintf("cmd: %s\ncwd: %s\nexit: %d  timedout: %t  duration: %s\n----\n",
		spec.Cmd, dir, res.ExitCode, res.TimedOut, res.Duration)
	if err := os.WriteFile(path, []byte(header+res.Output), 0o644); err != nil {
		return "", fmt.Errorf("check: write log: %w", err)
	}
	return path, nil
}

// tailBuffer is an io.Writer that retains only the trailing max bytes written to it.
// It is safe for the concurrent Stdout/Stderr writers os/exec spawns.
type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
	max int
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if len(b.buf) > b.max {
		b.buf = b.buf[len(b.buf)-b.max:]
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}
