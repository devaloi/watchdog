// Package watcher wraps fsnotify to provide recursive directory watching
// with automatic tracking of new directories.
package watcher

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// EventType represents the kind of file system change.
type EventType string

// Supported event types.
const (
	Create EventType = "create"
	Modify EventType = "modify"
	Delete EventType = "delete"
	Rename EventType = "rename"
)

// Event represents a single file system change.
type Event struct {
	Path string
	Type EventType
	Name string
	Dir  string
}

// Watcher recursively watches directories and emits Events.
type Watcher struct {
	fsw    *fsnotify.Watcher
	Events chan Event
	Errors chan error
	done   chan struct{}
	wg     sync.WaitGroup
	root   string
}

// New creates a Watcher that recursively watches the given root directory.
func New(root string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fsw:    fsw,
		Events: make(chan Event, 128),
		Errors: make(chan error, 16),
		done:   make(chan struct{}),
		root:   root,
	}

	addErr := w.addRecursive(root)
	if addErr != nil {
		_ = fsw.Close()

		return nil, addErr
	}

	w.wg.Add(1)

	go w.loop()

	return w, nil
}

// Close stops the watcher and releases resources.
func (w *Watcher) Close() error {
	close(w.done)
	err := w.fsw.Close()
	w.wg.Wait()

	return err
}

// WatchedDirs returns the list of directories currently being watched.
func (w *Watcher) WatchedDirs() []string {
	list := w.fsw.WatchList()

	return list
}

// ParseEventType converts a string to an EventType, returning an error for unknown values.
func ParseEventType(s string) (EventType, error) {
	switch EventType(s) {
	case Create, Modify, Delete, Rename:
		return EventType(s), nil
	default:
		return "", errors.New("unknown event type: " + s)
	}
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return w.fsw.Add(path)
		}

		return nil
	})
}

func (w *Watcher) loop() {
	defer w.wg.Done()
	defer close(w.Events)
	defer close(w.Errors)

	for {
		select {
		case <-w.done:
			return

		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			event := translate(ev)
			if event == nil {
				continue
			}

			if event.Type == Create {
				info, statErr := os.Stat(ev.Name)
				if statErr == nil && info.IsDir() {
					_ = w.addRecursive(ev.Name)
				}
			}

			select {
			case w.Events <- *event:
			case <-w.done:
				return
			}

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}

			select {
			case w.Errors <- err:
			case <-w.done:
				return
			}
		}
	}
}

func translate(ev fsnotify.Event) *Event {
	var t EventType

	switch {
	case ev.Op.Has(fsnotify.Create):
		t = Create
	case ev.Op.Has(fsnotify.Write):
		t = Modify
	case ev.Op.Has(fsnotify.Remove):
		t = Delete
	case ev.Op.Has(fsnotify.Rename):
		t = Rename
	default:
		return nil
	}

	rel := ev.Name

	return &Event{
		Path: rel,
		Type: t,
		Name: filepath.Base(rel),
		Dir:  filepath.Dir(rel),
	}
}
