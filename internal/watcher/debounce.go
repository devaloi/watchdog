package watcher

import (
	"sync"
	"time"
)

// Debouncer batches rapid events per file path, firing a callback
// only after no new events arrive within the configured delay.
type Debouncer struct {
	delay   time.Duration
	mu      sync.Mutex
	timers  map[string]*time.Timer
	done    chan struct{}
	doneOne sync.Once
}

// NewDebouncer creates a Debouncer with the given default delay.
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:  delay,
		timers: make(map[string]*time.Timer),
		done:   make(chan struct{}),
	}
}

// Trigger schedules fn to fire after the debounce delay for the given key.
// If Trigger is called again for the same key before the delay expires,
// the timer resets.
func (d *Debouncer) Trigger(key string, fn func()) {
	d.TriggerWithDelay(key, d.delay, fn)
}

// TriggerWithDelay is like Trigger but uses a custom delay.
func (d *Debouncer) TriggerWithDelay(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	select {
	case <-d.done:
		return
	default:
	}

	if existing, ok := d.timers[key]; ok {
		existing.Stop()
	}

	d.timers[key] = time.AfterFunc(delay, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()

		select {
		case <-d.done:
			return
		default:
			fn()
		}
	})
}

// Stop cancels all pending timers.
func (d *Debouncer) Stop() {
	d.doneOne.Do(func() {
		close(d.done)
	})

	d.mu.Lock()
	defer d.mu.Unlock()

	for key, timer := range d.timers {
		timer.Stop()
		delete(d.timers, key)
	}
}
