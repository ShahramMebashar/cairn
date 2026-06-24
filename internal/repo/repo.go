// Package repo handles initializing a cairn workspace in a project: scaffolding the
// .cairn/ tree, writing a default config, ensuring the runs dir is gitignored, dropping a
// generic .cairn/WORKFLOW.md, and embedding a cairn agent-loop block in AGENTS.md / CLAUDE.md.
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

var gitignoreEntries = []string{
	".cairn/runs/",
	".cairn/live/",
	".cairn/write.lock",
}

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
		filepath.Join(cairnDir(root), "sessions"),
		filepath.Join(cairnDir(root), "live"),
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
	// Embed the agent-loop block in the agent docs so any tool sees the loop in-context
	// (harnesses auto-load these files but don't follow the WORKFLOW.md link).
	for _, doc := range []string{"AGENTS.md", "CLAUDE.md"} {
		if err := ensureAgentLoop(root, doc); err != nil {
			return err
		}
	}
	return nil
}

// EnsureSessionDirs creates just the session and live directories and ensures the
// ephemeral paths stay gitignored. It is the targeted counterpart to Init for the session
// hot path: a begin call must not rewrite config.yaml or the agent docs, but live/ must
// still never be committed.
func EnsureSessionDirs(root string) error {
	for _, dir := range []string{
		filepath.Join(cairnDir(root), "sessions"),
		filepath.Join(cairnDir(root), "live"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("repo: create %s: %w", dir, err)
		}
	}
	return ensureGitignore(root)
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

// Markers wrap the cairn-managed block in an agent doc so re-init can refresh it in place
// without disturbing anything a human wrote around it. They are HTML comments, invisible in
// rendered Markdown.
const (
	agentLoopStart = "<!-- cairn:agent-loop:start -->"
	agentLoopEnd   = "<!-- cairn:agent-loop:end -->"
)

// agentLoopBlock is the concise, self-contained onboarding block embedded in AGENTS.md /
// CLAUDE.md. It lives inline (not just a link) because agent harnesses auto-load these docs
// but do NOT follow links to .cairn/WORKFLOW.md — so the required loop must be in the file
// itself. Kept short on purpose: it lands in every agent's context. WORKFLOW.md holds depth.
const agentLoopBlock = "## Agent loop — required\n\n" +
	"All work in this repo is tracked in **cairn** (the task graph under `.cairn/`). Drive every\n" +
	"non-trivial change through a task using cairn's MCP tools — don't edit task files by hand:\n\n" +
	"1. **identity** — confirm your bound actor.\n" +
	"2. **find work** — list ready tasks in the initial state.\n" +
	"3. **begin** — claim a task and open a session (`expected_actor` + a unique `idempotency_key`).\n" +
	"4. **build + heartbeat** — make the change; report concise progress.\n" +
	"5. **note** — add a short provenance note at each meaningful decision.\n" +
	"6. **run_checks** — run the task's checks before handoff.\n" +
	"7. **finish** — end the session into review with a summary.\n" +
	"8. **close** — transition to a closed state once reviewed (re-runs checks).\n\n" +
	"Full lifecycle, gates, and note discipline: [" + workflowPath + "](" + workflowPath + ")."

// ensureAgentLoop embeds (or refreshes) the cairn agent-loop block in an agent doc. If the
// markers are present it replaces just the content between them; otherwise it appends the
// block to an existing file, or creates the file with a header. Content outside the markers
// is never touched, so a project can write its own guidance alongside.
func ensureAgentLoop(root, name string) error {
	path := filepath.Join(root, name)
	body := agentLoopStart + "\n" + agentLoopBlock + "\n" + agentLoopEnd

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("repo: read %s: %w", name, err)
	}
	if os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte("# "+name+"\n\n"+body+"\n"), 0o644); err != nil {
			return fmt.Errorf("repo: write %s: %w", name, err)
		}
		return nil
	}

	content := string(existing)
	if i := strings.Index(content, agentLoopStart); i >= 0 {
		if j := strings.Index(content, agentLoopEnd); j >= i {
			content = content[:i] + body + content[j+len(agentLoopEnd):]
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("repo: update %s: %w", name, err)
			}
			return nil
		}
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("repo: update %s: %w", name, err)
	}
	return nil
}

// ensureGitignore appends Cairn's ephemeral paths when they are not already present.
func ensureGitignore(root string) error {
	path := filepath.Join(root, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("repo: read .gitignore: %w", err)
	}
	content := string(existing)
	existingEntries := make(map[string]bool)
	for line := range strings.SplitSeq(content, "\n") {
		existingEntries[strings.TrimSpace(line)] = true
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	for _, entry := range gitignoreEntries {
		if !existingEntries[entry] {
			content += entry + "\n"
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("repo: write .gitignore: %w", err)
	}
	return nil
}
