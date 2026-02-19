package matcher

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// Exact matches
		{"exact file", "main.go", "main.go", true},
		{"exact mismatch", "main.go", "other.go", false},

		// Single star
		{"star go files", "*.go", "main.go", true},
		{"star no match subdir", "*.go", "cmd/main.go", false},

		// Question mark
		{"question mark", "?.go", "a.go", true},
		{"question mark no match", "?.go", "ab.go", false},

		// Double star
		{"doublestar go files", "**/*.go", "main.go", true},
		{"doublestar nested", "**/*.go", "cmd/server/main.go", true},
		{"doublestar deep nested", "**/*.go", "a/b/c/d/e.go", true},
		{"doublestar no match ext", "**/*.go", "cmd/server/main.rs", false},
		{"doublestar prefix", "src/**/*.ts", "src/app/index.ts", true},
		{"doublestar prefix nested", "src/**/*.ts", "src/a/b/c.ts", true},
		{"doublestar prefix no match", "src/**/*.ts", "lib/index.ts", false},

		// Brace expansion
		{"brace expansion go", "*.{go,rs}", "main.go", true},
		{"brace expansion rs", "*.{go,rs}", "lib.rs", true},
		{"brace expansion no match", "*.{go,rs}", "app.ts", false},

		// Directory patterns
		{"dir pattern", "cmd/**", "cmd/main.go", true},
		{"dir pattern nested", "cmd/**", "cmd/server/main.go", true},

		// Bare name patterns (ignore-style)
		{"bare name root", ".git", ".git/config", true},
		{"bare name nested", "node_modules", "some/path/node_modules/pkg.js", true},
		{"bare name no match", ".git", "legit/file.go", false},

		// Wildcard ignore patterns
		{"wildcard ignore", "*.tmp", "cache.tmp", true},
		{"doublestar ignore", "**/*.swp", "src/deep/file.swp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatcher(t *testing.T) {
	m := New(
		[]string{"**/*.go", "**/*.yaml"},
		[]string{".git", "node_modules", "*.tmp"},
	)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"matches go file", "cmd/main.go", true},
		{"matches yaml", "config.yaml", true},
		{"ignores git dir", ".git/config", false},
		{"ignores node_modules", "node_modules/pkg/index.js", false},
		{"ignores tmp files", "cache.tmp", false},
		{"no match ts file", "src/app.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.Match(tt.path)
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatcherEmptyIncludes(t *testing.T) {
	m := New(nil, nil)
	if m.Match("anything.go") {
		t.Error("empty matcher should not match anything")
	}
}

func TestDoubleStarMatchesZeroSegments(t *testing.T) {
	if !MatchPattern("**/*.go", "main.go") {
		t.Error("** should match zero directory segments")
	}
}
