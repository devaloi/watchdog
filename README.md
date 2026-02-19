# watchdog

A configurable file system watcher that triggers actions on file changes — supports glob patterns, debouncing, command execution, webhooks, and YAML-based rule configuration.

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features

- **Glob pattern matching** with `**` (doublestar) support — `**/*.go` matches files at any depth
- **Per-path debouncing** — rapid saves trigger a single action
- **YAML rule engine** — configure watch patterns, event filters, and actions
- **Command execution** with template variables (`{{.Path}}`, `{{.Event}}`, `{{.Dir}}`, `{{.Name}}`)
- **Webhook delivery** — HTTP POST with JSON event payload
- **Structured logging** — configurable format templates
- **Recursive watching** — automatically watches new subdirectories
- **Graceful shutdown** — clean Ctrl+C handling, no zombie processes
- **Dry-run mode** — preview what would trigger without executing
- **Zero external deps** beyond fsnotify and yaml.v3

## Install

```bash
go install github.com/devaloi/watchdog/cmd/watchdog@latest
```

Or build from source:

```bash
git clone https://github.com/devaloi/watchdog.git
cd watchdog
make build
```

## Quick Start

### Config file mode

Create a `watchdog.yaml`:

```yaml
global:
  debounce: 500ms
  ignore:
    - .git
    - node_modules

rules:
  - name: "Go rebuild"
    watch:
      - "**/*.go"
    events: [create, modify]
    action:
      type: command
      command: "go build ./..."
      dir: "."
```

Run:

```bash
watchdog
```

### Quick mode (no config file)

```bash
watchdog -w "**/*.go" -x "go test ./..."
```

## Usage

```
watchdog                        # Use ./watchdog.yaml
watchdog -c myconfig.yaml       # Custom config file
watchdog -w "**/*.go" -x "go test ./..."   # Quick mode
watchdog --dry-run              # Preview mode
watchdog --verbose              # Show filtered events
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

## Configuration

### Config File Format

```yaml
global:
  debounce: 500ms          # Default debounce delay
  ignore:                  # Global ignore patterns
    - .git
    - node_modules
    - "*.tmp"
    - "**/*.swp"

rules:
  - name: "Rule name"      # Display name
    watch:                  # Glob patterns to match
      - "**/*.go"
    events: [create, modify]  # Event type filter
    debounce: 1s           # Per-rule debounce override
    action:
      type: command        # Action type
      command: "go build"  # Type-specific config
      dir: "."
```

### Event Types

| Event | Trigger |
|-------|---------|
| `create` | New file or directory created |
| `modify` | File content modified |
| `delete` | File or directory removed |
| `rename` | File or directory renamed |

### Action Types

#### Command

Runs a shell command. Kills the previous instance if still running.

```yaml
action:
  type: command
  command: "go build ./..."
  dir: "."
```

Template variables: `{{.Path}}`, `{{.Event}}`, `{{.Dir}}`, `{{.Name}}`, `{{.Time}}`

#### Webhook

Sends an HTTP request with a JSON event payload.

```yaml
action:
  type: webhook
  url: "http://localhost:3000/reload"
  method: POST
  headers:
    Authorization: "Bearer token"
  timeout: 5s
```

Payload:

```json
{"path": "src/main.go", "event": "modify", "time": "2025-01-01T12:00:00Z"}
```

#### Log

Writes formatted log lines.

```yaml
action:
  type: log
  format: "[{{.Time}}] {{.Event}} {{.Path}}"
```

## Glob Patterns

| Pattern | Matches |
|---------|---------|
| `*.go` | Go files in current directory |
| `**/*.go` | Go files at any depth |
| `src/**/*.ts` | TypeScript files under `src/` |
| `*.{go,rs}` | Go and Rust files |
| `?_test.go` | Single-char prefix test files |

Ignore patterns can be bare names (e.g., `.git`, `node_modules`) which match any path segment.

## Tech Stack

| Component | Choice |
|-----------|--------|
| Language | Go 1.26 |
| File watching | [fsnotify](https://github.com/fsnotify/fsnotify) |
| YAML config | [yaml.v3](https://gopkg.in/yaml.v3) |
| Glob matching | Custom with `**` support |
| CLI flags | stdlib `flag` |
| Terminal color | ANSI escape codes |

## Development

```bash
make build     # Build binary
make test      # Run tests with race detector
make lint      # Run golangci-lint
make all       # Lint + test + build
```

### Prerequisites

- Go 1.26+
- golangci-lint v2 (optional, for linting)

## Comparison

| Feature | watchdog | nodemon | entr | watchexec |
|---------|----------|---------|------|-----------|
| Glob `**` patterns | ✅ | ✅ | ❌ | ✅ |
| Per-path debouncing | ✅ | ✅ | ❌ | ✅ |
| YAML config | ✅ | JSON | ❌ | ❌ |
| Multiple rules | ✅ | ❌ | ❌ | ❌ |
| Webhooks | ✅ | ❌ | ❌ | ❌ |
| Single binary | ✅ | ❌ | ✅ | ✅ |
| No runtime deps | ✅ | Node.js | ❌ | ❌ |

## License

[MIT](LICENSE)
