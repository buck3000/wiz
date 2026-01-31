package lock_test

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/firewood-buck-3000/wiz/internal/lock"
)

func TestWithLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	l := lock.New(path)

	called := false
	err := l.WithLock(context.Background(), func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("fn not called")
	}
}

func TestLockExclusivity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	var counter int64
	var maxConcurrent int64
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := lock.New(path)
			err := l.WithLock(context.Background(), func() error {
				cur := atomic.AddInt64(&counter, 1)
				if cur > 1 {
					atomic.StoreInt64(&maxConcurrent, cur)
				}
				time.Sleep(5 * time.Millisecond)
				atomic.AddInt64(&counter, -1)
				return nil
			})
			if err != nil {
				t.Errorf("lock error: %v", err)
			}
		}()
	}
	wg.Wait()

	if maxConcurrent > 1 {
		t.Errorf("max concurrent = %d, want 1", maxConcurrent)
	}
}

func TestTryAcquire(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")
	l1 := lock.New(path)
	l2 := lock.New(path)

	ok, err := l1.TryAcquire()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("first TryAcquire should succeed")
	}

	ok2, err := l2.TryAcquire()
	if err != nil {
		t.Fatal(err)
	}
	if ok2 {
		t.Fatal("second TryAcquire should fail while lock is held")
	}

	l1.Release()

	ok3, err := l2.TryAcquire()
	if err != nil {
		t.Fatal(err)
	}
	if !ok3 {
		t.Fatal("TryAcquire should succeed after release")
	}
	l2.Release()
}
