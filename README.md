# claudectx

Fast way to switch between Claude Code configuration profiles.

Inspired by [kubectx](https://github.com/ahmetb/kubectx).

## Why?

Claude Code doesn't have built-in profile or context switching. If you work with:
- Multiple Claude accounts (work vs personal)
- Different API providers (Anthropic, Bedrock, custom endpoints)
- Client-specific configurations
- Different tool permissions and MCP servers

...then you need `claudectx`.

## Features

- **Simple commands**: `claudectx work`, `claudectx -` (toggle)
- **Comprehensive management**: Settings, auth, MCP servers, CLAUDE.md files
- **Safe switching**: Automatic backups and validation
- **Cross-platform**: Single binary for macOS, Linux, Windows
- **No dependencies**: Just a single Go binary

## Status

✅ **Phase 2 Complete** - Production ready!

- ✅ Phase 1: Core MVP - All commands working
- ✅ Phase 2: Polish & Safety - Backups, validation, colored output
- 🔮 Phase 3: Enhancement - fzf, shell completion (planned)
- 🔮 Phase 4: Distribution - Homebrew, releases (planned)

See [PHASE2_COMPLETE.md](PHASE2_COMPLETE.md) for Phase 2 details.

## Installation

```bash
# Quick install (no sudo required)
cd /Users/johnfox/Documents/claudectx
make install-user

# Or install system-wide
make install

# See INSTALL.md for detailed instructions
```

## Usage

```bash
# List all profiles
claudectx

# Switch to a profile
claudectx work

# Toggle to previous profile
claudectx -

# Show current profile
claudectx -c

# Create new profile from current config
claudectx -n personal

# Delete profile
claudectx -d old-profile
```

## Installation

Coming soon! Will support:
- Homebrew: `brew install claudectx`
- Direct download: Binary releases for all platforms
- Build from source: `go install github.com/johnfox/claudectx@latest`

## What claudectx Manages

- `~/.claude/settings.json` - User-level settings
- `~/.claude/CLAUDE.md` - Global instructions
- Authentication configuration (API keys, base URLs)
- MCP servers (user scope)
- Environment variables

## What it Doesn't Manage

- Project-level settings (stays in your projects)
- OAuth session tokens (too risky)
- Conversation history
- Project state

## Comparison with Existing Tools

| Tool | Scope | Backups | Validation | Status |
|------|-------|---------|------------|--------|
| [cctx](https://github.com/nwiizo/cctx) | settings.json only | ❌ | ❌ | Active |
| Shell aliases | Full config dir | ❌ | ❌ | Manual |
| **claudectx** | Comprehensive | ✅ Auto | ✅ Full | **Production ready** |

## Architecture

Profiles are stored in `~/.claude/profiles/`:

```
~/.claude/
├── .claudectx-current      # Current profile tracker
├── .claudectx-previous     # Previous profile (for toggle)
├── profiles/               # Profile storage
│   ├── default/
│   │   ├── settings.json
│   │   └── CLAUDE.md
│   ├── work/
│   └── personal/
├── backups/                # Automatic backups (Phase 2)
│   ├── backup-1234567890/
│   │   ├── settings.json
│   │   └── CLAUDE.md
│   └── backup-1234567891/
├── settings.json           # Active config (managed by claudectx)
└── CLAUDE.md               # Active instructions (managed by claudectx)
```

## Development

```bash
# Build
go build -o claudectx

# Run
./claudectx --help

# Test
go test ./...
```

## Roadmap

- [x] Phase 1: Core MVP ✅
  - [x] Project setup
  - [x] Profile storage
  - [x] All commands (switch/list/create/delete/toggle/current)
  - [x] 44 tests, all passing
- [x] Phase 2: Polish & Safety ✅
  - [x] Automatic backups
  - [x] Validation & rollback on failure
  - [x] Colored output
  - [x] Better error messages
  - [x] 68 tests, all passing
- [ ] Phase 3: Enhancement (optional)
  - [ ] fzf integration for interactive selection
  - [ ] Shell completion (bash/zsh/fish)
  - [ ] Export/import profiles
  - [ ] Health checks
- [ ] Phase 4: Distribution
  - [ ] Homebrew formula
  - [ ] Release automation (GoReleaser)
  - [ ] Public release

## Contributing

Not accepting contributions yet - project is in early development.

## License

MIT License - see [LICENSE](LICENSE)

## Acknowledgments

- [kubectx](https://github.com/ahmetb/kubectx) - For the excellent UX patterns
- [cctx](https://github.com/nwiizo/cctx) - First attempt at Claude context switching
- Claude Code community for identifying the need

## Author

John Fox - [@johnfox](https://github.com/johnfox)
