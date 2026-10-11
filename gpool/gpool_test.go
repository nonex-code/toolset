package gpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunAllTasks(t *testing.T) {
	const n = 100
	pool := NewTaskPool(10)
	var count int64
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		if err := pool.Submit(func() {
			atomic.AddInt64(&count, 1)
			wg.Done()
		}); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
	pool.Close()

	if got := atomic.LoadInt64(&count); got != n {
		t.Errorf("executed %d tasks, want %d", got, n)
	}
}

func TestRunningNeverExceedsSize(t *testing.T) {
	pool := NewTaskPool(3)
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 20; i++ {
		if err := pool.Submit(func() {
			time.Sleep(time.Millisecond)
			wg.Done()
		}); err != nil {
			t.Fatal(err)
		}
		if got := pool.Running(); got > pool.GetSize() {
			t.Errorf("running=%d exceeds size=%d", got, pool.GetSize())
		}
	}
	wg.Wait()
	pool.Close()
}

func TestSubmitAfterClose(t *testing.T) {
	pool := NewTaskPool(2)
	pool.Close()
	if err := pool.Submit(func() {}); err == nil {
		t.Error("expected error when submitting to a closed pool")
	}
}

func TestSubmitNilTask(t *testing.T) {
	pool := NewTaskPool(2)
	defer pool.Close()
	if err := pool.Submit(nil); err == nil {
		t.Error("expected error when submitting a nil task")
	}
}

func TestCloseWaitsForTasks(t *testing.T) {
	pool := NewTaskPool(4)
	const n = 20
	var done int64

	for i := 0; i < n; i++ {
		if err := pool.Submit(func() {
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&done, 1)
		}); err != nil {
			t.Fatal(err)
		}
	}
	pool.Close()

	if got := atomic.LoadInt64(&done); got != n {
		t.Errorf("Close returned before all tasks finished: %d/%d", got, n)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	pool := NewTaskPool(2)
	pool.Close()
	// Must not panic (close of closed channel) and must return.
	pool.Close()
}

func TestTaskPanicDoesNotHang(t *testing.T) {
	pool := NewTaskPool(2)
	if err := pool.Submit(func() { panic("boom") }); err != nil {
		t.Fatal(err)
	}
	if err := pool.Submit(func() {}); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		pool.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Close hung after a task panicked")
	}
}

func TestZeroSizePoolDoesNotDeadlock(t *testing.T) {
	pool := NewTaskPool(0)

	done := make(chan struct{})
	go func() {
		if err := pool.Submit(func() {}); err != nil {
			t.Error(err)
		}
		pool.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("zero-size pool deadlocked")
	}
}

func TestConcurrentSubmit(t *testing.T) {
	pool := NewTaskPool(8)
	const goroutines = 16
	const perGoroutine = 50
	var count int64

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				if err := pool.Submit(func() {
					atomic.AddInt64(&count, 1)
				}); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	pool.Close()

	if got := atomic.LoadInt64(&count); got != goroutines*perGoroutine {
		t.Errorf("executed %d tasks, want %d", got, goroutines*perGoroutine)
	}
}
