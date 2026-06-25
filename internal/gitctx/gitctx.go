// Package gitctx derives review evidence from a local Git checkout.
package gitctx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	shortTimeout = 2 * time.Second
	longTimeout  = 5 * time.Second
)

var ErrUnavailable = errors.New("git context unavailable")

// Ref is the current Git position of a repository.
type Ref struct {
	Branch string `json:"branch,omitempty"`
	Head   string `json:"head,omitempty"`
}

// ChangedFile is one path reported by Git.
type ChangedFile struct {
	Status  string `json:"status"`
	Path    string `json:"path"`
	OldPath string `json:"oldPath,omitempty"`
}

// Commit is one commit in a session range.
type Commit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

// Warning is an actionable review caveat derived from Git evidence.
type Warning struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Context is the Git evidence for one session.
type Context struct {
	Available    bool          `json:"available"`
	Error        string        `json:"error,omitempty"`
	Branch       string        `json:"branch,omitempty"`
	HeadStarted  string        `json:"headStarted,omitempty"`
	HeadFinished string        `json:"headFinished,omitempty"`
	CurrentHead  string        `json:"currentHead,omitempty"`
	FilesChanged []ChangedFile `json:"filesChanged,omitempty"`
	Uncommitted  []ChangedFile `json:"uncommitted,omitempty"`
	Commits      []Commit      `json:"commits,omitempty"`
	Dirty        bool          `json:"dirty"`
	Warnings     []Warning     `json:"warnings,omitempty"`
}

// Current returns the current branch and HEAD.
func Current(ctx context.Context, repo string) (Ref, error) {
	branch, branchErr := gitOutput(ctx, shortTimeout, repo, "rev-parse", "--abbrev-ref", "HEAD")
	head, headErr := gitOutput(ctx, shortTimeout, repo, "rev-parse", "--verify", "HEAD")
	if branchErr != nil && headErr != nil {
		return Ref{}, fmt.Errorf("%w: %v", ErrUnavailable, headErr)
	}
	ref := Ref{Branch: strings.TrimSpace(branch), Head: strings.TrimSpace(head)}
	if ref.Branch == "HEAD" {
		ref.Branch = ""
	}
	return ref, nil
}

// Session derives code context for a session from durable Git anchors.
func Session(ctx context.Context, repo string, started, finished, branch, latestCheckHead string, active bool) Context {
	out := Context{
		Available:    true,
		Branch:       branch,
		HeadStarted:  started,
		HeadFinished: finished,
	}
	ref, err := Current(ctx, repo)
	if err != nil {
		return unavailable(err)
	}
	out.CurrentHead = ref.Head
	if out.Branch == "" {
		out.Branch = ref.Branch
	}

	status, err := Status(ctx, repo)
	if err != nil {
		out.Warnings = append(out.Warnings, Warning{Kind: "status_unavailable", Message: "Working tree status could not be read."})
	} else {
		out.Uncommitted = status
		out.Dirty = len(status) > 0
	}

	switch {
	case started == "":
		out.Warnings = append(out.Warnings, Warning{Kind: "missing_start", Message: "Session has no starting Git commit."})
	case active:
		// Three-dot shows the session's own changes since it diverged from `started`,
		// ignoring work that landed on the start side meanwhile.
		out.FilesChanged = diffOrWarn(ctx, repo, started+"...HEAD", &out)
	case finished != "":
		// Diff uses three-dot for the same session-changes semantics as the active
		// case; log uses two-dot to list exactly the commits this session added.
		out.FilesChanged = diffOrWarn(ctx, repo, started+"..."+finished, &out)
		out.Commits = logOrWarn(ctx, repo, started+".."+finished, &out)
	default:
		out.Warnings = append(out.Warnings, Warning{Kind: "missing_finish", Message: "Finished session has no ending Git commit."})
	}

	if !active && out.Dirty {
		out.Warnings = append(out.Warnings, Warning{Kind: "dirty_finish", Message: "Repository has uncommitted changes while this session is finished."})
	}
	if !active && started != "" && finished != "" && len(out.Commits) == 0 && len(out.FilesChanged) == 0 {
		out.Warnings = append(out.Warnings, Warning{Kind: "no_changes", Message: "No commits or file changes were found for this session range."})
	}
	if latestCheckHead != "" && finished != "" && latestCheckHead != finished {
		out.Warnings = append(out.Warnings, Warning{Kind: "stale_checks", Message: "Latest checks ran on a different commit than the session finish commit."})
	}
	return out
}

// Diff returns name-status entries for a revision range.
func Diff(ctx context.Context, repo, revRange string) ([]ChangedFile, error) {
	raw, err := gitBytes(ctx, longTimeout, repo, "diff", "--name-status", "-z", revRange, "--")
	if err != nil {
		return nil, err
	}
	return parseNameStatus(raw), nil
}

// Status returns porcelain status entries for uncommitted changes.
func Status(ctx context.Context, repo string) ([]ChangedFile, error) {
	raw, err := gitBytes(ctx, shortTimeout, repo, "status", "--porcelain=v1", "-z")
	if err != nil {
		return nil, err
	}
	return parseStatus(raw), nil
}

func parseStatus(raw []byte) []ChangedFile {
	parts := splitNUL(raw)
	files := make([]ChangedFile, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if len(p) < 4 {
			continue
		}
		status := strings.TrimSpace(p[:2])
		file := ChangedFile{Status: status, Path: p[3:]}
		// porcelain -z orders renames NEW\0OLD (the opposite of diff --name-status),
		// so the path on the status line is the destination and OldPath follows.
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			if i+1 < len(parts) {
				file.OldPath = parts[i+1]
				i++
			}
		}
		files = append(files, file)
	}
	return files
}

// Log returns commits in newest-first order for a revision range.
func Log(ctx context.Context, repo, revRange string) ([]Commit, error) {
	raw, err := gitBytes(ctx, longTimeout, repo, "log", "--format=%H%x00%s%x00", revRange, "--")
	if err != nil {
		return nil, err
	}
	parts := splitNUL(raw)
	commits := make([]Commit, 0, len(parts)/2)
	for i := 0; i+1 < len(parts); i += 2 {
		commits = append(commits, Commit{Hash: parts[i], Subject: parts[i+1]})
	}
	return commits, nil
}

func diffOrWarn(ctx context.Context, repo, revRange string, out *Context) []ChangedFile {
	files, err := Diff(ctx, repo, revRange)
	if err != nil {
		out.Warnings = append(out.Warnings, Warning{Kind: "diff_unavailable", Message: "Changed files could not be derived for the session range."})
		return nil
	}
	return files
}

func logOrWarn(ctx context.Context, repo, revRange string, out *Context) []Commit {
	commits, err := Log(ctx, repo, revRange)
	if err != nil {
		out.Warnings = append(out.Warnings, Warning{Kind: "log_unavailable", Message: "Commit history could not be derived for the session range."})
		return nil
	}
	return commits
}

func unavailable(err error) Context {
	return Context{Available: false, Error: err.Error(), Warnings: []Warning{{
		Kind:    "git_unavailable",
		Message: "Git context is unavailable for this repository.",
	}}}
}

func gitOutput(ctx context.Context, timeout time.Duration, repo string, args ...string) (string, error) {
	b, err := gitBytes(ctx, timeout, repo, args...)
	return string(b), err
}

// gitBytes runs git directly (never through a shell), so caller-supplied refs cannot
// inject commands. It bounds the call with timeout and returns ErrUnavailable on failure.
func gitBytes(ctx context.Context, timeout time.Duration, repo string, args ...string) ([]byte, error) {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}
	var cancel context.CancelFunc
	runCtx, cancel = context.WithTimeout(runCtx, timeout)
	defer cancel()

	fullArgs := append([]string{"-C", repo}, args...)
	cmd := exec.CommandContext(runCtx, "git", fullArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%w: git %s: %s", ErrUnavailable, strings.Join(args, " "), msg)
	}
	return out, nil
}

func parseNameStatus(raw []byte) []ChangedFile {
	parts := splitNUL(raw)
	files := make([]ChangedFile, 0, len(parts)/2)
	for i := 0; i < len(parts); i++ {
		status := parts[i]
		if status == "" || i+1 >= len(parts) {
			continue
		}
		file := ChangedFile{Status: status, Path: parts[i+1]}
		i++
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			if i+1 < len(parts) {
				file.OldPath = file.Path
				file.Path = parts[i+1]
				i++
			}
		}
		files = append(files, file)
	}
	return files
}

func splitNUL(raw []byte) []string {
	raw = bytes.TrimRight(raw, "\x00")
	if len(raw) == 0 {
		return nil
	}
	chunks := bytes.Split(raw, []byte{0})
	out := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		out = append(out, string(chunk))
	}
	return out
}
