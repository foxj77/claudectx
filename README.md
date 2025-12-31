# claudectx

**Fast, safe profile switching for Claude Code**

Switch between Claude Code configurations in seconds. Perfect for managing multiple accounts, API providers, or client-specific settings.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Tests](https://img.shields.io/badge/tests-85%2B%20passing-brightgreen)

---

## What is claudectx?

claudectx is a command-line tool that lets you quickly switch between different Claude Code configurations. Think of it like profiles for your browser, but for Claude Code.

**Perfect for:**
- 👔 Switching between work and personal Claude accounts
- 🏢 Managing different client configurations
- 🔌 Testing different API providers (Anthropic, Bedrock, custom)
- 🛠️ Using different tool permissions and MCP servers
- 📝 Maintaining separate instruction sets (CLAUDE.md files)

---

## Quick Start

### Interactive Mode

The easiest way to switch profiles:

```bash
claudectx
```

Use **↑/↓ arrow keys** to navigate, **Enter** to select:

```
Select a profile:

  work
❯ personal (current)
  client-acme

Use ↑/↓ to navigate, Enter to select, Esc/Ctrl+C to cancel
```

### Direct Switch

If you know the profile name:

```bash
claudectx work
```

### Toggle Between Profiles

Quickly switch back and forth:

```bash
claudectx -
```

---

## Installation

### macOS / Linux (Homebrew)

```bash
brew install foxj77/tap/claudectx
```

This will install claudectx and shell completions automatically.

### Manual Installation

**Option 1: Install to your user directory (no sudo required)**

```bash
cd ~/Downloads
git clone https://github.com/foxj77/claudectx.git
cd claudectx
make install-user
```

This installs to `~/go/bin/claudectx`. Make sure `~/go/bin` is in your PATH:

```bash
# For bash
echo 'export PATH="$PATH:~/go/bin"' >> ~/.bashrc
source ~/.bashrc

# For zsh
echo 'export PATH="$PATH:~/go/bin"' >> ~/.zshrc
source ~/.zshrc
```

**Option 2: Install system-wide**

```bash
cd ~/Downloads
git clone https://github.com/foxj77/claudectx.git
cd claudectx
sudo make install
```

This installs to `/usr/local/bin/claudectx`.

**Option 3: Download pre-built binary**

Download pre-built binaries for your platform from the [releases page](https://github.com/foxj77/claudectx/releases).

Available for:
- macOS (Intel and Apple Silicon)
- Linux (x64 and ARM64)
- Windows (x64)

---

## Complete Usage Guide

### Managing Profiles

**Create a new profile** from your current settings:
```bash
claudectx -n work
```

**List all profiles** (simple text output):
```bash
claudectx -l
```

**Show current profile**:
```bash
claudectx -c
```

**Delete a profile**:
```bash
claudectx -d old-client
```

### Advanced Features

**Export a profile** to share with teammates:
```bash
# Export to file
claudectx export work work-profile.json

# Export to stdout (for piping)
claudectx export work
```

**Import a profile**:
```bash
# Import from file
claudectx import work-profile.json

# Import and rename
claudectx import work-profile.json client-new

# Import from stdin
cat work-profile.json | claudectx import
```

**Check profile health** (validates settings):
```bash
# Check current profile
claudectx health

# Check specific profile
claudectx health work
```

**Transfer profiles between machines**:
```bash
claudectx export work | ssh remote-machine 'claudectx import - work'
```

---

## Real-World Examples

### Freelancer Managing Multiple Clients

```bash
# Create profiles for each client
claudectx -n client-acme
claudectx -n client-globex
claudectx -n personal

# Switch to client work
claudectx client-acme

# Quick toggle between client and personal
claudectx -

# Export client profile for backup
claudectx export client-acme ~/backups/acme-$(date +%Y%m%d).json
```

### Developer with Work and Personal Accounts

```bash
# Create work profile
claudectx -n work

# Create personal profile
claudectx -n personal

# Start work day
claudectx work

# End of day, switch to personal
claudectx personal

# Or use interactive mode
claudectx
```

### Team Sharing Configuration

```bash
# Team lead creates and exports standard config
claudectx -n team-standard
claudectx export team-standard team-config.json

# Share file with team (email, Slack, git repo, etc.)

# Team members import
claudectx import team-config.json work
```

---

## What Gets Switched?

When you switch profiles, claudectx manages:

- ✅ `~/.claude/settings.json` - All your Claude Code settings
- ✅ `~/.claude/CLAUDE.md` - Your global instructions
- ✅ Model preferences (opus, sonnet, haiku)
- ✅ Environment variables
- ✅ Tool permissions
- ✅ API configuration

**What stays the same:**
- ❌ Project-level settings in `.claude/` folders
- ❌ OAuth session tokens
- ❌ Conversation history
- ❌ Project-specific configurations

---

## Safety Features

claudectx is designed to be **safe and reliable**:

🛡️ **Automatic Backups**
Every switch creates a timestamped backup in `~/.claude/backups/`

🔍 **Validation**
Profiles are validated before switching to prevent corruption

↩️ **Automatic Rollback**
If anything goes wrong during a switch, your previous config is automatically restored

💾 **Atomic Operations**
Settings files are updated atomically - no partial updates

🎨 **Clear Feedback**
Color-coded output shows success (green), warnings (yellow), and errors (red)

---

## Shell Completion

Enable tab completion for your shell:

**Bash:**
```bash
source /path/to/claudectx/completion/bash_completion.sh
```

**Zsh:**
```bash
# Copy to a directory in your $fpath
cp /path/to/claudectx/completion/zsh_completion.sh /usr/local/share/zsh/site-functions/_claudectx
```

**Fish:**
```bash
cp /path/to/claudectx/completion/fish_completion.fish ~/.config/fish/completions/
```

---

## Troubleshooting

### "command not found: claudectx"

Make sure the installation directory is in your PATH:

```bash
# Check if ~/go/bin is in PATH
echo $PATH | grep go/bin

# If not, add it:
echo 'export PATH="$PATH:~/go/bin"' >> ~/.zshrc  # or ~/.bashrc
source ~/.zshrc  # or source ~/.bashrc
```

### "profile does not exist"

List available profiles to see what exists:

```bash
claudectx -l
```

### Accidentally deleted a profile

Check your backups:

```bash
ls ~/.claude/backups/
```

Each backup directory contains a complete copy of your settings.

### Settings aren't taking effect

Make sure Claude Code isn't running when you switch profiles. Restart Claude Code after switching.

---

## How It Works

claudectx stores each profile in `~/.claude/profiles/`:

```
~/.claude/
├── .claudectx-current      # Tracks which profile is active
├── .claudectx-previous     # Enables toggle with 'claudectx -'
├── profiles/
│   ├── work/
│   │   ├── settings.json
│   │   └── CLAUDE.md
│   └── personal/
│       ├── settings.json
│       └── CLAUDE.md
├── backups/                # Automatic backups
│   └── backup-1234567890/
└── settings.json           # Active config (symlinked)
```

When you switch profiles, claudectx copies the profile's files to the active locations.

---

## Get Help

**View all commands:**
```bash
claudectx --help
```

**Check version:**
```bash
claudectx --version
```

**Found a bug?**
[Open an issue](https://github.com/foxj77/claudectx/issues) on GitHub

---

## Comparison with Other Tools

| Feature | claudectx | cctx | Manual switching |
|---------|-----------|------|------------------|
| Interactive selection | ✅ | ❌ | ❌ |
| Automatic backups | ✅ | ❌ | ❌ |
| Validation | ✅ | ❌ | ❌ |
| CLAUDE.md support | ✅ | ❌ | ✅ |
| Export/import | ✅ | ❌ | ❌ |
| Shell completion | ✅ | ❌ | ❌ |
| Health checks | ✅ | ❌ | ❌ |
| Rollback on error | ✅ | ❌ | ❌ |

---

## License

MIT License - see [LICENSE](LICENSE)

---

## Credits

Inspired by [kubectx](https://github.com/ahmetb/kubectx) - the excellent Kubernetes context switcher.

Built by [John Fox](https://github.com/foxj77) for the Claude Code community.
