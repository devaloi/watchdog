package rule

import (
	"testing"

	"github.com/devaloi/watchdog/internal/config"
	"github.com/devaloi/watchdog/internal/watcher"
)

func testConfig() *config.Config {
	return &config.Config{
		Global: config.Global{
			Ignore: []string{".git", "node_modules"},
		},
		Rules: []config.Rule{
			{
				Name:   "Go rebuild",
				Watch:  []string{"**/*.go"},
				Events: []string{"create", "modify"},
				Action: config.Action{Type: "command", Command: "go build ./..."},
			},
			{
				Name:   "CSS reload",
				Watch:  []string{"assets/**/*.css"},
				Events: []string{"modify"},
				Action: config.Action{Type: "webhook", URL: "http://localhost:3000/reload"},
			},
			{
				Name:   "Log all",
				Watch:  []string{"**/*"},
				Events: []string{"create", "modify", "delete", "rename"},
				Action: config.Action{Type: "log", Format: "{{.Event}} {{.Path}}"},
			},
		},
	}
}

func TestEvaluateMatchingGoFile(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: "cmd/main.go", Type: watcher.Create, Name: "main.go", Dir: "cmd"}
	matches := eng.Evaluate(ev)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches (Go rebuild + Log all), got %d", len(matches))
	}

	if matches[0].RuleName != "Go rebuild" {
		t.Errorf("first match = %q, want %q", matches[0].RuleName, "Go rebuild")
	}

	if matches[1].RuleName != "Log all" {
		t.Errorf("second match = %q, want %q", matches[1].RuleName, "Log all")
	}
}

func TestEvaluateCSSFile(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: "assets/style/main.css", Type: watcher.Modify, Name: "main.css", Dir: "assets/style"}
	matches := eng.Evaluate(ev)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches (CSS reload + Log all), got %d", len(matches))
	}

	if matches[0].RuleName != "CSS reload" {
		t.Errorf("first match = %q, want %q", matches[0].RuleName, "CSS reload")
	}
}

func TestEvaluateIgnoredPath(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: ".git/config", Type: watcher.Modify, Name: "config", Dir: ".git"}
	matches := eng.Evaluate(ev)

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for ignored path, got %d", len(matches))
	}
}

func TestEvaluateNoMatchingEvent(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: "cmd/main.go", Type: watcher.Delete, Name: "main.go", Dir: "cmd"}
	matches := eng.Evaluate(ev)

	// Only "Log all" matches delete events for .go files
	if len(matches) != 1 {
		t.Fatalf("expected 1 match (Log all), got %d", len(matches))
	}

	if matches[0].RuleName != "Log all" {
		t.Errorf("match = %q, want %q", matches[0].RuleName, "Log all")
	}
}

func TestEvaluateNoMatch(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: "readme.md", Type: watcher.Delete, Name: "readme.md", Dir: "."}
	matches := eng.Evaluate(ev)

	// Only "Log all" matches
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestEvaluateNodeModulesIgnored(t *testing.T) {
	eng := NewEngine(testConfig())

	ev := watcher.Event{Path: "node_modules/pkg/index.js", Type: watcher.Create, Name: "index.js", Dir: "node_modules/pkg"}
	matches := eng.Evaluate(ev)

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for node_modules path, got %d", len(matches))
	}
}
