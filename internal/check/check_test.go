package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runner(t *testing.T) Runner {
	t.Helper()
	root := t.TempDir()
	return Runner{Root: root, LogDir: filepath.Join(root, ".cairn", "runs")}
}

func TestRunPass(t *testing.T) {
	r := runner(t)
	res, err := r.Run("PROJ-001", Spec{Cmd: "exit 0"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Pass || res.ExitCode != 0 || res.TimedOut {
		t.Fatalf("got %+v, want pass exit 0", res)
	}
	if res.LogPath == "" {
		t.Fatal("expected a log path")
	}
	if _, err := os.Stat(res.LogPath); err != nil {
		t.Fatalf("log not written: %v", err)
	}
}

func TestRunFailExitCode(t *testing.T) {
	r := runner(t)
	res, err := r.Run("PROJ-001", Spec{Cmd: "exit 7"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Pass || res.ExitCode != 7 {
		t.Fatalf("got %+v, want fail exit 7", res)
	}
	if res.Reason == "" {
		t.Fatal("expected a reason on failure")
	}
}

func TestRunCapturesStdoutAndStderr(t *testing.T) {
	r := runner(t)
	res, err := r.Run("PROJ-001", Spec{Cmd: "echo out; echo err 1>&2"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(res.Output, "out") || !strings.Contains(res.Output, "err") {
		t.Fatalf("output %q missing stdout/stderr", res.Output)
	}
	// The log file on disk contains the captured output.
	b, err := os.ReadFile(res.LogPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(b), "out") {
		t.Fatalf("log %q missing output", b)
	}
}

func TestRunCwdRelativeToRoot(t *testing.T) {
	r := runner(t)
	if err := os.MkdirAll(filepath.Join(r.Root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := r.Run("PROJ-001", Spec{Cmd: "pwd", Cwd: "sub"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(strings.TrimSpace(res.Output), "sub") {
		t.Fatalf("pwd %q did not run in sub", res.Output)
	}
}

func TestRunTimeoutKills(t *testing.T) {
	r := runner(t)
	start := time.Now()
	res, err := r.Run("PROJ-001", Spec{Cmd: "sleep 5", Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.TimedOut || res.Pass {
		t.Fatalf("got %+v, want timed out", res)
	}
	if !strings.Contains(strings.ToLower(res.Reason), "tim") {
		t.Fatalf("reason %q does not mention timeout", res.Reason)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("timeout took %s, process not killed promptly", elapsed)
	}
}

func TestRunContextCancellationKills(t *testing.T) {
	r := runner(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := r.RunContext(ctx, "PROJ-001", Spec{Cmd: "sleep 5"})
	if err != nil {
		t.Fatalf("RunContext: %v", err)
	}
	if res.Pass {
		t.Fatalf("got %+v, want canceled failure", res)
	}
}

func TestRunTailTruncatesLog(t *testing.T) {
	r := runner(t)
	res, err := r.Run("PROJ-001", Spec{Cmd: "for i in $(seq 1 20000); do echo line$i; done"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Output) > TailBytes {
		t.Fatalf("output kept %d bytes, want <= %d", len(res.Output), TailBytes)
	}
}

func TestRunEmptyCmdIsError(t *testing.T) {
	r := runner(t)
	if _, err := r.Run("PROJ-001", Spec{Cmd: ""}); err == nil {
		t.Fatal("expected error for empty cmd (manual checks are not run here)")
	}
}
