// Package repo handles initializing a cairn workspace in a project: scaffolding the
// .cairn/ tree, writing a default config, ensuring the runs dir is gitignored, dropping a
// generic .cairn/WORKFLOW.md, and pointing AGENTS.md / CLAUDE.md at that workflow.
//
// It is the single source of init logic shared by every front-end — the `cairn init` CLI,
// the auto-init on `cairn serve`, and the web `POST /api/init` endpoint all call Init, so
// they can't drift (SPEC §0: one rule-set, thin adapters).
package repo

import (
	_ "embed"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"cairn/internal/config"
)

const gitignoreEntry = ".cairn/runs/"

// workflowPath is the repo-relative location of the generic task workflow, referenced from
// the agent docs.
const workflowPath = ".cairn/WORKFLOW.md"

//go:embed workflow.md
var workflowTemplate string

func cairnDir(root string) string   { return filepath.Join(root, ".cairn") }
func configPath(root string) string { return filepath.Join(cairnDir(root), "config.yaml") }

// IsInitialized reports whether root already has a cairn workspace.
func IsInitialized(root string) bool {
	_, err := os.Stat(configPath(root))
	return err == nil
}

// CurrentPrefix returns the configured id prefix, or "" if the repo is not initialized
// or its config can't be read.
func CurrentPrefix(root string) string {
	cfg, err := config.Load(configPath(root))
	if err != nil {
		return ""
	}
	return cfg.Prefix
}

// DerivePrefix suggests an id prefix from a project directory name: uppercased, stripped
// to [A-Z0-9], falling back to TASK when nothing usable remains.
func DerivePrefix(root string) string {
	base := filepath.Base(root)
	if abs, err := filepath.Abs(root); err == nil {
		base = filepath.Base(abs)
	}
	var b strings.Builder
	for _, r := range strings.ToUpper(base) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "TASK"
	}
	if out[0] >= '0' && out[0] <= '9' {
		return "TASK" + out // ids must start with a letter
	}
	return out
}

// DeriveActor suggests a human identity from the OS login name, e.g. "human:shaho". It is a
// default only — the web UI lets each person override it. Falls back to "human:dev".
func DeriveActor() string {
	name := ""
	if u, err := user.Current(); err == nil {
		name = u.Username
	}
	// On Windows the username can be "DOMAIN\\user"; keep the last segment.
	if i := strings.LastIndexAny(name, `\/`); i >= 0 {
		name = name[i+1:]
	}
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "human:dev"
	}
	return "human:" + out
}

// Init scaffolds a cairn workspace at root. It is idempotent: an existing config.yaml is
// left untouched (counter and edits preserved), and missing pieces (dirs, .gitignore
// entry) are filled in. An empty prefix is derived from the directory name.
func Init(root, prefix string) error {
	if prefix == "" {
		prefix = DerivePrefix(root)
	}
	for _, dir := range []string{
		filepath.Join(cairnDir(root), "tasks"),
		filepath.Join(cairnDir(root), "runs"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("repo: create %s: %w", dir, err)
		}
	}
	if !IsInitialized(root) {
		if err := config.Save(configPath(root), config.Default(prefix)); err != nil {
			return err
		}
	}
	if err := ensureGitignore(root); err != nil {
		return err
	}
	if err := ensureWorkflow(root); err != nil {
		return err
	}
	// Point the agent docs at the workflow so any tool/human starts there.
	for _, doc := range []string{"AGENTS.md", "CLAUDE.md"} {
		if err := ensureWorkflowRef(root, doc); err != nil {
			return err
		}
	}
	return nil
}

// ensureWorkflow writes the generic task workflow if absent. It never overwrites an existing
// WORKFLOW.md, so a project can customize it.
func ensureWorkflow(root string) error {
	path := filepath.Join(cairnDir(root), "WORKFLOW.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.WriteFile(path, []byte(workflowTemplate), 0o644); err != nil {
		return fmt.Errorf("repo: write WORKFLOW.md: %w", err)
	}
	return nil
}

// ensureWorkflowRef makes an agent doc (AGENTS.md / CLAUDE.md) reference the workflow:
// it appends a section if the file exists without a reference, or creates a stub if absent.
// Idempotent — a file that already mentions the workflow is left untouched.
func ensureWorkflowRef(root, name string) error {
	path := filepath.Join(root, name)
	section := "## Task workflow\n\nThis repo tracks work with **cairn**. See [" +
		workflowPath + "](" + workflowPath + ") for the task lifecycle, the agent loop, and " +
		"note discipline.\n"

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("repo: read %s: %w", name, err)
	}
	if err == nil {
		if strings.Contains(string(existing), workflowPath) {
			return nil
		}
		content := string(existing)
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + section
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("repo: update %s: %w", name, err)
		}
		return nil
	}
	if err := os.WriteFile(path, []byte("# "+name+"\n\n"+section), 0o644); err != nil {
		return fmt.Errorf("repo: write %s: %w", name, err)
	}
	return nil
}

// ensureGitignore appends the runs-dir entry to .gitignore if not already present.
func ensureGitignore(root string) error {
	path := filepath.Join(root, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("repo: read .gitignore: %w", err)
	}
	content := string(existing)
	for line := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(line) == gitignoreEntry {
			return nil
		}
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += gitignoreEntry + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("repo: write .gitignore: %w", err)
	}
	return nil
}
