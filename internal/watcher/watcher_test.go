package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherCreateEvent(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	time.Sleep(50 * time.Millisecond)

	path := filepath.Join(dir, "test.txt")

	writeErr := os.WriteFile(path, []byte("hello"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	ev := waitForEvent(t, w)
	if ev.Type != Create {
		t.Errorf("expected create event, got %s", ev.Type)
	}

	if ev.Name != "test.txt" {
		t.Errorf("expected name test.txt, got %s", ev.Name)
	}
}

func TestWatcherModifyEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.txt")

	writeErr := os.WriteFile(path, []byte("initial"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	time.Sleep(50 * time.Millisecond)

	writeErr = os.WriteFile(path, []byte("updated"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	ev := waitForEvent(t, w)
	if ev.Type != Modify {
		t.Errorf("expected modify event, got %s", ev.Type)
	}
}

func TestWatcherDeleteEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todelete.txt")

	writeErr := os.WriteFile(path, []byte("bye"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	time.Sleep(50 * time.Millisecond)

	rmErr := os.Remove(path)
	if rmErr != nil {
		t.Fatal(rmErr)
	}

	ev := waitForEvent(t, w)
	if ev.Type != Delete && ev.Type != Rename {
		t.Errorf("expected delete or rename event, got %s", ev.Type)
	}
}

func TestWatcherRecursiveSubdir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")

	mkErr := os.MkdirAll(sub, 0o750)
	if mkErr != nil {
		t.Fatal(mkErr)
	}

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	time.Sleep(50 * time.Millisecond)

	path := filepath.Join(sub, "nested.txt")

	writeErr := os.WriteFile(path, []byte("data"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	ev := waitForEvent(t, w)
	if ev.Type != Create {
		t.Errorf("expected create event from subdir, got %s", ev.Type)
	}
}

func TestWatcherNewDirAutoWatch(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	time.Sleep(50 * time.Millisecond)

	newSub := filepath.Join(dir, "newdir")

	mkErr := os.MkdirAll(newSub, 0o750)
	if mkErr != nil {
		t.Fatal(mkErr)
	}

	_ = waitForEvent(t, w)

	time.Sleep(100 * time.Millisecond)

	path := filepath.Join(newSub, "auto.txt")

	writeErr := os.WriteFile(path, []byte("auto"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	ev := waitForEvent(t, w)
	if ev.Name != "auto.txt" {
		t.Errorf("expected auto.txt event from auto-watched dir, got %s", ev.Name)
	}
}

func TestParseEventType(t *testing.T) {
	tests := []struct {
		input string
		want  EventType
		err   bool
	}{
		{"create", Create, false},
		{"modify", Modify, false},
		{"delete", Delete, false},
		{"rename", Rename, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		got, err := ParseEventType(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseEventType(%q): error=%v, wantErr=%v", tt.input, err, tt.err)
		}

		if got != tt.want {
			t.Errorf("ParseEventType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func waitForEvent(t *testing.T, w *Watcher) Event {
	t.Helper()

	select {
	case ev := <-w.Events:
		return ev
	case err := <-w.Errors:
		t.Fatalf("watcher error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	return Event{}
}
