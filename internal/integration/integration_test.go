package integration

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devaloi/watchdog/internal/action"
	"github.com/devaloi/watchdog/internal/config"
	"github.com/devaloi/watchdog/internal/rule"
	"github.com/devaloi/watchdog/internal/watcher"
)

type syncWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.Write(p)
}

func (w *syncWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.String()
}

var _ io.Writer = (*syncWriter)(nil)

func TestWatcherToRuleToAction(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Global: config.Global{
			Debounce: config.Duration{Duration: 50 * time.Millisecond},
			Ignore:   []string{".git"},
		},
		Rules: []config.Rule{
			{
				Name:   "Log Go changes",
				Watch:  []string{"**/*.go"},
				Events: []string{"create", "modify"},
				Action: config.Action{Type: "log", Format: "{{.Event}} {{.Path}}"},
			},
		},
	}

	eng := rule.NewEngine(cfg)

	logBuf := &syncWriter{}
	logAction := action.NewLogAction("{{.Event}} {{.Name}}", logBuf)

	w, err := watcher.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	deb := watcher.NewDebouncer(cfg.Global.Debounce.Duration)
	defer deb.Stop()

	done := make(chan struct{})

	go func() {
		for ev := range w.Events {
			relPath, relErr := filepath.Rel(dir, ev.Path)
			if relErr == nil {
				ev.Path = relPath
				ev.Name = filepath.Base(relPath)
				ev.Dir = filepath.Dir(relPath)
			}

			matches := eng.Evaluate(ev)
			for _, m := range matches {
				capturedEv := ev

				deb.Trigger(m.RuleName+":"+ev.Path, func() {
					_ = logAction.Execute(capturedEv)
				})
			}
		}

		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	goFile := filepath.Join(dir, "main.go")

	writeErr := os.WriteFile(goFile, []byte("package main"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	// Wait for debounce to fire
	time.Sleep(200 * time.Millisecond)

	output := logBuf.String()
	if !strings.Contains(output, "main.go") {
		t.Errorf("expected log output to contain 'main.go', got %q", output)
	}
}

func TestWatcherIgnoresFilteredPaths(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Global: config.Global{
			Debounce: config.Duration{Duration: 50 * time.Millisecond},
			Ignore:   []string{"*.tmp"},
		},
		Rules: []config.Rule{
			{
				Name:   "Log all",
				Watch:  []string{"**/*"},
				Events: []string{"create"},
				Action: config.Action{Type: "log", Format: "{{.Name}}"},
			},
		},
	}

	eng := rule.NewEngine(cfg)

	logBuf := &syncWriter{}
	logAction := action.NewLogAction("{{.Name}}", logBuf)

	w, err := watcher.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	deb := watcher.NewDebouncer(cfg.Global.Debounce.Duration)
	defer deb.Stop()

	go func() {
		for ev := range w.Events {
			relPath, relErr := filepath.Rel(dir, ev.Path)
			if relErr == nil {
				ev.Path = relPath
				ev.Name = filepath.Base(relPath)
			}

			matches := eng.Evaluate(ev)
			for _, m := range matches {
				capturedEv := ev

				deb.Trigger(m.RuleName+":"+ev.Path, func() {
					_ = logAction.Execute(capturedEv)
				})
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	tmpFile := filepath.Join(dir, "cache.tmp")

	writeErr := os.WriteFile(tmpFile, []byte("temp"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	time.Sleep(200 * time.Millisecond)

	if strings.Contains(logBuf.String(), "cache.tmp") {
		t.Error("expected .tmp file to be ignored")
	}
}

func TestMultipleRulesMatchSameEvent(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		Global: config.Global{
			Debounce: config.Duration{Duration: 50 * time.Millisecond},
		},
		Rules: []config.Rule{
			{
				Name:   "Rule A",
				Watch:  []string{"**/*.go"},
				Events: []string{"create"},
				Action: config.Action{Type: "log", Format: "A:{{.Name}}"},
			},
			{
				Name:   "Rule B",
				Watch:  []string{"**/*"},
				Events: []string{"create"},
				Action: config.Action{Type: "log", Format: "B:{{.Name}}"},
			},
		},
	}

	eng := rule.NewEngine(cfg)

	logBuf := &syncWriter{}

	actionA := action.NewLogAction("A:{{.Name}}", logBuf)
	actionB := action.NewLogAction("B:{{.Name}}", logBuf)

	actions := map[string]action.Action{
		"Rule A": actionA,
		"Rule B": actionB,
	}

	w, err := watcher.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = w.Close() }()

	deb := watcher.NewDebouncer(cfg.Global.Debounce.Duration)
	defer deb.Stop()

	go func() {
		for ev := range w.Events {
			relPath, relErr := filepath.Rel(dir, ev.Path)
			if relErr == nil {
				ev.Path = relPath
				ev.Name = filepath.Base(relPath)
			}

			matches := eng.Evaluate(ev)
			for _, m := range matches {
				a := actions[m.RuleName]
				capturedEv := ev

				deb.Trigger(m.RuleName+":"+ev.Path, func() {
					_ = a.Execute(capturedEv)
				})
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	goFile := filepath.Join(dir, "app.go")

	writeErr := os.WriteFile(goFile, []byte("package app"), 0o600)
	if writeErr != nil {
		t.Fatal(writeErr)
	}

	time.Sleep(200 * time.Millisecond)

	output := logBuf.String()
	if !strings.Contains(output, "A:app.go") {
		t.Errorf("expected Rule A to fire, got %q", output)
	}

	if !strings.Contains(output, "B:app.go") {
		t.Errorf("expected Rule B to fire, got %q", output)
	}
}

func TestConfigLoadFromFile(t *testing.T) {
	cfg, err := config.Load("../../testdata/watchdog.yaml")
	if err != nil {
		t.Fatalf("failed to load testdata config: %v", err)
	}

	if len(cfg.Rules) == 0 {
		t.Error("expected at least one rule from testdata config")
	}

	if cfg.Global.Debounce.Duration != 100*time.Millisecond {
		t.Errorf("debounce = %v, want 100ms", cfg.Global.Debounce.Duration)
	}
}
