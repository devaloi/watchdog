// Package rule implements the event-to-action matching engine.
package rule

import (
	"slices"

	"github.com/devaloi/watchdog/internal/config"
	"github.com/devaloi/watchdog/internal/matcher"
	"github.com/devaloi/watchdog/internal/watcher"
)

// Match represents a rule that matched an event, along with the action config.
type Match struct {
	RuleName string
	Action   config.Action
}

// Engine evaluates file system events against configured rules.
type Engine struct {
	rules         []config.Rule
	globalIgnores []string
}

// NewEngine creates an Engine from a parsed config.
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		rules:         cfg.Rules,
		globalIgnores: cfg.Global.Ignore,
	}
}

// Evaluate checks the event against all rules and returns matching actions.
func (e *Engine) Evaluate(ev watcher.Event) []Match {
	// Check global ignore patterns first
	for _, ign := range e.globalIgnores {
		if matcher.MatchPattern(ign, ev.Path) {
			return nil
		}
	}

	var matches []Match

	for _, r := range e.rules {
		if !e.matchesRule(r, ev) {
			continue
		}

		matches = append(matches, Match{
			RuleName: r.Name,
			Action:   r.Action,
		})
	}

	return matches
}

func (e *Engine) matchesRule(r config.Rule, ev watcher.Event) bool {
	if len(r.Events) > 0 && !containsEvent(r.Events, ev.Type) {
		return false
	}

	for _, pattern := range r.Watch {
		if matcher.MatchPattern(pattern, ev.Path) {
			return true
		}
	}

	return false
}

func containsEvent(events []string, t watcher.EventType) bool {
	return slices.Contains(events, string(t))
}
