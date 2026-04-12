# Installation & Testing Guide

## Quick Install (Local Development)

### Option 1: Install to System (Recommended)

```bash
# Navigate to the project
cd /Users/johnfox/Documents/claudectx

# Build and install to /usr/local/bin
go build -o claudectx
sudo mv claudectx /usr/local/bin/

# Verify installation
claudectx --version
```

Now you can use `claudectx` from anywhere!

### Option 2: Install via Go Install

```bash
# From the project directory
go install

# This installs to ~/go/bin/claudectx
# Make sure ~/go/bin is in your PATH
export PATH="$PATH:~/go/bin"

# Verify
claudectx --version
```

### Option 3: Use Without Installing

```bash
cd /Users/johnfox/Documents/claudectx
go build -o claudectx
./claudectx --version
```

## Testing It Out

### 1. Create Your First Profile from Current Config

```bash
# This saves your current ~/.claude/settings.json to a profile
claudectx -n default

# Output: Created profile "default" from current configuration
```

### 2. Create Additional Profiles

```bash
# Create a work profile
claudectx -n work

# Create a personal profile
claudectx -n personal

# View all profiles
claudectx
# Output:
# default
# personal
# work
```

### 3. Modify a Profile

Profiles are stored in `~/.claude/profiles/<name>/`. You can edit them directly:

```bash
# Edit the work profile settings
code ~/.claude/profiles/work/settings.json

# Or use vim
vim ~/.claude/profiles/work/settings.json

# Example: Add a custom API endpoint
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://my-proxy.example.com",
    "API_KEY": "sk-..."
  },
  "model": "opus"
}
```

### 4. Switch Between Profiles

```bash
# Switch to work profile
claudectx work
# Output: Switched to profile "work"

# Your ~/.claude/settings.json is now the work settings!

# Switch to personal
claudectx personal

# Toggle back to work
claudectx -
# Output: Switched to profile "work"
```

### 5. Verify the Switch Worked

```bash
# Check current profile
claudectx -c
# Output: work

# View current settings
cat ~/.claude/settings.json

# List all profiles (current is marked)
claudectx
# Output:
# default
# personal
# work (current)
```

### 6. Test with Claude Code

```bash
# Switch to a profile
claudectx work

# Start Claude Code
claude

# Claude will now use the work profile settings!
# Try switching profiles and restarting Claude to see different configs
```

## Use Cases

### Use Case 1: Work vs Personal Accounts

```bash
# Create work profile with company API key
claudectx -n work
vim ~/.claude/profiles/work/settings.json
# Add: {"env": {"ANTHROPIC_API_KEY": "sk-work-..."}}

# Create personal profile
claudectx -n personal
vim ~/.claude/profiles/personal/settings.json
# Add: {"env": {"ANTHROPIC_API_KEY": "sk-personal-..."}}

# Switch to work
claudectx work

# Switch to personal
claudectx personal
```

### Use Case 2: Different API Providers

```bash
# Official Anthropic API
claudectx -n anthropic
# Edit to use official endpoint

# Bedrock
claudectx -n bedrock
vim ~/.claude/profiles/bedrock/settings.json
# Add: {"env": {"CLAUDE_CODE_USE_BEDROCK": "1"}}

# Custom proxy
claudectx -n custom
# Add custom ANTHROPIC_BASE_URL
```

### Use Case 3: Different Tool Permissions

```bash
# Strict profile (limited tools)
claudectx -n strict
vim ~/.claude/profiles/strict/settings.json
# Add: {"permissions": {"allow": ["Read", "Grep"]}}

# Full access profile
claudectx -n full
# Add: {"permissions": {"allow": ["*"]}}

# Switch based on task
claudectx strict    # For reading only
claudectx full      # For making changes
```

### Use Case 4: Client-Specific Settings (For Consultants)

```bash
# Create profile for each client
claudectx -n client-acme
claudectx -n client-globex
claudectx -n client-initech

# Each can have different:
# - API keys (if clients provide their own)
# - Model preferences
# - CLAUDE.md instructions specific to that client
# - Tool permissions
```

## Managing CLAUDE.md

Each profile can have its own CLAUDE.md file:

```bash
# Create a profile
claudectx -n python-dev

# Edit its CLAUDE.md
vim ~/.claude/profiles/python-dev/CLAUDE.md
```

Example CLAUDE.md:
```markdown
# Python Development Guidelines

- Always use type hints
- Follow PEP 8
- Prefer pytest for testing
- Use black for formatting
```

When you switch to `python-dev`, this CLAUDE.md becomes active at `~/.claude/CLAUDE.md`.

## Troubleshooting

### Profile Not Switching?

```bash
# Check if profile exists
claudectx

# Verify current profile
claudectx -c

# Check settings were actually updated
cat ~/.claude/settings.json
```

### Lost Your Profiles?

Profiles are stored in `~/.claude/profiles/`:

```bash
# List all profiles
ls -la ~/.claude/profiles/

# Check a specific profile's files
ls -la ~/.claude/profiles/work/
```

### Want to Start Fresh?

```bash
# Delete all profiles and tracking
rm -rf ~/.claude/profiles/
rm ~/.claude/.claudectx-*

# Create a new default
claudectx -n default
```

### Can't Delete Current Profile?

You can't delete the active profile. Switch first:

```bash
# This will fail
claudectx -d work
# Error: cannot delete current profile "work"

# Switch first
claudectx personal

# Now you can delete
claudectx -d work
```

## Uninstall

```bash
# Remove the binary
sudo rm /usr/local/bin/claudectx

# Optionally remove profiles
rm -rf ~/.claude/profiles/
rm ~/.claude/.claudectx-*
```

## Advanced: Building from Source

```bash
# Clone the repo (when public)
git clone https://github.com/foxj77/claudectx.git
cd claudectx

# Run tests
go test ./...

# Build
go build -o claudectx

# Install
sudo mv claudectx /usr/local/bin/
```

## Getting Help

```bash
# Show help
claudectx --help

# Show version
claudectx --version
```

## What's Safe?

✅ **Safe to do:**
- Create as many profiles as you want
- Switch between profiles frequently
- Edit profile files directly
- Delete old profiles

⚠️ **Be careful:**
- Don't delete `~/.claude/profiles/` while profiles are in use
- Don't manually edit `.claudectx-current` or `.claudectx-previous`
- Always create a backup profile before experimenting

🚫 **Not managed by claudectx:**
- Project-level `.claude/settings.json` files
- OAuth session tokens
- Conversation history
- Project state
