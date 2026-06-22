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
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("config: write %s: %w", path, err)
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
	return nil
}

func (c Config) isState(s string) bool {
	return slices.Contains(c.States, s)
}

// NewID mints the next task id and returns a copy of the config with the counter
// advanced. It is pure — the caller persists `next` (under the store mutex, SPEC §3) so
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
