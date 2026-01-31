package lock

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// Lock provides exclusive locking for wiz state mutations.
// It uses both an in-process mutex (for goroutine safety) and a file lock
// (for cross-process safety).
type Lock struct {
	mu   sync.Mutex
	fl   *flock.Flock
	path string
}

// New creates a new Lock at the given path.
func New(path string) *Lock {
	return &Lock{
		fl:   flock.New(path),
		path: path,
	}
}

// Acquire obtains an exclusive lock, blocking until acquired or context cancelled.
func (l *Lock) Acquire(ctx context.Context) error {
	l.mu.Lock()
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		l.mu.Unlock()
		return fmt.Errorf("create lock dir: %w", err)
	}
	ok, err := l.fl.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		l.mu.Unlock()
		return fmt.Errorf("acquire lock %s: %w", l.path, err)
	}
	if !ok {
		l.mu.Unlock()
		return fmt.Errorf("could not acquire lock %s (timed out)", l.path)
	}
	return nil
}

// TryAcquire attempts to get the lock without blocking.
func (l *Lock) TryAcquire() (bool, error) {
	if !l.mu.TryLock() {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		l.mu.Unlock()
		return false, err
	}
	ok, err := l.fl.TryLock()
	if !ok || err != nil {
		l.mu.Unlock()
		return false, err
	}
	return true, nil
}

// Release drops the lock.
func (l *Lock) Release() error {
	defer l.mu.Unlock()
	return l.fl.Unlock()
}

// WithLock acquires the lock, runs fn, then releases. Returns fn's error or lock errors.
func (l *Lock) WithLock(ctx context.Context, fn func() error) error {
	if err := l.Acquire(ctx); err != nil {
		return err
	}
	defer l.Release()
	return fn()
}
