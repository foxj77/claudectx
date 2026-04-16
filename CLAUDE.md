# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/foxj77/claudectx/badge)](https://securityscorecards.dev/viewer/?uri=github.com/foxj77/claudectx)

## Project Overview

**claudectx** is a Go CLI tool for fast, safe profile switching in Claude Code. It manages configuration profiles stored in `~/.claude/profiles/`, enabling switching between different settings, CLAUDE.md files, and MCP servers.

## Build & Test Commands

```bash
make build           # Build binary (outputs ./claudectx)
make install-user    # Install to ~/go/bin (recommended, no sudo)
make test            # Run all unit tests
make test-coverage   # Run tests with coverage
go test ./... -v     # Verbose test output
go vet ./...         # Static analysis
```

## Architecture

### Core Data Flow
```
main.go (CLI argument routing)
  → cmd/*.go (command implementations)
    → internal/store (profile CRUD)
      → internal/config, internal/mcpconfig, internal/profile (data structures)
        → ~/.claude/ (filesystem)
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `internal/store` | Profile persistence, tracks current/previous profile |
| `internal/config` | Settings struct (`Env`, `Model`, `Permissions`), JSON I/O — unknown fields are preserved transparently via `extras map[string]json.RawMessage` |
| `internal/mcpconfig` | MCP server configs from `~/.claude.json` |
| `internal/paths` | Centralized path resolution for Claude files |
| `internal/printer` | Colored terminal output (respects `NO_COLOR`) |
| `internal/backup` | Timestamped backups before switching |
| `internal/validator` | Settings and CLAUDE.md validation |

### Key Paths
- `~/.claude/settings.json` - Active settings
- `~/.claude/CLAUDE.md` - Global instructions
- `~/.claude.json` - MCP server configurations
- `~/.claude/profiles/<name>/` - Stored profiles
- `~/.claude/.claudectx-current` - Current profile tracker

## Coding Conventions

### Error Handling
```go
if err != nil {
    return fmt.Errorf("failed to load profile: %w", err)
}
```

### User Output
```go
printer.Success("Created profile %q", name)  // Green
printer.Warning("Warning: %v", err)          // Yellow
printer.Error("Error: %v", err)              // Red, stderr
printer.Info("  Model: %s", model)           // Blue
```

### Store Pattern
```go
s, err := store.NewStore()
if s.Exists(name) { /* ... */ }
prof, err := s.Load(name)
err = s.Save(prof)
```

## Safety Requirements

When modifying profile switching:
1. **Backup first**: Call `backup.NewManager().Create()` before switching
2. **Validate**: Use `validator.ValidateSettings()` before applying
3. **Atomic writes**: Write to temp file, then rename
4. **Auto-sync**: Save changes to current profile before switching
5. **Rollback**: Restore from backup if switch fails

## Don'ts

- Don't hardcode paths - use `internal/paths`
- Don't print directly to stdout - use `internal/printer`
- Don't skip validation before profile operations
- Don't modify profiles without creating backups
- Don't add external dependencies without strong justification
- Don't add new top-level fields to `settings.json` by expanding the `Settings` struct — unknown fields are already preserved by the `extras` mechanism, so this is only needed when a field requires programmatic access from Go code

## Commands Reference

| Command | File |
|---------|------|
| `claudectx` (interactive) | `cmd/list_interactive.go` |
| `claudectx <name>` (switch) | `cmd/switch.go` |
| `claudectx -n <name>` | `cmd/create.go` |
| `claudectx -l` | `cmd/list.go` |
| `claudectx -c` | `cmd/current.go` |
| `claudectx -d <name>` | `cmd/delete.go` |
| `claudectx -` (toggle) | `cmd/toggle.go` |
| `claudectx export/import` | `cmd/export.go`, `cmd/import.go` |
| `claudectx health` | `cmd/health.go` |
| `claudectx sync` | `cmd/sync.go` |
