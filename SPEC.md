# G10: watchdog — Configurable File System Watcher

**Catalog ID:** G10 | **Size:** S | **Language:** Go
**Repo name:** `watchdog`
**One-liner:** A file system watcher that triggers configurable actions on file changes — supports glob patterns, debouncing, command execution, webhooks, and YAML-based rule configuration.

---

## Why This Stands Out

- **Beyond raw fsnotify** — glob filtering, debouncing, ignore patterns, recursive watching with a clean rule engine
- **Configurable actions** — shell commands, HTTP webhooks, structured logging — all from YAML rules
- **Debouncing** — configurable delay to batch rapid file saves into a single action trigger
- **Glob pattern matching** — watch `**/*.go`, `src/**/*.ts`, ignore `.git/`, `node_modules/`
- **CLI with live output** — colorized real-time event log, clean startup banner, graceful Ctrl+C
- **Production patterns** — graceful shutdown, structured logging, YAML config, signal handling
- **Zero external deps** beyond fsnotify — all pattern matching, debouncing, and execution built from stdlib
- **Useful as a real tool** — can replace nodemon, entr, or watchexec for simple use cases

---

## Architecture

```
watchdog/
├── cmd/
│   └── watchdog/
│       └── main.go              # CLI entry: parse args, load config, run watcher
├── internal/
│   ├── config/
│   │   ├── config.go            # YAML config parsing, validation
│   │   └── config_test.go
│   ├── watcher/
│   │   ├── watcher.go           # Core watcher: fsnotify + recursive dir walking
│   │   ├── watcher_test.go
│   │   └── debounce.go          # Debounce timer per file path
│   ├── matcher/
│   │   ├── matcher.go           # Glob pattern matching + ignore filtering
│   │   └── matcher_test.go
│   ├── action/
│   │   ├── action.go            # Action interface
│   │   ├── command.go           # Shell command execution
│   │   ├── webhook.go           # HTTP POST webhook
│   │   ├── log.go               # Structured log output
│   │   ├── command_test.go
│   │   └── webhook_test.go
│   ├── rule/
│   │   ├── engine.go            # Rule engine: match event → trigger actions
│   │   └── engine_test.go
│   └── display/
│       └── output.go            # Colorized terminal output, event formatting
├── watchdog.yaml                # Example config file
├── testdata/
│   ├── watchdog.yaml            # Test config fixtures
│   └── sample/                  # Sample directory tree for watcher tests
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── .golangci.yml
├── LICENSE
└── README.md
```

---

## Config File Format

```yaml
# watchdog.yaml
global:
  debounce: 500ms
  ignore:
    - .git
    - node_modules
    - "*.tmp"
    - "**/*.swp"

rules:
  - name: "Go rebuild"
    watch:
      - "**/*.go"
    events: [create, modify]
    action:
      type: command
      command: "go build ./..."
      dir: "."

  - name: "CSS reload"
    watch:
      - "assets/**/*.css"
    events: [modify]
    debounce: 1s
    action:
      type: webhook
      url: "http://localhost:3000/reload"
      method: POST

  - name: "Log all changes"
    watch:
      - "**/*"
    events: [create, modify, delete, rename]
    action:
      type: log
      format: "[{{.Time}}] {{.Event}} {{.Path}}"
```

---

## Event Types

| Event | Trigger |
|-------|---------|
| `create` | New file or directory created |
| `modify` | File content modified |
| `delete` | File or directory removed |
| `rename` | File or directory renamed |

---

## CLI Usage

```
watchdog                        # Use ./watchdog.yaml
watchdog -c myconfig.yaml       # Custom config file
watchdog -w "**/*.go" -x "go test ./..."   # Quick mode (no config file)
watchdog --dry-run              # Show what would trigger without executing
watchdog --verbose              # Show all events including filtered ones
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `watchdog.yaml` | Config file path |
| `-w, --watch` | — | Quick mode: glob pattern to watch |
| `-x, --exec` | — | Quick mode: command to execute |
| `-d, --debounce` | `500ms` | Debounce delay |
| `--dry-run` | `false` | Preview mode, no actions executed |
| `--verbose` | `false` | Show all events including ignored |

---

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go 1.26 |
| File watching | `github.com/fsnotify/fsnotify` |
| YAML config | `gopkg.in/yaml.v3` |
| Glob matching | `path/filepath` (stdlib) + custom `**` support |
| HTTP client | stdlib `net/http` |
| CLI flags | stdlib `flag` |
| Terminal color | ANSI escape codes (no deps) |
| Testing | stdlib `testing` |
| Linting | golangci-lint |

---

## Phased Build Plan

### Phase 1: Core Watcher & Matching

**1.1 — Project setup**
- `go mod init github.com/devaloi/watchdog`
- Dependencies: `github.com/fsnotify/fsnotify`, `gopkg.in/yaml.v3`
- Create directory structure, Makefile, .gitignore, .golangci.yml
- Add LICENSE (MIT)

**1.2 — Glob matcher**
- `Match(pattern, path) bool` — support `*`, `**`, `?`, `{a,b}`
- `**` matches across directory boundaries (`**/*.go` matches `cmd/server/main.go`)
- Ignore list: if any ignore pattern matches, exclude the path
- Table-driven tests: exact, wildcard, doublestar, ignore, edge cases

**1.3 — Core watcher**
- Wrap fsnotify: create watcher, add directories recursively
- Walk directory tree, add each dir to fsnotify
- Watch for new directories (auto-add on create)
- Handle fsnotify events: translate to internal event types (create, modify, delete, rename)
- Tests: create temp dir, write files, verify events received

**1.4 — Debouncer**
- Per-path debounce timer: reset on each event, fire after delay
- Configurable delay (global default + per-rule override)
- If same path fires 10 times in 100ms with 500ms debounce → action fires once
- Tests: rapid events debounced to one, different paths debounce independently

### Phase 2: Config & Rule Engine

**2.1 — Config parsing**
- YAML config struct: global settings, rules array
- Rule struct: name, watch patterns, event types, debounce override, action config
- Action config: type (command/webhook/log) + type-specific fields
- Validation: at least one rule, valid patterns, valid action types
- Tests: parse valid config, missing fields, invalid values

**2.2 — Rule engine**
- `Evaluate(event) []Action` — for each rule, check if event matches patterns and event type
- Apply ignore filters before rule matching
- Return list of actions to execute
- Tests: matching rules, non-matching, multiple rules match same event

### Phase 3: Actions

**3.1 — Action interface**
```go
type Action interface {
    Execute(event Event) error
}
```

**3.2 — Command action**
- Execute shell command via `os/exec`
- Template variables in command: `{{.Path}}`, `{{.Event}}`, `{{.Dir}}`, `{{.Name}}`
- Working directory configuration
- Capture stdout/stderr, log output
- Kill previous command if still running (restart on change)
- Tests: command executes, template substitution, working dir

**3.3 — Webhook action**
- HTTP POST to configured URL
- JSON body: `{"path": "...", "event": "modify", "time": "..."}`
- Configurable method, headers, timeout
- Tests: mock HTTP server receives correct payload

**3.4 — Log action**
- Configurable format string with template variables
- Output to stdout or file
- Tests: format string produces expected output

### Phase 4: CLI & Polish

**4.1 — CLI entry point**
- Parse flags: config path, quick mode (--watch + --exec), dry-run, verbose
- Load config or build config from quick mode flags
- Start watcher, run until Ctrl+C
- Graceful shutdown: stop watcher, drain pending actions, exit clean

**4.2 — Live terminal output**
- Colorized event log: green for create, yellow for modify, red for delete, blue for rename
- Startup banner showing watched paths and rules
- Action execution output with timing
- Dry-run mode: show "[DRY RUN] would execute: ..." instead of running

**4.3 — Example config**
- `watchdog.yaml` in repo root: realistic example with Go rebuild, CSS reload, log all

**4.4 — README**
- Badges, install (`go install`), quick start
- Config file format reference
- CLI flags reference
- Event types table
- Action types with examples
- Comparison: when to use watchdog vs nodemon, entr, watchexec

**4.5 — Final checks**
- `go build ./...` clean
- `go test -race ./...` all pass
- `golangci-lint run` clean
- Manual test: run watchdog, edit files, verify actions fire
- Fresh clone → build → run works

---

## Commit Plan

1. `feat: scaffold project with glob matcher and tests`
2. `feat: add core file watcher with recursive directory walking`
3. `feat: add per-path debouncer with configurable delay`
4. `feat: add YAML config parsing and validation`
5. `feat: add rule engine with pattern and event type matching`
6. `feat: add command action with template variables`
7. `feat: add webhook and log actions`
8. `feat: add CLI with flags, live output, and graceful shutdown`
9. `feat: add dry-run and verbose modes`
10. `test: add integration tests for watcher → rule → action pipeline`
11. `docs: add README with config reference and usage guide`
12. `chore: final lint pass and cleanup`
