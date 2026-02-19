package action

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"text/template"

	"github.com/devaloi/watchdog/internal/watcher"
)

// CommandAction runs a shell command when triggered.
// It kills any previously running instance before starting a new one.
type CommandAction struct {
	CmdTemplate string
	Dir         string
	DryRun      bool
	Output      io.Writer

	mu      sync.Mutex
	cancel  context.CancelFunc
	running *exec.Cmd
}

// NewCommandAction creates a CommandAction with the given command template and working directory.
func NewCommandAction(cmdTemplate, dir string) *CommandAction {
	return &CommandAction{
		CmdTemplate: cmdTemplate,
		Dir:         dir,
		Output:      os.Stdout,
	}
}

// Execute kills any running previous command and starts the command with template variables.
func (c *CommandAction) Execute(ev watcher.Event) error {
	rendered, err := renderTemplate(c.CmdTemplate, NewTemplateData(ev))
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.killPrevious()

	if c.DryRun {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	cmd := exec.CommandContext(ctx, "sh", "-c", rendered) //nolint:gosec // user-configured command
	cmd.Dir = c.Dir
	cmd.Stdout = c.Output
	cmd.Stderr = c.Output

	c.running = cmd

	return cmd.Start()
}

// Stop kills any running command.
func (c *CommandAction) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.killPrevious()
}

func (c *CommandAction) killPrevious() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	if c.running != nil && c.running.Process != nil {
		_ = c.running.Wait()
		c.running = nil
	}
}

func renderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("cmd").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
