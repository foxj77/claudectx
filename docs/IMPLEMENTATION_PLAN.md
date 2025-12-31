# claudectx Implementation Plan

## Project Overview

Build a kubectx-inspired CLI tool for managing and switching between multiple Claude Code configuration profiles. This addresses the lack of official profile/context switching in Claude Code and provides a better user experience than existing community solutions.

## Why This Project is Worthwhile

### The Problem
- Claude Code has no built-in profile/context switching
- Users need to manage multiple configurations (work/personal, different API providers, client-specific settings)
- Existing community tools are fragmented and incomplete:
  - `cctx` only manages settings.json
  - Shell-based switchers use buggy undocumented `CLAUDE_CONFIG_DIR`
  - No tool comprehensively manages auth + settings + MCP + environment variables

### The Solution
A kubectx-style tool that:
- Provides simple, intuitive commands (`claudectx work`, `claudectx -`)
- Manages complete configuration profiles (settings, auth, MCP servers, CLAUDE.md)
- Validates configurations before switching
- Works cross-platform via single Go binary
- Distributes easily via Homebrew

## Technology Choice: Go

**Rationale:**
1. **Follows kubectx's proven evolution** - bash → Go rewrite for performance and maintainability
2. **Cross-platform** - Single binary for Linux, macOS, Windows
3. **Strong JSON handling** - Built-in encoding/json, no external dependencies
4. **Easy distribution** - Homebrew, apt, single binary downloads
5. **Performance** - Fast startup critical for frequently-used CLI tools
6. **No runtime dependencies** - Unlike bash (needs jq), Python (needs interpreter)

## Architecture Design

### Directory Structure

```
~/.claude/                         # Claude's existing config directory
├── .claudectx-current            # Tracks active profile name
├── .claudectx-previous           # Enables toggle with claudectx -
├── profiles/                     # Profile storage (created by tool)
│   ├── default/
│   │   ├── settings.json         # User settings
│   │   ├── auth.json             # Authentication config
│   │   ├── mcp-servers.json      # MCP server definitions
│   │   └── CLAUDE.md             # Global instructions
│   ├── work/
│   └── personal/
└── settings.json                 # Active config (managed by tool)
```

### Profile Scope (Medium - Recommended)

**What claudectx manages:**
- `~/.claude/settings.json` - User-level settings
- `~/.claude/CLAUDE.md` - Global instructions
- Authentication configuration (OAuth/API keys)
- MCP servers (user scope from `~/.claude.json`)
- Environment variables

**What it does NOT manage:**
- Project-level settings (`.claude/settings.json`, `.claude/settings.local.json`)
- Project-specific instructions
- Full `~/.claude.json` state (too risky - contains session info, caches, etc.)

### Command Interface (kubectx-style)

```bash
# List all profiles (default action)
claudectx

# Switch to a profile
claudectx work

# Toggle to previous profile
claudectx -

# Show current profile
claudectx -c, --current

# Create new profile from current config
claudectx -n personal

# Initialize new empty profile
claudectx --init work-api

# Rename profile
claudectx work-2025=work

# Delete profile
claudectx -d old-profile

# Validate profile without switching
claudectx --validate work

# Export profile to stdout
claudectx export work

# Import profile from stdin
claudectx import personal
```

## Implementation Phases

### Phase 1: Core MVP (Days 1-2)

**Goal:** Basic profile switching functionality

**Deliverables:**
1. Go project structure
2. Profile storage and retrieval
3. Core commands: list, switch, create, delete
4. Basic config file manipulation (settings.json)
5. Current/previous profile tracking

**Files to create:**
- `main.go` - Entry point and CLI parsing
- `internal/profile/` - Profile management
- `internal/config/` - Config file manipulation
- `internal/store/` - Profile storage operations
- `cmd/` - Command implementations (list, switch, create, delete)
- `go.mod` - Dependencies
- `README.md` - Basic documentation

**Key features:**
- ✅ List profiles with current highlighted
- ✅ Switch profiles atomically
- ✅ Create profile from current config
- ✅ Delete profiles with confirmation
- ✅ Toggle with `-` between profiles
- ✅ Basic error handling

### Phase 2: Polish & Safety (Days 3-4)

**Goal:** Production-ready with safety features

**Deliverables:**
1. JSON validation before switching
2. Automatic backups
3. Rollback on failure
4. Structure preservation (formatted JSON)
5. Color output
6. Better error messages

**Files to create/modify:**
- `internal/backup/` - Backup management
- `internal/validator/` - Config validation
- `internal/printer/` - Colored output
- Tests for all packages

**Key features:**
- ✅ Validate JSON syntax before applying
- ✅ Backup current config before switch
- ✅ Rollback if switch fails
- ✅ Preserve JSON formatting and indentation
- ✅ Color-coded output (current profile highlighted)
- ✅ Comprehensive error messages

### Phase 3: Progressive Enhancement (Days 5-6)

**Goal:** Best-in-class UX

**Deliverables:**
1. fzf integration for interactive selection
2. Shell completion (bash/zsh/fish)
3. Export/import functionality
4. Health checks (API connectivity)
5. Template-based profile creation

**Files to create/modify:**
- `internal/fzf/` - Fuzzy finder integration
- `completion/` - Shell completion scripts
- `internal/health/` - Health check validators
- `internal/template/` - Profile templates

**Key features:**
- ✅ Interactive mode with fzf
- ✅ Tab completion for all commands
- ✅ Export/import profiles for sharing
- ✅ Verify API keys work before switching
- ✅ Profile templates (work, personal, custom-api)

### Phase 4: Distribution (Day 7)

**Goal:** Easy installation and discovery

**Deliverables:**
1. Homebrew formula
2. Release automation (GoReleaser)
3. Comprehensive documentation
4. Installation guide

**Files to create:**
- `homebrew/claudectx.rb` - Homebrew formula
- `.goreleaser.yml` - Release configuration
- `docs/` - User documentation
- `INSTALL.md` - Installation instructions

## Go Project Structure

```
claudectx/
├── main.go                       # Entry point, CLI arg parsing
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── README.md                     # Project documentation
├── LICENSE                       # License (MIT recommended)
├── .gitignore                    # Git ignore rules
├── Makefile                      # Build automation
│
├── cmd/                          # Command implementations
│   ├── list.go                   # List profiles
│   ├── switch.go                 # Switch to profile
│   ├── create.go                 # Create new profile
│   ├── delete.go                 # Delete profile
│   ├── current.go                # Show current profile
│   ├── rename.go                 # Rename profile
│   ├── export.go                 # Export profile
│   ├── import.go                 # Import profile
│   └── validate.go               # Validate profile
│
├── internal/                     # Private packages
│   ├── profile/                  # Profile management
│   │   ├── profile.go            # Profile struct and operations
│   │   ├── storage.go            # Profile storage backend
│   │   └── profile_test.go
│   │
│   ├── config/                   # Configuration file handling
│   │   ├── settings.go           # settings.json manipulation
│   │   ├── auth.go               # Authentication config
│   │   ├── mcp.go                # MCP server config
│   │   ├── claude_md.go          # CLAUDE.md handling
│   │   └── config_test.go
│   │
│   ├── store/                    # Storage operations
│   │   ├── store.go              # Profile store interface
│   │   ├── filesystem.go         # Filesystem implementation
│   │   └── store_test.go
│   │
│   ├── backup/                   # Backup management
│   │   ├── backup.go             # Create/restore backups
│   │   └── backup_test.go
│   │
│   ├── validator/                # Configuration validation
│   │   ├── validator.go          # JSON validation, health checks
│   │   └── validator_test.go
│   │
│   ├── printer/                  # Output formatting
│   │   ├── printer.go            # Colored output, tables
│   │   └── printer_test.go
│   │
│   ├── fzf/                      # Fuzzy finder integration
│   │   ├── fzf.go                # fzf integration
│   │   └── fzf_test.go
│   │
│   └── paths/                    # Path utilities
│       ├── paths.go              # ~/.claude path resolution
│       └── paths_test.go
│
├── completion/                   # Shell completion
│   ├── bash_completion.sh
│   ├── zsh_completion.sh
│   └── fish_completion.fish
│
├── docs/                         # Documentation
│   ├── installation.md
│   ├── usage.md
│   └── architecture.md
│
└── .goreleaser.yml              # Release automation
```

## Core Package Design

### 1. Profile Package

```go
package profile

// Profile represents a complete Claude configuration profile
type Profile struct {
    Name        string
    Settings    *Settings      // settings.json content
    Auth        *Auth          // Authentication config
    MCPServers  []MCPServer    // MCP server definitions
    ClaudeMD    string         // CLAUDE.md content
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Store manages profile persistence
type Store interface {
    List() ([]Profile, error)
    Get(name string) (*Profile, error)
    Save(profile *Profile) error
    Delete(name string) error
    Current() (*Profile, error)
    SetCurrent(name string) error
    Previous() (*Profile, error)
    SetPrevious(name string) error
}
```

### 2. Config Package

```go
package config

// Settings represents ~/.claude/settings.json
type Settings struct {
    Env         map[string]string `json:"env,omitempty"`
    Model       string           `json:"model,omitempty"`
    Permissions *Permissions     `json:"permissions,omitempty"`
    // ... other fields
}

// Auth represents authentication configuration
type Auth struct {
    Type        string // "oauth", "api_key", "custom"
    APIKey      string `json:"api_key,omitempty"`
    BaseURL     string `json:"base_url,omitempty"`
    // OAuth fields extracted from ~/.claude.json
}

// MCPServer represents an MCP server configuration
type MCPServer struct {
    Name    string
    Command string
    Args    []string
    Env     map[string]string
}
```

### 3. Switcher Logic

**Atomic switching process:**
1. Validate target profile (JSON syntax, required fields)
2. Load current configuration
3. Create backup of current config
4. Update `.claudectx-previous` with current profile name
5. Copy target profile files to active locations
6. Update `.claudectx-current` with target profile name
7. Verify switch succeeded
8. If any step fails, rollback from backup

### 4. Safety Features

**Validation checks:**
- JSON syntax validation
- Required fields present
- API endpoint reachable (optional health check)
- No conflicting configurations

**Backup strategy:**
- Create timestamped backup before every switch
- Keep last N backups (configurable, default 5)
- Easy restore command

**Rollback process:**
- Detect failed switch (file operation errors, validation failures)
- Automatically restore from backup
- Log failure reason for debugging

## Critical Files to Modify

### What claudectx Touches

**Reads:**
- `~/.claude.json` - Extract MCP servers, auth info (read-only)
- `~/.claude/settings.json` - Current settings

**Writes/Manages:**
- `~/.claude/settings.json` - Replaced during switch
- `~/.claude/CLAUDE.md` - Replaced during switch
- `~/.claude/.claudectx-current` - Current profile tracker
- `~/.claude/.claudectx-previous` - Previous profile tracker
- `~/.claude/profiles/` - Profile storage directory

**Never touches:**
- Project-level configs (`.claude/settings.json`, `.claude/settings.local.json`)
- OAuth session tokens (too risky)
- Conversation history
- Project state

## Testing Strategy

### Unit Tests
- All packages have `_test.go` files
- Test JSON marshaling/unmarshaling
- Test profile storage operations
- Test validation logic

### Integration Tests
- End-to-end switch scenarios
- Rollback on failure
- Concurrent profile operations
- Edge cases (missing files, corrupt JSON)

### Manual Testing Checklist
- [ ] Create profile from scratch
- [ ] Switch between profiles
- [ ] Toggle with `-`
- [ ] Delete profile
- [ ] Rename profile
- [ ] Export/import profile
- [ ] fzf integration
- [ ] Color output
- [ ] Shell completion
- [ ] Test on macOS, Linux, Windows

## Distribution Setup

### Homebrew Formula

```ruby
class Claudectx < Formula
  desc "Fast way to switch between Claude Code configuration profiles"
  homepage "https://github.com/yourusername/claudectx"
  url "https://github.com/yourusername/claudectx/archive/v0.1.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"claudectx"

    # Install shell completions
    bash_completion.install "completion/bash_completion.sh" => "claudectx"
    zsh_completion.install "completion/zsh_completion.sh" => "_claudectx"
    fish_completion.install "completion/fish_completion.fish"
  end

  test do
    system "#{bin}/claudectx", "--version"
  end
end
```

### GoReleaser Configuration

```yaml
project_name: claudectx

before:
  hooks:
    - go mod download

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: claudectx
    repository:
      owner: yourusername
      name: homebrew-tap
    description: "Fast way to switch between Claude Code profiles"
    homepage: "https://github.com/yourusername/claudectx"
```

## Risk Mitigation

### Potential Issues

1. **Concurrent Claude Code instances**
   - Mitigation: File locking, detect running instances

2. **Corrupt configuration files**
   - Mitigation: Validation before write, atomic operations, backups

3. **OAuth token invalidation**
   - Mitigation: Don't touch OAuth tokens, only manage settings/auth config

4. **Cross-platform path differences**
   - Mitigation: Use Go's filepath package, test on all platforms

5. **Claude Code updates breaking config format**
   - Mitigation: Version detection, graceful degradation, clear error messages

## Success Metrics

### MVP Success Criteria
- [ ] Can create profiles from current config
- [ ] Can list all profiles
- [ ] Can switch between profiles
- [ ] Claude Code works after switch
- [ ] Can toggle between two profiles with `-`
- [ ] Profiles persist across restarts

### Polish Success Criteria
- [ ] Zero failed switches (rollback works)
- [ ] Sub-second switch time
- [ ] Works on macOS, Linux, Windows
- [ ] Installs via `brew install claudectx`
- [ ] Shell completion works

### Community Success Criteria
- [ ] 100+ GitHub stars in first month
- [ ] Featured in Claude Code community resources
- [ ] Other developers contributing
- [ ] Used by consultants managing multiple clients

## Timeline Estimate

- **Phase 1 (Core MVP)**: 2 days
- **Phase 2 (Polish)**: 2 days
- **Phase 3 (Enhancement)**: 2 days
- **Phase 4 (Distribution)**: 1 day
- **Total**: ~7 days of focused development

## Next Immediate Steps

1. **Initialize Go project** with proper module structure
2. **Create GitHub repo** (private initially)
3. **Implement paths package** - Get Claude config directory location working
4. **Implement profile storage** - Basic create/read/list profiles
5. **Implement switch command** - The core functionality
6. **Test manually** - Verify Claude Code works after switch

## Open Questions to Resolve

1. **MCP Server Handling**: Should we extract MCP servers from `~/.claude.json` or manage them separately?
   - Recommendation: Extract from `.claude.json` into separate `mcp-servers.json` in profile

2. **Auth Storage**: How to safely store OAuth tokens vs API keys?
   - Recommendation: Don't store OAuth tokens, only API keys and base URLs

3. **Migration Path**: How to help users migrate from existing solutions (cctx, shell aliases)?
   - Recommendation: Provide import command that detects and migrates existing setups

4. **Default Profile**: What happens on first run with no profiles?
   - Recommendation: Auto-create "default" profile from current config

## References

- [kubectx source code](https://github.com/ahmetb/kubectx) - Architecture inspiration
- [Claude Code settings docs](https://code.claude.com/docs/en/settings) - Config format
- [cctx implementation](https://github.com/nwiizo/cctx) - Existing community tool
- [Go CLI best practices](https://github.com/spf13/cobra) - CLI framework options
