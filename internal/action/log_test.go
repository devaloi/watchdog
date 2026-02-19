package action

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devaloi/watchdog/internal/watcher"
)

func TestLogActionExecute(t *testing.T) {
	var buf bytes.Buffer

	la := NewLogAction("{{.Event}} {{.Path}}", &buf)

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := la.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "modify main.go" {
		t.Errorf("got %q, want %q", got, "modify main.go")
	}
}

func TestLogActionTemplateVars(t *testing.T) {
	var buf bytes.Buffer

	la := NewLogAction("[{{.Time}}] {{.Event}} {{.Name}} in {{.Dir}}", &buf)

	ev := watcher.Event{Path: "src/app.go", Type: watcher.Create, Name: "app.go", Dir: "src"}

	err := la.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, want := range []string{"create", "app.go", "src"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestLogActionDryRun(t *testing.T) {
	var buf bytes.Buffer

	la := NewLogAction("{{.Event}} {{.Path}}", &buf)
	la.DryRun = true

	ev := watcher.Event{Path: "main.go", Type: watcher.Modify, Name: "main.go", Dir: "."}

	err := la.Execute(ev)
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() > 0 {
		t.Errorf("expected no output in dry run, got %q", buf.String())
	}
}
