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

🚧 **Currently in development** - Phase 1 (MVP)

See [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for detailed roadmap.

## Planned Usage

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

| Tool | Scope | Status |
|------|-------|--------|
| [cctx](https://github.com/nwiizo/cctx) | settings.json only | Active |
| Shell aliases | Full config dir | Manual |
| **claudectx** | Comprehensive, safe | In development |

## Architecture

Profiles are stored in `~/.claude/profiles/`:

```
~/.claude/
├── .claudectx-current      # Current profile tracker
├── .claudectx-previous     # Previous profile (for toggle)
├── profiles/
│   ├── default/
│   │   ├── settings.json
│   │   ├── auth.json
│   │   ├── mcp-servers.json
│   │   └── CLAUDE.md
│   ├── work/
│   └── personal/
└── settings.json           # Active config (managed by claudectx)
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

- [x] Phase 1: Core MVP (in progress)
  - [x] Project setup
  - [ ] Profile storage
  - [ ] Switch/list/create/delete commands
- [ ] Phase 2: Polish & Safety
  - [ ] Validation & backups
  - [ ] Rollback on failure
  - [ ] Color output
- [ ] Phase 3: Enhancement
  - [ ] fzf integration
  - [ ] Shell completion
  - [ ] Export/import
- [ ] Phase 4: Distribution
  - [ ] Homebrew formula
  - [ ] Release automation

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
