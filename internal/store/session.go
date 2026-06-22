package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"cairn/internal/session"
)

var (
	// ErrSessionNotFound is returned when a session id has no durable file.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionConflict is returned when a stale session document would overwrite a newer one.
	ErrSessionConflict = errors.New("session changed since it was read")
	// ErrLiveSession is returned when a task already has a live session.
	ErrLiveSession = errors.New("task already has a live session")
)

// SessionDoc is a typed session plus its lossless YAML representation.
type SessionDoc struct {
	Session session.Session

	node    yaml.Node
	version string
}

func (s *Store) sessionsDir() string { return filepath.Join(s.root, ".cairn", "sessions") }
func (s *Store) liveDir() string     { return filepath.Join(s.root, ".cairn", "live") }
func (s *Store) sessionPath(id string) string {
	return filepath.Join(s.sessionsDir(), id+".yaml")
}
func (s *Store) livePath(id string) string {
	return filepath.Join(s.liveDir(), id+".json")
}

// GetSession reads one durable session.
func (s *Store) GetSession(id string) (*SessionDoc, error) {
	b, err := os.ReadFile(s.sessionPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("store: read session %s: %w", id, err)
	}
	d, err := parseSession(b)
	if err != nil {
		return nil, err
	}
	d.version = hashBytes(b)
	return d, nil
}

// ListSessions reads every durable session, newest first.
func (s *Store) ListSessions() ([]*SessionDoc, error) {
	return s.listSessions()
}

func (s *Store) listSessions() ([]*SessionDoc, error) {
	entries, err := os.ReadDir(s.sessionsDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: scan sessions: %w", err)
	}
	docs := make([]*SessionDoc, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.sessionsDir(), entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("store: read session %s: %w", entry.Name(), err)
		}
		d, err := parseSession(b)
		if err != nil {
			return nil, fmt.Errorf("store: parse session %s: %w", entry.Name(), err)
		}
		d.version = hashBytes(b)
		docs = append(docs, d)
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Session.StartedAt.After(docs[j].Session.StartedAt)
	})
	return docs, nil
}

// TaskSessions returns durable sessions for one task, newest first.
func (s *Store) TaskSessions(taskID string) ([]*SessionDoc, error) {
	docs, err := s.listSessions()
	if err != nil {
		return nil, err
	}
	out := make([]*SessionDoc, 0, len(docs))
	for _, d := range docs {
		if d.Session.TaskID == taskID {
			out = append(out, d)
		}
	}
	return out, nil
}

// FindSessionByIdempotency returns the matching task operation, if any.
func (s *Store) FindSessionByIdempotency(taskID, key string) (*SessionDoc, error) {
	if key == "" {
		return nil, nil
	}
	docs, err := s.TaskSessions(taskID)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if d.Session.IdempotencyKey == key {
			return d, nil
		}
	}
	return nil, nil
}

// FindSessionByIdempotency reads a retry key inside an existing transaction.
func (tx *WriteTx) FindSessionByIdempotency(taskID, key string) (*SessionDoc, error) {
	return tx.store.FindSessionByIdempotency(taskID, key)
}

// LiveSession returns the active durable session for taskID, if one exists.
func (s *Store) LiveSession(taskID string) (*SessionDoc, error) {
	docs, err := s.TaskSessions(taskID)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		if d.Session.Status == session.StatusActive {
			return d, nil
		}
	}
	return nil, nil
}

// LiveSession reads the task's current live session inside an existing transaction.
func (tx *WriteTx) LiveSession(taskID string) (*SessionDoc, error) {
	return tx.store.LiveSession(taskID)
}

// GetSession reads one durable session inside an existing transaction.
func (tx *WriteTx) GetSession(id string) (*SessionDoc, error) { return tx.store.GetSession(id) }

// ReadLive reads ephemeral session state inside an existing transaction.
func (tx *WriteTx) ReadLive(id string) (*session.Live, error) { return tx.store.ReadLive(id) }

// CreateSession persists a new durable session and optional live state.
func (s *Store) CreateSession(ctx context.Context, actor string, value session.Session, live *session.Live) (*SessionDoc, error) {
	var created *SessionDoc
	err := s.Write(ctx, actor, "create session", func(tx *WriteTx) error {
		if current, err := tx.store.LiveSession(value.TaskID); err != nil {
			return err
		} else if current != nil {
			return fmt.Errorf("%w: %s", ErrLiveSession, current.Session.ID)
		}
		d, err := tx.CreateSession(value)
		if err != nil {
			return err
		}
		if live != nil {
			if err := tx.WriteLive(*live); err != nil {
				return err
			}
		}
		created = d
		return nil
	})
	return created, err
}

// SaveSession atomically updates a durable session under the repository lock.
func (s *Store) SaveSession(ctx context.Context, actor string, d *SessionDoc) error {
	return s.Write(ctx, actor, "save session", func(tx *WriteTx) error {
		return tx.SaveSession(d)
	})
}

// CreateSession writes a durable session inside an existing transaction.
func (tx *WriteTx) CreateSession(value session.Session) (*SessionDoc, error) {
	if _, err := os.Stat(tx.store.sessionPath(value.ID)); err == nil {
		return nil, fmt.Errorf("store: session already exists: %s", value.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("store: stat session %s: %w", value.ID, err)
	}
	if err := os.MkdirAll(tx.store.sessionsDir(), 0o755); err != nil {
		return nil, fmt.Errorf("store: create sessions dir: %w", err)
	}
	d, err := newSessionDoc(value)
	if err != nil {
		return nil, err
	}
	if err := tx.SaveSession(d); err != nil {
		return nil, err
	}
	return d, nil
}

// SaveSession writes a durable session inside an existing transaction.
func (tx *WriteTx) SaveSession(d *SessionDoc) error {
	if d.version != "" {
		current, err := os.ReadFile(tx.store.sessionPath(d.Session.ID))
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s was deleted", ErrSessionConflict, d.Session.ID)
		}
		if err != nil {
			return fmt.Errorf("store: reread session %s: %w", d.Session.ID, err)
		}
		if hashBytes(current) != d.version {
			return fmt.Errorf("%w: %s", ErrSessionConflict, d.Session.ID)
		}
	}
	b, err := renderSession(d)
	if err != nil {
		return err
	}
	if err := atomicWrite(tx.store.sessionPath(d.Session.ID), b); err != nil {
		return err
	}
	d.version = hashBytes(b)
	return nil
}

// ReadLive reads ephemeral state. A missing file means no heartbeat has been recorded.
func (s *Store) ReadLive(sessionID string) (*session.Live, error) {
	b, err := os.ReadFile(s.livePath(sessionID))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: read live session %s: %w", sessionID, err)
	}
	var live session.Live
	if err := json.Unmarshal(b, &live); err != nil {
		return nil, fmt.Errorf("store: parse live session %s: %w", sessionID, err)
	}
	return &live, nil
}

// WriteLive atomically replaces ephemeral state inside an existing transaction.
func (tx *WriteTx) WriteLive(live session.Live) error {
	if err := os.MkdirAll(tx.store.liveDir(), 0o755); err != nil {
		return fmt.Errorf("store: create live dir: %w", err)
	}
	b, err := json.MarshalIndent(live, "", "  ")
	if err != nil {
		return fmt.Errorf("store: marshal live session: %w", err)
	}
	b = append(b, '\n')
	return atomicWrite(tx.store.livePath(live.SessionID), b)
}

// DeleteLive removes ephemeral state inside an existing transaction.
func (tx *WriteTx) DeleteLive(sessionID string) error {
	if err := os.Remove(tx.store.livePath(sessionID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("store: remove live session %s: %w", sessionID, err)
	}
	return nil
}

// Replace updates the known session fields while preserving unknown YAML keys.
func (d *SessionDoc) Replace(next session.Session) {
	m := d.mapping()
	setScalar(m, "status", strNode(string(next.Status)))
	if next.EndedAt == nil {
		removeKey(m, "ended_at")
	} else {
		setScalar(m, "ended_at", tsNode(*next.EndedAt))
	}
	setOptionalScalar(m, "head_finished", next.HeadFinished)
	setOptionalScalar(m, "summary", next.Summary)
	setOptionalScalar(m, "cancel_reason", next.CancelReason)
	setUsage(m, next.Usage)
	d.Session = next
}

func newSessionDoc(value session.Session) (*SessionDoc, error) {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		return nil, fmt.Errorf("store: encode session: %w", err)
	}
	return &SessionDoc{Session: value, node: node}, nil
}

func parseSession(b []byte) (*SessionDoc, error) {
	d := &SessionDoc{}
	if err := yaml.Unmarshal(b, &d.node); err != nil {
		return nil, fmt.Errorf("store: parse session: %w", err)
	}
	if err := d.node.Decode(&d.Session); err != nil {
		return nil, fmt.Errorf("store: decode session: %w", err)
	}
	return d, nil
}

func renderSession(d *SessionDoc) ([]byte, error) {
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(d.mapping()); err != nil {
		return nil, fmt.Errorf("store: encode session: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("store: close session encoder: %w", err)
	}
	return out.Bytes(), nil
}

func (d *SessionDoc) mapping() *yaml.Node {
	if d.node.Kind == yaml.DocumentNode && len(d.node.Content) > 0 {
		return d.node.Content[0]
	}
	return &d.node
}

func setOptionalScalar(m *yaml.Node, key, value string) {
	if value == "" {
		removeKey(m, key)
		return
	}
	setScalar(m, key, strNode(value))
}

func setUsage(m *yaml.Node, usage session.Usage) {
	if usage == (session.Usage{}) {
		removeKey(m, "usage")
		return
	}
	value := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if usage.InputTokens > 0 {
		value.Content = append(value.Content, strNode("input_tokens"), int64Node(usage.InputTokens))
	}
	if usage.OutputTokens > 0 {
		value.Content = append(value.Content, strNode("output_tokens"), int64Node(usage.OutputTokens))
	}
	if usage.CachedTokens > 0 {
		value.Content = append(value.Content, strNode("cached_tokens"), int64Node(usage.CachedTokens))
	}
	if usage.ToolCalls > 0 {
		value.Content = append(value.Content, strNode("tool_calls"), int64Node(usage.ToolCalls))
	}
	setScalar(m, "usage", value)
}
