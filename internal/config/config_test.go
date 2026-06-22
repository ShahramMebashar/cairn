package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const sample = `prefix: PROJ
counter: 2
states: [backlog, in_progress, in_review, done, canceled]
closed: [done, canceled]
initial: backlog
check_timeout_default: 120
`

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	c, err := Load(writeConfig(t, sample))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Prefix != "PROJ" || c.Counter != 2 || c.Initial != "backlog" || c.CheckTimeoutDefault != 120 {
		t.Fatalf("unexpected config: %+v", c)
	}
	if len(c.States) != 5 || len(c.Closed) != 2 {
		t.Fatalf("unexpected states/closed: %+v", c)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSaveRoundTrips(t *testing.T) {
	path := writeConfig(t, sample)
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	c.Counter = 41
	if err := Save(path, c); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Counter != 41 || got.Prefix != "PROJ" {
		t.Fatalf("round-trip lost data: %+v", got)
	}
}

func TestNewID(t *testing.T) {
	c, err := Load(writeConfig(t, sample))
	if err != nil {
		t.Fatal(err)
	}
	id, next := c.NewID()
	if id != "PROJ-003" {
		t.Fatalf("id = %q, want PROJ-003", id)
	}
	if next.Counter != 3 {
		t.Fatalf("counter = %d, want 3", next.Counter)
	}
	if c.Counter != 2 {
		t.Fatalf("NewID mutated the receiver: counter = %d", c.Counter)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{"initial not a state", "prefix: P\ncounter: 0\nstates: [a, b]\nclosed: [b]\ninitial: zzz\ncheck_timeout_default: 1\n", ErrInvalidConfig},
		{"closed not a subset", "prefix: P\ncounter: 0\nstates: [a, b]\nclosed: [c]\ninitial: a\ncheck_timeout_default: 1\n", ErrInvalidConfig},
		{"empty states", "prefix: P\ncounter: 0\nstates: []\nclosed: []\ninitial: a\ncheck_timeout_default: 1\n", ErrInvalidConfig},
		{"valid", sample, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.body))
			if !errors.Is(err, tt.want) {
				t.Fatalf("Load = %v, want errors.Is %v", err, tt.want)
			}
		})
	}
}
