package action

import (
	"bytes"
	"io"
	"text/template"

	"github.com/devaloi/watchdog/internal/watcher"
)

// LogAction writes formatted log lines when triggered.
type LogAction struct {
	Format string
	Output io.Writer
	DryRun bool
}

// NewLogAction creates a LogAction with the given format template.
func NewLogAction(format string, output io.Writer) *LogAction {
	return &LogAction{
		Format: format,
		Output: output,
	}
}

// Execute renders the format template and writes to the output.
func (l *LogAction) Execute(ev watcher.Event) error {
	if l.DryRun {
		return nil
	}

	data := NewTemplateData(ev)

	t, err := template.New("log").Parse(l.Format)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		return err
	}

	buf.WriteString("\n")

	_, err = l.Output.Write(buf.Bytes())

	return err
}
