// Package action implements the action types triggered by rule matches.
package action

import (
	"time"

	"github.com/devaloi/watchdog/internal/watcher"
)

// Action is the interface for all executable actions.
type Action interface {
	Execute(ev watcher.Event) error
}

// TemplateData is passed to command and log templates.
type TemplateData struct {
	Path  string
	Event string
	Dir   string
	Name  string
	Time  string
}

// NewTemplateData builds template data from a watcher event.
func NewTemplateData(ev watcher.Event) TemplateData {
	return TemplateData{
		Path:  ev.Path,
		Event: string(ev.Type),
		Dir:   ev.Dir,
		Name:  ev.Name,
		Time:  time.Now().Format(time.RFC3339),
	}
}
