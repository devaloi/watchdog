package config

import (
	"net/http"
	"testing"
	"time"
)

const actionTypeCommand = "command"

const validYAML = `
global:
  debounce: 500ms
  ignore:
    - .git
    - node_modules
rules:
  - name: "Go rebuild"
    watch:
      - "**/*.go"
    events: [create, modify]
    action:
      type: command
      command: "go build ./..."
      dir: "."
  - name: "Log all"
    watch:
      - "**/*"
    events: [create, modify, delete, rename]
    action:
      type: log
      format: "[{{.Time}}] {{.Event}} {{.Path}}"
`

func TestParseValid(t *testing.T) {
	cfg, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Global.Debounce.Duration != 500*time.Millisecond {
		t.Errorf("debounce = %v, want 500ms", cfg.Global.Debounce.Duration)
	}

	if len(cfg.Global.Ignore) != 2 {
		t.Errorf("expected 2 ignore patterns, got %d", len(cfg.Global.Ignore))
	}

	if len(cfg.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(cfg.Rules))
	}

	if cfg.Rules[0].Name != "Go rebuild" {
		t.Errorf("first rule name = %q, want %q", cfg.Rules[0].Name, "Go rebuild")
	}

	if cfg.Rules[0].Action.Type != actionTypeCommand {
		t.Errorf("first rule action type = %q, want %q", cfg.Rules[0].Action.Type, "command")
	}
}

func TestParseNoRules(t *testing.T) {
	_, err := Parse([]byte(`global: {}`))
	if err == nil {
		t.Fatal("expected error for config with no rules")
	}
}

func TestParseMissingName(t *testing.T) {
	input := `
rules:
  - watch: ["**/*.go"]
    action:
      type: command
      command: "go build"
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for rule without name")
	}
}

func TestParseMissingWatch(t *testing.T) {
	input := `
rules:
  - name: "test"
    action:
      type: command
      command: "go build"
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for rule without watch patterns")
	}
}

func TestParseInvalidActionType(t *testing.T) {
	input := `
rules:
  - name: "test"
    watch: ["*.go"]
    action:
      type: invalid
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid action type")
	}
}

func TestParseMissingCommand(t *testing.T) {
	input := `
rules:
  - name: "test"
    watch: ["*.go"]
    action:
      type: command
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for command action without command")
	}
}

func TestParseMissingWebhookURL(t *testing.T) {
	input := `
rules:
  - name: "test"
    watch: ["*.go"]
    action:
      type: webhook
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for webhook action without url")
	}
}

func TestParseMissingLogFormat(t *testing.T) {
	input := `
rules:
  - name: "test"
    watch: ["*.go"]
    action:
      type: log
`

	_, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for log action without format")
	}
}

func TestParseWebhookAction(t *testing.T) {
	input := `
rules:
  - name: "Notify"
    watch: ["**/*.css"]
    events: [modify]
    debounce: 1s
    action:
      type: webhook
      url: "http://localhost:3000/reload"
      method: POST
`

	cfg, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := cfg.Rules[0]
	if r.Debounce.Duration != time.Second {
		t.Errorf("debounce = %v, want 1s", r.Debounce.Duration)
	}

	if r.Action.URL != "http://localhost:3000/reload" {
		t.Errorf("url = %q", r.Action.URL)
	}

	if r.Action.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", r.Action.Method)
	}
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse([]byte(`{invalid yaml`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
