# Changelog

All notable changes to claudectx will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - TBD

### Added
- Initial public release
- Core profile management (create, switch, delete, list, toggle)
- Interactive profile selector with arrow key navigation
- Automatic backups before every switch
- Validation and rollback on failure
- Colored terminal output
- Export/import profiles to JSON
- Health checks for profile validation
- Shell completion for bash, zsh, and fish
- Comprehensive test suite (85+ tests)

### Features
- **Profile Management**: Create, switch, delete, and list Claude Code profiles
- **Interactive Mode**: Arrow key navigation to select profiles
- **Safety**: Automatic backups, validation, and rollback on errors
- **Import/Export**: Share profiles with teammates or backup configurations
- **Health Checks**: Validate profile settings before use
- **Shell Completion**: Tab completion for all major shells

## [0.4.0] - 2025-12-31

### Added
- Interactive profile selector with arrow key navigation
- `-l/--list` flag for simple list output (scripting-friendly)
- Automatic fallback to simple list when not in TTY

### Changed
- Default behavior now shows interactive selector instead of simple list
- Improved terminal rendering with proper line breaks

## [0.3.0] - 2025-12-31

### Added
- Export/import functionality with JSON format versioning
- Shell completion for bash, zsh, and fish
- Health check system for profile validation
- `claudectx export` command
- `claudectx import` command
- `claudectx health` command
- Profile validation warnings and errors

### Changed
- Updated help text with new commands
- Improved documentation

## [0.2.0] - 2025-12-31

### Added
- Automatic backups before every profile switch
- JSON validation before applying profiles
- Automatic rollback on switch failure
- Colored terminal output (green/red/yellow/blue)
- Comprehensive validation system
- Better error messages
- NO_COLOR environment variable support

### Changed
- All commands now include colored output
- Switch command now creates backups automatically
- Enhanced error handling with detailed messages

## [0.1.0] - 2025-12-31

### Added
- Initial implementation
- Basic profile management (create, switch, delete, list)
- Toggle between profiles with `-`
- Current profile tracking
- Profile storage in `~/.claude/profiles/`
- Basic test suite (44 tests)

[Unreleased]: https://github.com/foxj77/claudectx/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/foxj77/claudectx/releases/tag/v1.0.0
[0.4.0]: https://github.com/foxj77/claudectx/releases/tag/v0.4.0
[0.3.0]: https://github.com/foxj77/claudectx/releases/tag/v0.3.0
[0.2.0]: https://github.com/foxj77/claudectx/releases/tag/v0.2.0
[0.1.0]: https://github.com/foxj77/claudectx/releases/tag/v0.1.0
