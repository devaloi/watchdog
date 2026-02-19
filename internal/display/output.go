// Package display provides colorized terminal output for watchdog events.
package display

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/devaloi/watchdog/internal/config"
	"github.com/devaloi/watchdog/internal/watcher"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
)

// Output writes formatted event information to the terminal.
type Output struct {
	Writer io.Writer
}

// NewOutput creates an Output writing to stdout.
func NewOutput() *Output {
	return &Output{Writer: os.Stdout}
}

// Banner prints the startup banner with watched paths and rules.
func (o *Output) Banner(cfg *config.Config, configPath string) {
	w := o.Writer

	write(w, "\n")
	write(w, colorBold+"  🐕 watchdog"+colorReset+" — file system watcher\n")
	write(w, colorDim+"  config: "+configPath+colorReset+"\n")

	if cfg.Global.Debounce.Duration > 0 {
		write(w, colorDim+"  debounce: "+cfg.Global.Debounce.String()+colorReset+"\n")
	}

	if len(cfg.Global.Ignore) > 0 {
		write(w, colorDim+"  ignore: "+strings.Join(cfg.Global.Ignore, ", ")+colorReset+"\n")
	}

	write(w, "\n")

	for _, r := range cfg.Rules {
		write(w, "  "+colorCyan+"▸"+colorReset+" "+r.Name)
		write(w, colorDim+" ["+r.Action.Type+"]"+colorReset)
		write(w, colorDim+" "+strings.Join(r.Watch, ", ")+colorReset+"\n")
	}

	write(w, "\n")
	write(w, colorDim+"  Watching for changes... (Ctrl+C to stop)"+colorReset+"\n\n")
}

// Event prints a colorized event line.
func (o *Output) Event(ev watcher.Event, ruleName string) {
	color := colorForEvent(ev.Type)
	ts := time.Now().Format("15:04:05")

	write(o.Writer, colorDim+ts+colorReset+" "+color+string(ev.Type)+colorReset+" "+ev.Path)

	if ruleName != "" {
		write(o.Writer, colorDim+" → "+ruleName+colorReset)
	}

	write(o.Writer, "\n")
}

// ActionResult prints the result of an action execution.
func (o *Output) ActionResult(ruleName string, err error, elapsed time.Duration) {
	if err != nil {
		write(o.Writer, "  "+colorRed+"✗"+colorReset+" "+ruleName+": "+err.Error()+"\n")

		return
	}

	write(o.Writer, "  "+colorGreen+"✓"+colorReset+" "+ruleName+
		colorDim+" ("+elapsed.Truncate(time.Millisecond).String()+")"+colorReset+"\n")
}

// DryRun prints a dry-run notice instead of executing the action.
func (o *Output) DryRun(ruleName, actionType string) {
	write(o.Writer, "  "+colorYellow+"[DRY RUN]"+colorReset+" would execute: "+ruleName+" ("+actionType+")\n")
}

// Verbose prints filtered events that would otherwise be hidden.
func (o *Output) Verbose(ev watcher.Event, reason string) {
	ts := time.Now().Format("15:04:05")

	write(o.Writer, colorDim+ts+" [filtered] "+string(ev.Type)+" "+ev.Path+" ("+reason+")"+colorReset+"\n")
}

// Shutdown prints a clean exit message.
func (o *Output) Shutdown() {
	write(o.Writer, "\n"+colorDim+"  Shutting down..."+colorReset+"\n")
}

func colorForEvent(t watcher.EventType) string {
	switch t {
	case watcher.Create:
		return colorGreen
	case watcher.Modify:
		return colorYellow
	case watcher.Delete:
		return colorRed
	case watcher.Rename:
		return colorBlue
	default:
		return colorReset
	}
}

func write(w io.Writer, s string) {
	_, _ = io.WriteString(w, s)
}
