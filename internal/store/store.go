// Package store is cairn's read-fresh file layer (SPEC §8). It parses task files
// losslessly, mutates frontmatter at the yaml.Node level (so unknown keys, ordering,
// and comments survive every write), saves atomically via temp+rename, scans the tasks
// directory, validates the dep graph on load, and serializes all writes behind one
// repository-wide advisory lock.
//
// The split is deliberate: reads decode into the typed task.Task convenience view, but
// writes operate on the raw node — never a struct round-trip, which would drop unknown
// keys and reorder fields.
package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"cairn/internal/config"
	"cairn/internal/task"
)

// ErrNotFound is returned when a task id has no file.
var ErrNotFound = errors.New("task not found")

// ErrConflict is returned when a Doc read by Get is saved after its file changed underneath
// it (another process/actor wrote first). The caller should re-read and retry rather than
// clobber the newer state (optimistic concurrency, SPEC §8).
var ErrConflict = errors.New("task changed since it was read")

// Store is rooted at a repo containing .cairn/. The mutex serializes goroutines sharing
// this instance; Write also takes a repository-wide OS lock for other Cairn processes.
// Reads stay lock-free because atomic rename prevents half-written files.
type Store struct {
	root string
	mu   sync.Mutex
}

// New returns a Store rooted at the given repo directory.
func New(root string) *Store { return &Store{root: root} }

func (s *Store) tasksDir() string   { return filepath.Join(s.root, ".cairn", "tasks") }
func (s *Store) configPath() string { return filepath.Join(s.root, ".cairn", "config.yaml") }
func (s *Store) lockPath() string   { return filepath.Join(s.root, ".cairn", "write.lock") }
func (s *Store) taskPath(id string) string {
	return filepath.Join(s.tasksDir(), id+".md")
}

// Root returns the repo root the store is bound to.
func (s *Store) Root() string { return s.root }

// RunsDir is the gitignored directory for check-run logs (SPEC §1, §6).
func (s *Store) RunsDir() string { return filepath.Join(s.root, ".cairn", "runs") }

// Config loads config.yaml fresh (read-fresh, SPEC §8).
func (s *Store) Config() (config.Config, error) { return config.Load(s.configPath()) }

// Provenance is one append-only audit entry (SPEC §2, §7).
type Provenance struct {
	Who  string `yaml:"who"`
	At   string `yaml:"at"`
	Did  string `yaml:"did"`
	Text string `yaml:"text,omitempty"`
}

// Doc is a parsed task file: the typed view for reads plus the raw frontmatter node and
// body needed for lossless writes. Body is immutable to the engine after create.
type Doc struct {
	Task       task.Task
	Provenance []Provenance
	Body       string // markdown after the frontmatter, preserved byte-for-byte

	node    yaml.Node // document node of the frontmatter
	version string    // content hash of the file at read time; "" for a never-saved Doc
}

// docFields is the read-side decode target. Unknown keys are intentionally absent here —
// they live only in the node and are preserved through it.
type docFields struct {
	ID            string        `yaml:"id"`
	Title         string        `yaml:"title"`
	Status        string        `yaml:"status"`
	Assignee      string        `yaml:"assignee"`
	Deps          []string      `yaml:"deps"`
	Labels        []string      `yaml:"labels"`
	Priority      string        `yaml:"priority"`
	Parent        string        `yaml:"parent"`
	Rank          float64       `yaml:"rank"`
	ActiveAttempt string        `yaml:"active_attempt"`
	Checks        []checkFields `yaml:"checks"`
	Provenance    []Provenance  `yaml:"provenance"`
}

type checkFields struct {
	Desc    string `yaml:"desc"`
	Cmd     string `yaml:"cmd"`
	Type    string `yaml:"type"`
	Result  string `yaml:"result"`
	Cwd     string `yaml:"cwd"`
	Timeout int    `yaml:"timeout"`
}

func (c checkFields) toTask() task.Check {
	return task.Check{Desc: c.Desc, Cmd: c.Cmd, Type: c.Type, Result: c.Result, Cwd: c.Cwd, Timeout: c.Timeout}
}

// Get reads and parses one task file.
func (s *Store) Get(id string) (*Doc, error) {
	b, err := os.ReadFile(s.taskPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", id, err)
	}
	d, err := parse(b)
	if err != nil {
		return nil, err
	}
	d.version = hashBytes(b) // baseline for optimistic-concurrency check on Save
	return d, nil
}

// hashBytes is the content fingerprint used to detect a concurrent write.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ListDocs scans and parses all task files (read-fresh) and validates the dep graph. It
// returns full Docs (including provenance) so callers can derive things like last-activity.
// A dangling dep or a cycle is a loud load failure (SPEC §4).
func (s *Store) ListDocs() ([]*Doc, error) {
	entries, err := os.ReadDir(s.tasksDir())
	if err != nil {
		return nil, fmt.Errorf("store: scan tasks: %w", err)
	}
	var docs []*Doc
	all := make(map[string]task.Task)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.tasksDir(), e.Name()))
		if err != nil {
			return nil, fmt.Errorf("store: read %s: %w", e.Name(), err)
		}
		d, err := parse(b)
		if err != nil {
			return nil, fmt.Errorf("store: parse %s: %w", e.Name(), err)
		}
		docs = append(docs, d)
		all[d.Task.ID] = d.Task
	}
	if err := task.ValidateDeps(all); err != nil {
		return nil, err
	}
	if err := task.ValidateParents(all); err != nil {
		return nil, err
	}
	return docs, nil
}

// List scans the tasks directory (read-fresh) and validates the dep graph, returning the
// typed task views keyed by id.
func (s *Store) List() (map[string]task.Task, error) {
	docs, err := s.ListDocs()
	if err != nil {
		return nil, err
	}
	all := make(map[string]task.Task, len(docs))
	for _, d := range docs {
		all[d.Task.ID] = d.Task
	}
	return all, nil
}

// parse splits frontmatter from body, decodes both the typed view and the raw node.
func parse(b []byte) (*Doc, error) {
	fm, body, err := splitFrontmatter(b)
	if err != nil {
		return nil, err
	}
	d := &Doc{Body: string(body)}
	if err := yaml.Unmarshal(fm, &d.node); err != nil {
		return nil, fmt.Errorf("store: parse frontmatter: %w", err)
	}
	var f docFields
	if err := d.node.Decode(&f); err != nil {
		return nil, fmt.Errorf("store: decode frontmatter: %w", err)
	}
	d.Task = task.Task{ID: f.ID, Title: f.Title, Status: f.Status, Assignee: f.Assignee,
		Deps: f.Deps, Labels: f.Labels, Priority: f.Priority, Parent: f.Parent, Rank: f.Rank,
		ActiveAttempt: f.ActiveAttempt}
	for _, c := range f.Checks {
		d.Task.Checks = append(d.Task.Checks, c.toTask())
	}
	d.Provenance = f.Provenance
	return d, nil
}

// splitFrontmatter separates the leading `---`-fenced YAML from the markdown body.
func splitFrontmatter(b []byte) (fm, body []byte, err error) {
	if !bytes.HasPrefix(b, []byte("---\n")) {
		return nil, nil, errors.New("store: missing frontmatter '---' fence")
	}
	rest := b[len("---\n"):]
	if i := bytes.Index(rest, []byte("\n---\n")); i >= 0 {
		return rest[:i+1], rest[i+len("\n---\n"):], nil
	}
	if bytes.HasSuffix(rest, []byte("\n---")) {
		return rest[:len(rest)-len("---")], nil, nil
	}
	return nil, nil, errors.New("store: missing closing '---' fence")
}

// mapping returns the frontmatter's top-level mapping node.
func (d *Doc) mapping() *yaml.Node {
	if d.node.Kind == yaml.DocumentNode && len(d.node.Content) > 0 {
		return d.node.Content[0]
	}
	return &d.node
}

// SetStatus surgically updates the status value node (creating the key if absent) and
// the typed mirror. Engine-owned per SPEC §2.
func (d *Doc) SetStatus(status string) error {
	setScalar(d.mapping(), "status", strNode(status))
	d.Task.Status = status
	return nil
}

// SetPriority/SetParent/SetLabels surgically update the optional organization fields,
// removing the key entirely when cleared so frontmatter stays clean.
func (d *Doc) SetPriority(p string) error {
	if p == "" {
		removeKey(d.mapping(), "priority")
	} else {
		setScalar(d.mapping(), "priority", strNode(p))
	}
	d.Task.Priority = p
	return nil
}

func (d *Doc) SetParent(parent string) error {
	if parent == "" {
		removeKey(d.mapping(), "parent")
	} else {
		setScalar(d.mapping(), "parent", strNode(parent))
	}
	d.Task.Parent = parent
	return nil
}

func (d *Doc) SetLabels(labels []string) error {
	if len(labels) == 0 {
		removeKey(d.mapping(), "labels")
	} else {
		setScalar(d.mapping(), "labels", strSeq(labels))
	}
	d.Task.Labels = labels
	return nil
}

// SetRank sets the manual board ordering value (0 clears it).
func (d *Doc) SetRank(rank float64) error {
	if rank == 0 {
		removeKey(d.mapping(), "rank")
	} else {
		setScalar(d.mapping(), "rank", floatNode(rank))
	}
	d.Task.Rank = rank
	return nil
}

// SetAssignee sets the assignee (used by claim, SPEC §7).
func (d *Doc) SetAssignee(who string) error {
	if who == "" {
		removeKey(d.mapping(), "assignee")
	} else {
		setScalar(d.mapping(), "assignee", strNode(who))
	}
	d.Task.Assignee = who
	return nil
}

// SetActiveAttempt records the session attempt currently eligible for review.
func (d *Doc) SetActiveAttempt(id string) error {
	if id == "" {
		removeKey(d.mapping(), "active_attempt")
	} else {
		setScalar(d.mapping(), "active_attempt", strNode(id))
	}
	d.Task.ActiveAttempt = id
	return nil
}

// SetCheckResult writes the result of the check at index i (engine-managed, SPEC §6).
func (d *Doc) SetCheckResult(i int, result string) error {
	checks, ok := mapGet(d.mapping(), "checks")
	if !ok || checks.Kind != yaml.SequenceNode || i < 0 || i >= len(checks.Content) {
		return fmt.Errorf("store: check index %d out of range", i)
	}
	setScalar(checks.Content[i], "result", strNode(result))
	if i < len(d.Task.Checks) {
		d.Task.Checks[i].Result = result
	}
	return nil
}

// AppendProvenance appends one audit entry to the provenance sequence, creating it if
// absent (SPEC §7: every write appends one). Entries are flow-style for clean diffs.
func (d *Doc) AppendProvenance(who, did, text string, at time.Time) error {
	m := d.mapping()
	seq, ok := mapGet(m, "provenance")
	if !ok {
		seq = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		m.Content = append(m.Content, strNode("provenance"), seq)
	}
	entry := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Style: yaml.FlowStyle,
		Content: []*yaml.Node{strNode("who"), strNode(who), strNode("at"), tsNode(at), strNode("did"), strNode(did)}}
	if text != "" {
		entry.Content = append(entry.Content, strNode("text"), strNode(text))
	}
	seq.Content = append(seq.Content, entry)
	d.Provenance = append(d.Provenance, Provenance{Who: who, At: at.UTC().Format(time.RFC3339), Did: did, Text: text})
	return nil
}

// Save renders the Doc and writes it atomically under the repository write lock (SPEC §8).
func (s *Store) Save(d *Doc) error {
	return s.Write(context.Background(), "", "save task", func(tx *WriteTx) error {
		return tx.SaveTask(d)
	})
}

// save renders frontmatter from the node and recomposes the file. Callers must hold the
// repository write transaction.
// A Doc carrying a version (i.e. read via Get) must still match the on-disk file, else a
// concurrent writer changed it and we refuse with ErrConflict rather than clobber.
func (s *Store) save(d *Doc) error {
	if d.version != "" {
		if err := s.checkVersion(d); err != nil {
			return err
		}
	}

	var fm bytes.Buffer
	enc := yaml.NewEncoder(&fm)
	enc.SetIndent(2) // match the 2-space convention so diffs stay clean
	if err := enc.Encode(d.mapping()); err != nil {
		return fmt.Errorf("store: encode frontmatter: %w", err)
	}
	enc.Close()

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(fm.Bytes())
	out.WriteString("---\n")
	out.WriteString(d.Body)
	data := out.Bytes()
	if err := atomicWrite(s.taskPath(d.Task.ID), data); err != nil {
		return err
	}
	d.version = hashBytes(data) // advance so the same Doc can be saved again
	return nil
}

// checkVersion compares the on-disk file against the Doc's read-time fingerprint.
func (s *Store) checkVersion(d *Doc) error {
	cur, err := os.ReadFile(s.taskPath(d.Task.ID))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s was deleted", ErrConflict, d.Task.ID)
	}
	if err != nil {
		return fmt.Errorf("store: reread %s: %w", d.Task.ID, err)
	}
	if hashBytes(cur) != d.version {
		return fmt.Errorf("%w: %s", ErrConflict, d.Task.ID)
	}
	return nil
}

// Create mints the next id under the repository lock, writes a new task file in the initial state
// with a `created` provenance entry, and persists the advanced counter (SPEC §3, §7).
// Draft is the caller-supplied content for a new task. Id and status are engine-assigned.
type Draft struct {
	Title    string
	Body     string
	Deps     []string
	Checks   []task.Check
	Labels   []string
	Priority string
	Parent   string
	Rank     float64
}

func (s *Store) Create(draft Draft, actor string, at time.Time) (*Doc, error) {
	var created *Doc
	err := s.Write(context.Background(), actor, "create task", func(tx *WriteTx) error {
		cfg, err := config.Load(s.configPath())
		if err != nil {
			return err
		}
		id, next := cfg.NewID()

		d := &Doc{Body: draft.Body}
		d.node = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
		m := d.mapping()
		m.Content = append(m.Content, strNode("id"), strNode(id))
		m.Content = append(m.Content, strNode("title"), strNode(draft.Title))
		m.Content = append(m.Content, strNode("status"), strNode(cfg.Initial))
		if draft.Priority != "" {
			m.Content = append(m.Content, strNode("priority"), strNode(draft.Priority))
		}
		if len(draft.Labels) > 0 {
			m.Content = append(m.Content, strNode("labels"), strSeq(draft.Labels))
		}
		if draft.Parent != "" {
			m.Content = append(m.Content, strNode("parent"), strNode(draft.Parent))
		}
		if draft.Rank != 0 {
			m.Content = append(m.Content, strNode("rank"), floatNode(draft.Rank))
		}
		if len(draft.Deps) > 0 {
			m.Content = append(m.Content, strNode("deps"), strSeq(draft.Deps))
		}
		if len(draft.Checks) > 0 {
			m.Content = append(m.Content, strNode("checks"), checksNode(draft.Checks))
		}
		d.Task = task.Task{ID: id, Title: draft.Title, Status: cfg.Initial, Deps: draft.Deps,
			Checks: draft.Checks, Labels: draft.Labels, Priority: draft.Priority, Parent: draft.Parent,
			Rank: draft.Rank}

		if err := d.AppendProvenance(actor, "created", "", at); err != nil {
			return err
		}
		if err := tx.SaveTask(d); err != nil {
			return err
		}
		if err := config.Save(s.configPath(), next); err != nil {
			return err
		}
		created = d
		return nil
	})
	return created, err
}

// --- yaml.Node helpers ---

func strNode(s string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s} }
func intNode(i int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(i)}
}
func int64Node(i int64) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(i, 10)}
}
func floatNode(f float64) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(f, 'f', -1, 64)}
}
func tsNode(t time.Time) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!timestamp", Value: t.UTC().Format(time.RFC3339)}
}

func strSeq(items []string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, it := range items {
		n.Content = append(n.Content, strNode(it))
	}
	return n
}

func checksNode(checks []task.Check) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, c := range checks {
		m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		m.Content = append(m.Content, strNode("desc"), strNode(c.Desc))
		if c.Cmd != "" {
			m.Content = append(m.Content, strNode("cmd"), strNode(c.Cmd))
		}
		if c.Type != "" {
			m.Content = append(m.Content, strNode("type"), strNode(c.Type))
		}
		if c.Cwd != "" {
			m.Content = append(m.Content, strNode("cwd"), strNode(c.Cwd))
		}
		if c.Timeout != 0 {
			m.Content = append(m.Content, strNode("timeout"), intNode(c.Timeout))
		}
		result := c.Result
		if result == "" {
			result = "pending"
		}
		m.Content = append(m.Content, strNode("result"), strNode(result))
		seq.Content = append(seq.Content, m)
	}
	return seq
}

// mapGet returns the value node for key in a mapping node.
func mapGet(m *yaml.Node, key string) (*yaml.Node, bool) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1], true
		}
	}
	return nil, false
}

// removeKey deletes a key and its value node from a mapping, if present.
func removeKey(m *yaml.Node, key string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

// setScalar replaces the value node for key, or appends key+val if the key is absent.
func setScalar(m *yaml.Node, key string, val *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = val
			return
		}
	}
	m.Content = append(m.Content, strNode(key), val)
}

// atomicWrite writes data to a temp file in the same dir then renames over the target,
// so a concurrent reader never sees a partial file (SPEC §8).
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("store: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("store: close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("store: rename: %w", err)
	}
	return nil
}
