// Package matcher provides glob pattern matching with ** (doublestar) support.
package matcher

import (
	"path/filepath"
	"strings"
)

// Matcher evaluates file paths against include and ignore glob patterns.
type Matcher struct {
	includes []string
	ignores  []string
}

// New creates a Matcher with the given include and ignore patterns.
func New(includes, ignores []string) *Matcher {
	return &Matcher{
		includes: includes,
		ignores:  ignores,
	}
}

// Match returns true if the path matches any include pattern
// and does not match any ignore pattern.
func (m *Matcher) Match(path string) bool {
	path = filepath.ToSlash(path)

	for _, ign := range m.ignores {
		if matchGlob(ign, path) {
			return false
		}
	}

	for _, inc := range m.includes {
		if matchGlob(inc, path) {
			return true
		}
	}

	return false
}

// MatchPattern checks a single pattern against a path.
func MatchPattern(pattern, path string) bool {
	return matchGlob(pattern, filepath.ToSlash(path))
}

// matchGlob matches a glob pattern supporting *, **, and ? wildcards.
// ** matches zero or more directory segments.
func matchGlob(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Handle ignore patterns that are bare names (e.g., ".git", "node_modules")
	if !strings.Contains(pattern, "/") && !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
		for seg := range strings.SplitSeq(path, "/") {
			if seg == pattern {
				return true
			}
		}

		return false
	}

	return globMatch(pattern, path)
}

func globMatch(pattern, name string) bool {
	// Split both pattern and path into segments
	patParts := strings.Split(pattern, "/")
	nameParts := strings.Split(name, "/")

	return matchSegments(patParts, nameParts)
}

func matchSegments(patParts, nameParts []string) bool {
	pi, ni := 0, 0

	for pi < len(patParts) && ni < len(nameParts) {
		if patParts[pi] == "**" {
			for skip := ni; skip <= len(nameParts); skip++ {
				if matchSegments(patParts[pi+1:], nameParts[skip:]) {
					return true
				}
			}

			return false
		}

		if !matchSegment(patParts[pi], nameParts[ni]) {
			return false
		}

		pi++
		ni++
	}

	// Consume trailing ** patterns
	for pi < len(patParts) && patParts[pi] == "**" {
		pi++
	}

	return pi == len(patParts) && ni == len(nameParts)
}

func matchSegment(pattern, name string) bool {
	// Handle brace expansion: {a,b}
	if idx := strings.Index(pattern, "{"); idx >= 0 {
		end := strings.Index(pattern[idx:], "}")
		if end >= 0 {
			prefix := pattern[:idx]

			suffix := pattern[idx+end+1:]
			for alt := range strings.SplitSeq(pattern[idx+1:idx+end], ",") {
				if matchSegment(prefix+alt+suffix, name) {
					return true
				}
			}

			return false
		}
	}

	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}

	return matched
}
