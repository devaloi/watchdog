# Build watchdog — Configurable File System Watcher

You are building a **portfolio project** for a Senior AI Engineer's public GitHub. It must be impressive, clean, and production-grade. Read these docs before writing any code:

1. **`G10-go-file-watcher.md`** — Complete project spec: architecture, phases, glob matching, debouncing, rule engine, action system, CLI, commit plan. This is your primary blueprint. Follow it phase by phase.
2. **`github-portfolio.md`** — Portfolio goals and Definition of Done (Level 1 + Level 2). Understand the quality bar.
3. **`github-portfolio-checklist.md`** — Pre-publish checklist. Every item must pass before you're done.

---

## Instructions

### Read first, build second
Read all three docs completely before writing a single line of code. Understand the recursive file watcher, glob pattern matching with `**` support, the per-path debouncer, the YAML-driven rule engine, and the action system (command, webhook, log).

### Follow the phases in order
The project spec has 4 phases. Do them in order:
1. **Core Watcher & Matching** — project setup, glob matcher with `**` support, core fsnotify watcher with recursive dir walking, per-path debouncer
2. **Config & Rule Engine** — YAML config parsing and validation, rule engine that evaluates events against patterns and triggers actions
3. **Actions** — action interface, shell command execution with template variables, HTTP webhook delivery, structured log output
4. **CLI & Polish** — CLI entry with flags, colorized live terminal output, dry-run mode, example config, comprehensive README

### Commit frequently
Follow the commit plan in the spec. Use **conventional commits**. Each commit should be a logical unit.

### Quality non-negotiables
- **Glob matching with `**` is critical.** `**/*.go` must match files in any subdirectory depth. Standard `filepath.Match` doesn't support `**` — you need custom logic. Test extensively.
- **Debouncing must be per-path.** Rapid saves to `main.go` should trigger one action. A save to `main.go` and `config.yaml` should trigger two separate actions (each debounced independently).
- **Command action kills previous.** If a rebuild command is still running when a new change arrives, kill it and restart. This prevents zombie processes.
- **Graceful shutdown.** On SIGINT/SIGTERM: stop the watcher, drain pending debounce timers, kill running commands, exit with code 0. No goroutine leaks.
- **YAML config is the primary interface.** The config file format must be clean, well-documented, and validated on load with clear error messages.
- **Template variables in commands.** `{{.Path}}`, `{{.Event}}`, `{{.Dir}}`, `{{.Name}}` must all work in command strings.
- **Lint clean.** `golangci-lint run` and `go vet` must pass with zero warnings.
- **No Docker.** Just `go build` and `go install`. This is a CLI tool.

### What NOT to do
- Don't use any file watcher framework beyond fsnotify. Build the glob matching, debouncing, and rule engine yourself.
- Don't use cobra or any CLI framework. Stdlib `flag` is sufficient for this tool.
- Don't shell out to `find` or `ls` for directory walking. Use `filepath.WalkDir`.
- Don't leave zombie processes. Command actions must track their child process and kill it on re-trigger or shutdown.
- Don't leave `// TODO` or `// FIXME` comments anywhere.
- Don't skip the debouncer tests. Timing-based tests are tricky — use short delays (10-50ms) in tests.

---

## GitHub Username

The GitHub username is **devaloi**. For Go module paths, use `github.com/devaloi/watchdog`. All internal imports must use this module path.

## Start

Read the three docs. Then begin Phase 1 from `G10-go-file-watcher.md`.
