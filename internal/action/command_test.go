package action

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/devaloi/watchdog/internal/watcher"
)

func TestCommandActionExecute(t *testing.T) {
	var buf bytes.Buffer

	cmd := NewCommandAction("echo hello", ".")
	cmd.Output = &buf

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := cmd.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for command to complete
	time.Sleep(100 * time.Millisecond)
	cmd.Stop()

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected output to contain 'hello', got %q", buf.String())
	}
}

func TestCommandActionTemplateVars(t *testing.T) {
	var buf bytes.Buffer

	cmd := NewCommandAction("echo {{.Path}} {{.Event}} {{.Name}} {{.Dir}}", ".")
	cmd.Output = &buf

	ev := watcher.Event{Path: "src/main.go", Type: watcher.Create, Name: "main.go", Dir: "src"}

	err := cmd.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	cmd.Stop()

	output := buf.String()
	for _, want := range []string{"src/main.go", "create", "main.go", "src"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestCommandActionKillsPrevious(t *testing.T) {
	cmd := NewCommandAction("sleep 10", ".")

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := cmd.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	// Second execution should kill the first
	err = cmd.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	cmd.Stop()
}

func TestCommandActionDryRun(t *testing.T) {
	var buf bytes.Buffer

	cmd := NewCommandAction("echo should-not-run", ".")
	cmd.Output = &buf
	cmd.DryRun = true

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := cmd.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	if buf.Len() > 0 {
		t.Errorf("expected no output in dry run, got %q", buf.String())
	}
}
