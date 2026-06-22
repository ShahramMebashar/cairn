// Package config loads and saves .cairn/config.yaml and owns id minting (SPEC §3).
// It is a leaf package: it does not import task, so it defines no gate logic — the
// store/mcp layer maps a Config into task.Rules. config.yaml is engine-owned and small,
// so a plain struct round-trip is fine here (unlike task files, which need lossless
// node-level writes per SPEC §8).
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrInvalidConfig wraps every validation failure so callers match with errors.Is.
var ErrInvalidConfig = errors.New("invalid config")

// Config mirrors config.yaml (SPEC §3). states are user-defined free strings; there is
// no hardcoded status enum. closed must be a subset of states; initial must be a state.
type Config struct {
	Prefix              string   `yaml:"prefix"`
	Counter             int      `yaml:"counter"`
	States              []string `yaml:"states"`
	Closed              []string `yaml:"closed"`
	Initial             string   `yaml:"initial"`
	CheckTimeoutDefault int      `yaml:"check_timeout_default"`
	WorkingState        string   `yaml:"working_state,omitempty"`
	ReviewState         string   `yaml:"review_state,omitempty"`
	SessionHeartbeat    int      `yaml:"session_heartbeat_interval,omitempty"`
	SessionStaleAfter   int      `yaml:"session_stale_after,omitempty"`
}

// Default returns the standard starting config for a freshly initialized repo. The
// prefix is caller-supplied (derived from the project name); the rest are sensible v0
// defaults matching SPEC §3.
func Default(prefix string) Config {
	return Config{
		Prefix:              prefix,
		Counter:             0,
		States:              []string{"backlog", "in_progress", "in_review", "done", "canceled"},
		Closed:              []string{"done", "canceled"},
		Initial:             "backlog",
		CheckTimeoutDefault: 120,
		WorkingState:        "in_progress",
		ReviewState:         "in_review",
		SessionHeartbeat:    30,
		SessionStaleAfter:   180,
	}
}

// Load reads and validates config.yaml.
func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Save writes config back to path. The whole file is small and engine-owned, so a full
// rewrite is acceptable (no node-level surgery needed, unlike task files).
func Save(path string, c Config) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("config: create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("config: chmod temp file: %w", err)
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("config: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("config: close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("config: replace %s: %w", path, err)
	}
	return nil
}

// Validate enforces the §3 invariants: at least one state, a known initial state, and a
// closed set drawn from states.
func (c Config) Validate() error {
	if len(c.States) == 0 {
		return fmt.Errorf("%w: states is empty", ErrInvalidConfig)
	}
	if !c.isState(c.Initial) {
		return fmt.Errorf("%w: initial %q is not in states", ErrInvalidConfig, c.Initial)
	}
	for _, s := range c.Closed {
		if !c.isState(s) {
			return fmt.Errorf("%w: closed state %q is not in states", ErrInvalidConfig, s)
		}
	}
	for name, state := range map[string]string{"working_state": c.WorkingState, "review_state": c.ReviewState} {
		if state != "" && !c.isState(state) {
			return fmt.Errorf("%w: %s %q is not in states", ErrInvalidConfig, name, state)
		}
	}
	return nil
}

func (c Config) isState(s string) bool {
	return slices.Contains(c.States, s)
}

// NewID mints the next task id and returns a copy of the config with the counter
// advanced. It is pure — the caller persists `next` (under the repository lock, SPEC §3) so
// the monotonic counter is never lost or reused.
func (c Config) NewID() (id string, next Config) {
	next = c
	next.Counter = c.Counter + 1
	return fmt.Sprintf("%s-%03d", c.Prefix, next.Counter), next
}

// CheckTimeout resolves a per-check timeout in seconds to a duration, falling back to
// check_timeout_default when the check omits one (SPEC §6).
func (c Config) CheckTimeout(perCheckSeconds int) time.Duration {
	if perCheckSeconds <= 0 {
		perCheckSeconds = c.CheckTimeoutDefault
	}
	return time.Duration(perCheckSeconds) * time.Second
}

// Working returns the configured first working state, or the first non-initial, non-closed
// state for legacy configurations.
func (c Config) Working() string {
	if c.WorkingState != "" {
		return c.WorkingState
	}
	for _, state := range c.States {
		if state != c.Initial && !slices.Contains(c.Closed, state) {
			return state
		}
	}
	return c.Initial
}

// Review returns the configured review state when one exists.
func (c Config) Review() string {
	if c.ReviewState != "" {
		return c.ReviewState
	}
	for _, state := range c.States {
		if state == "in_review" {
			return state
		}
	}
	return ""
}

// SessionStaleDuration returns the heartbeat expiry, defaulting to three minutes.
func (c Config) SessionStaleDuration() time.Duration {
	seconds := c.SessionStaleAfter
	if seconds <= 0 {
		seconds = 180
	}
	return time.Duration(seconds) * time.Second
}

// SessionHeartbeatDuration returns the cadence at which agents should heartbeat,
// defaulting to thirty seconds. It mirrors SessionStaleDuration so configs missing the
// field (e.g. a zero-value Config) still resolve a sensible interval.
func (c Config) SessionHeartbeatDuration() time.Duration {
	seconds := c.SessionHeartbeat
	if seconds <= 0 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}
