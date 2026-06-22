package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"

	"cairn/internal/config"
	"cairn/internal/task"
)

const (
	defaultLockTimeout = 5 * time.Second
	lockRetryInterval  = 10 * time.Millisecond
)

// ErrLockTimeout is returned when another Cairn process holds the repository write lock
// beyond the caller's deadline.
var ErrLockTimeout = errors.New("repository write lock timeout")

// WriteTx is a short, repository-exclusive mutation scope. Long-running work such as
// command execution must happen before entering a transaction.
type WriteTx struct {
	store *Store
}

// Write serializes a short mutation across every Cairn process using this repository.
func (s *Store) Write(ctx context.Context, actor, operation string, fn func(*WriteTx) error) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultLockTimeout)
		defer cancel()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("store: open write lock: %w", err)
	}
	defer f.Close()

	if err := acquireLock(ctx, f); err != nil {
		return err
	}
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN) //nolint:errcheck // closing the fd also releases the lock

	if err := writeLockDiagnostic(f, actor, operation); err != nil {
		return err
	}
	return fn(&WriteTx{store: s})
}

func acquireLock(ctx context.Context, f *os.File) error {
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EAGAIN) {
			return fmt.Errorf("store: acquire write lock: %w", err)
		}

		timer := time.NewTimer(lockRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("%w: %w", ErrLockTimeout, ctx.Err())
		case <-timer.C:
		}
	}
}

func writeLockDiagnostic(f *os.File, actor, operation string) error {
	diagnostic := struct {
		PID       int    `json:"pid"`
		Actor     string `json:"actor,omitempty"`
		Operation string `json:"operation,omitempty"`
		Acquired  string `json:"acquiredAt"`
	}{
		PID:       os.Getpid(),
		Actor:     actor,
		Operation: operation,
		Acquired:  time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(diagnostic)
	if err != nil {
		return fmt.Errorf("store: marshal write-lock diagnostic: %w", err)
	}
	b = append(b, '\n')
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("store: truncate write lock: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("store: seek write lock: %w", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("store: write lock diagnostic: %w", err)
	}
	return nil
}

// SaveTask writes a task inside an existing repository transaction.
func (tx *WriteTx) SaveTask(d *Doc) error {
	return tx.store.save(d)
}

// GetTask reads a task inside an existing transaction.
func (tx *WriteTx) GetTask(id string) (*Doc, error) { return tx.store.Get(id) }

// Tasks reads the validated task graph inside an existing transaction.
func (tx *WriteTx) Tasks() (map[string]task.Task, error) { return tx.store.List() }

// Config reads repository configuration inside an existing transaction.
func (tx *WriteTx) Config() (config.Config, error) { return tx.store.Config() }
