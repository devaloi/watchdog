package watcher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncerSingleFire(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	var count atomic.Int32

	for range 10 {
		d.Trigger("main.go", func() {
			count.Add(1)
		})

		time.Sleep(5 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("expected 1 fire, got %d", got)
	}
}

func TestDebouncerIndependentPaths(t *testing.T) {
	d := NewDebouncer(30 * time.Millisecond)
	defer d.Stop()

	var mu sync.Mutex

	fired := make(map[string]int)

	for _, path := range []string{"a.go", "b.go", "c.go"} {
		p := path

		d.Trigger(p, func() {
			mu.Lock()
			fired[p]++
			mu.Unlock()
		})
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	for _, path := range []string{"a.go", "b.go", "c.go"} {
		if fired[path] != 1 {
			t.Errorf("expected %s to fire once, got %d", path, fired[path])
		}
	}
}

func TestDebouncerCustomDelay(t *testing.T) {
	d := NewDebouncer(1 * time.Second)
	defer d.Stop()

	var fired atomic.Bool

	d.TriggerWithDelay("fast.go", 20*time.Millisecond, func() {
		fired.Store(true)
	})

	time.Sleep(80 * time.Millisecond)

	if !fired.Load() {
		t.Error("expected custom delay to fire faster than default")
	}
}

func TestDebouncerStop(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)

	var fired atomic.Bool

	d.Trigger("test.go", func() {
		fired.Store(true)
	})

	d.Stop()

	time.Sleep(100 * time.Millisecond)

	if fired.Load() {
		t.Error("expected no fire after Stop()")
	}
}
