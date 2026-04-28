# Issue #21: Session-scoped profile launch

**Issue:** https://github.com/foxj77/claudectx/issues/21  
**Title:** `[RFE] Allow having multiple profiles enabled at the same time`  
**Status:** Proposed  
**Recommendation:** Implement as `claudectx run <profile> [-- <claude args...>]`

---

## Summary

Issue #21 asks for a way to use different claudectx profiles in separate terminal tabs at the same time. The example is:

- one terminal running Claude Code with the `work` profile
- another terminal running Claude Code with the `personal` profile

The current `claudectx <profile>` model cannot support this because it works by mutating global active files:

- `~/.claude/settings.json`
- `~/.claude/CLAUDE.md`
- MCP servers in `~/.claude.json`
- `~/.claude/.claudectx-current`
- `~/.claude/.claudectx-previous`

That design is correct for persistent global switching, but it is the wrong primitive for concurrent sessions. The right model is not "multiple global profiles enabled at the same time." The right model is "launch this one Claude Code process with this profile's settings, without changing global state."

The recommended feature is:

```bash
claudectx run <profile> [-- <claude args...>]
```

This command should start `claude` with profile-specific command-line flags and leave the user's active global configuration untouched.

---

## Is this a genuine need?

**Yes.** This is a real workflow gap.

The current tool answers:

> "Make this profile my default active Claude Code context."

Issue #21 asks for:

> "Start a Claude Code session using this profile, but do not change anything for other terminals."

Those are different use cases. Both are valid.

The request is especially strong because Claude Code already exposes session-scoped flags such as:

- `--settings`
- `--setting-sources`
- `--mcp-config`
- `--strict-mcp-config`
- `--append-system-prompt-file`
- `--model`
- `--permission-mode`

The installed local `claude --help` confirms these flags exist, and the Claude Code CLI reference documents the same direction: command-line arguments are session-level overrides and sit above user/project/local settings in precedence.

References:

- Claude Code CLI reference: https://code.claude.com/docs/en/cli-reference
- Claude Code configuration and settings precedence: https://code.claude.com/docs/en/configuration

---

## Primary use cases

### Work and personal sessions at the same time

A user may want a work terminal and a personal terminal open simultaneously. With current global switching, whichever profile was switched last affects new sessions and can affect assumptions in the user workflow. A process-scoped launcher avoids that.

### Client isolation

A consultant may work across several client repos. They may want:

```bash
claudectx run client-acme
claudectx run client-globex
```

in separate terminals. The goal is to avoid accidentally using the wrong model, permissions, MCP servers, or instructions for the wrong client.

### Side-by-side experimentation

Users may want to compare:

- two models
- two API providers
- strict vs permissive tool permissions
- different MCP server sets
- different global instruction sets

without rewriting their real active config each time.

### Safer one-off automation

This also enables commands like:

```bash
claudectx run review -- -p "Review this diff and list the highest-risk issues"
```

That is useful in scripts because the command becomes self-contained and does not depend on whatever profile was globally active before the script ran.

### Reduced accidental state changes

Some users may prefer `run` even for normal work because it does not auto-sync or mutate `~/.claude/settings.json`.

---

## Recommendation

Implement `claudectx run` as a **session launcher**.

It should:

- load the named profile from `~/.claude/profiles/<name>/`
- validate the profile
- build a `claude` command using profile-specific CLI flags
- execute `claude` with inherited stdin/stdout/stderr
- pass through any arguments after `--`
- return Claude's exit code
- avoid changing global claudectx state

It should not:

- call `SwitchProfile`
- write `~/.claude/settings.json`
- write `~/.claude/CLAUDE.md`
- update `~/.claude.json`
- create switch backups
- update `.claudectx-current`
- update `.claudectx-previous`
- auto-sync the previous current profile

This preserves the meaning of existing commands while adding a new workflow.

---

## Why not implement "multiple active profiles" directly?

The phrase "multiple profiles enabled at the same time" is understandable from the user's perspective, but it is not a good implementation model.

`claudectx` currently stores the active profile as a single file:

```text
~/.claude/.claudectx-current
```

Claude Code itself also has a default user config location:

```text
~/.claude/settings.json
```

Those are singleton locations. Trying to make them represent multiple active profiles would be confusing and brittle. It would create questions like:

- Which profile does `claudectx -c` show?
- Which profile does `claudectx -` toggle back to?
- Which profile receives auto-sync changes?
- What happens when two sessions both update settings?
- Which MCP servers are truly active?
- Which `CLAUDE.md` is global?

The better answer is to keep global switching global, and add a process-scoped launch path for concurrent usage.

---

## Current architecture impact

Relevant existing files:

| File | Current role |
|------|--------------|
| `main.go` | top-level argument routing; default unknown arg means switch profile |
| `cmd/switch.go` | mutates active global config and updates current/previous profile |
| `cmd/create.go` | snapshots active global config into a profile |
| `cmd/sync.go` | syncs active global config back into a profile |
| `internal/store/store.go` | loads and saves profile files |
| `internal/profile/profile.go` | in-memory profile model |
| `internal/paths/paths.go` | central path helpers |
| `internal/mcpconfig/mcpconfig.go` | reads/writes MCP server maps |

`run` can be added without disrupting most of this architecture. It should reuse:

- `store.Exists`
- `store.Load`
- `profile.Validate`
- `validator.ValidateSettings`
- `validator.ValidateClaudeMD`
- `paths.ProfileFile`
- `mcpconfig.SaveToFile` or a new MCP wrapper writer

The biggest new behavior is process execution.

---

## Proposed command interface

### Basic form

```bash
claudectx run <profile>
```

Starts an interactive Claude Code session using the named profile.

### Pass-through args

```bash
claudectx run <profile> -- <claude args...>
```

Examples:

```bash
claudectx run work -- --model opus
claudectx run personal -- --permission-mode plan
claudectx run review -- -p "Review this diff"
claudectx run ci -- -p "Check this repository for test failures" --output-format json
```

Using `--` avoids ambiguity between claudectx flags and Claude Code flags.

### Debug output

Add one of these:

```bash
claudectx run <profile> --dry-run
claudectx run <profile> --print-command
```

Recommended: `--dry-run`.

It should print the command that would be executed, plus generated temporary file paths, without launching Claude. This makes the feature easier to test and debug.

### Isolation mode

Consider:

```bash
claudectx run --isolated <profile>
```

This should attempt to avoid loading normal user-level settings and MCP servers. The exact flags depend on Claude Code behavior, but likely include:

```bash
--setting-sources project,local
--strict-mcp-config
```

This should be optional because users may expect normal Claude Code user preferences to remain available.

---

## Command routing

Add a `run` case in `main.go` before the default profile-switch case:

```go
case "run":
    if len(os.Args) < 3 {
        fmt.Fprintln(os.Stderr, "Error: profile name required")
        fmt.Fprintln(os.Stderr, "Usage: claudectx run <name> [-- <claude args...>]")
        os.Exit(1)
    }
    code, err := cmd.RunProfile(s, os.Args[2:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    os.Exit(code)
```

Do not overload `claudectx <name>` because that already means persistent global switch.

---

## How profile pieces should map to Claude CLI flags

### Settings

Profile file:

```text
~/.claude/profiles/<name>/settings.json
```

Claude flag:

```bash
--settings ~/.claude/profiles/<name>/settings.json
```

This is the strongest part of the feature. It maps directly to Claude Code's documented `--settings` flag.

Important nuance: `--settings` loads additional settings. It may not fully replace every other settings source unless combined with `--setting-sources`.

### MCP servers

Profile file today:

```text
~/.claude/profiles/<name>/mcp.json
```

Current stored format is a raw server map:

```json
{
  "github": {
    "type": "http",
    "url": "https://example.com/mcp"
  }
}
```

Claude's `--mcp-config` expects a config file shaped more like:

```json
{
  "mcpServers": {
    "github": {
      "type": "http",
      "url": "https://example.com/mcp"
    }
  }
}
```

Therefore `run` should generate a temporary MCP config wrapper:

```text
~/.claude/.claudectx-run/<pid-or-random>/mcp.json
```

with:

```json
{
  "mcpServers": {
    "...": {}
  }
}
```

Then invoke:

```bash
--mcp-config <generated-wrapper>
```

Recommended default:

```bash
--strict-mcp-config
```

Pros of default strict MCP:

- avoids leaking globally active MCP servers into a supposedly profile-scoped session
- makes `run` more reproducible
- aligns with the user expectation that a profile's MCP set is the one in use

Cons:

- users may expect Claude.ai MCP servers or normal user-scoped MCP servers to remain available
- strict behavior may surprise users whose profile has no MCP servers

Pragmatic recommendation:

- if the profile has MCP servers, pass `--mcp-config <file> --strict-mcp-config`
- if the profile has no MCP servers, do not pass either flag by default
- add `--isolated` to force strict empty MCP config if needed

### CLAUDE.md

Profile file:

```text
~/.claude/profiles/<name>/CLAUDE.md
```

Potential Claude flag:

```bash
--append-system-prompt-file ~/.claude/profiles/<name>/CLAUDE.md
```

This is useful but not perfect.

Normal persistent switching copies the profile's `CLAUDE.md` to:

```text
~/.claude/CLAUDE.md
```

Claude Code then discovers it as user-level memory/instructions. Appending a system prompt file may not be semantically identical to user-level `CLAUDE.md` discovery.

Options:

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| Append prompt file | Use `--append-system-prompt-file profile/CLAUDE.md` | no global mutation; simple; explicit | not exact `CLAUDE.md` semantics |
| System prompt replacement | Use `--system-prompt-file profile/CLAUDE.md` | very explicit | dangerous; replaces built-in/default prompt behavior |
| Temporary config dir | create alternate config home containing CLAUDE.md | closer parity | likely unsupported/fragile; may rely on undocumented env vars |
| Do nothing in v1 | only settings/MCP | safest technically | incomplete profile behavior |

Recommendation: use `--append-system-prompt-file` in v1 and document that it is an approximation of global `CLAUDE.md`.

If future Claude Code adds a direct flag for user memory files or config directory selection, migrate to that.

### Environment variables

Profile environment variables are already inside `settings.json` under `env`, so `--settings` should cover them.

There is no need for claudectx to manually inject them into the child process unless testing proves Claude Code does not apply `env` from `--settings`.

### Model and permissions

These are also settings-driven and should flow through `--settings`.

User-supplied pass-through args should take precedence:

```bash
claudectx run work -- --model opus
```

Claude Code's own precedence should handle this because command-line args outrank settings.

---

## Approaches considered

### Approach A: Keep current switch model only

Do nothing.

Pros:

- no implementation risk
- no change to current mental model
- no reliance on Claude CLI flags

Cons:

- does not solve the issue
- users cannot run multiple profile sessions concurrently
- encourages unsafe manual shell aliases or config copying

Verdict: not recommended.

### Approach B: Add `claudectx run` as a session launcher

Launch `claude` with profile-specific flags and do not mutate global state.

Pros:

- directly solves the issue
- low risk to existing behavior
- no migration needed
- aligns with Claude Code's session-scoped CLI model
- easy to explain: `switch` changes default, `run` launches one session
- useful for scripts and automation

Cons:

- may not perfectly match all behavior of persistent switching
- `CLAUDE.md` handling is approximate unless Claude supports an exact flag
- MCP requires temp config generation
- in-session settings changes may not sync back to the profile

Verdict: recommended.

### Approach C: Add temporary global swap around `claude`

The command could:

1. backup global config
2. switch to the profile
3. launch `claude`
4. restore previous config when Claude exits

Pros:

- closest parity with current switch behavior
- uses existing `SwitchProfile`
- `CLAUDE.md` semantics are exact

Cons:

- unsafe with multiple concurrent sessions
- race-prone: two `run` commands would overwrite each other
- a crash could leave the wrong global profile active
- background Claude processes may outlive the wrapper
- signal handling becomes tricky
- defeats the purpose of not mutating global state

Verdict: do not implement.

### Approach D: Use an alternate `CLAUDE_CONFIG_DIR`

Some tools historically used an environment variable or alternate config directory to point Claude Code at another config tree.

Pros:

- could provide high parity if officially supported
- would naturally isolate `settings.json` and `CLAUDE.md`

Cons:

- prior project notes mention `CLAUDE_CONFIG_DIR` as undocumented and buggy
- undocumented config directory overrides are fragile
- could break across Claude Code releases
- may interfere with auth/session state

Verdict: avoid unless Claude Code officially documents and supports it.

### Approach E: Generate a temporary project directory

Create a temp `.claude/` project config containing settings or instructions, then run Claude from there or with added dirs.

Pros:

- can avoid touching user global config
- may use documented project/local setting scopes

Cons:

- changes working directory semantics
- not appropriate for users who want Claude in the current repo
- `.claude` project settings have different precedence and sharing semantics
- not a clean fit for user-level profiles

Verdict: not recommended for this issue.

---

## Recommended v1 semantics

### Default

```bash
claudectx run work
```

Should roughly execute:

```bash
claude \
  --settings ~/.claude/profiles/work/settings.json \
  --append-system-prompt-file ~/.claude/profiles/work/CLAUDE.md \
  --mcp-config ~/.claude/.claudectx-run/<id>/mcp.json \
  --strict-mcp-config
```

Only include flags for files/features that exist:

- no `--append-system-prompt-file` if profile has no `CLAUDE.md`
- no `--mcp-config` if profile has no MCP servers, unless isolated mode asks for strict isolation

### Pass-through

```bash
claudectx run work -- --model opus --permission-mode plan
```

Should append pass-through args after claudectx-generated args so user-provided Claude flags can override profile defaults where Claude allows it.

### Dry run

```bash
claudectx run work --dry-run
```

Should print the generated command without executing it.

Example output:

```text
claude --settings /Users/me/.claude/profiles/work/settings.json --mcp-config /Users/me/.claude/.claudectx-run/abc123/mcp.json --strict-mcp-config --append-system-prompt-file /Users/me/.claude/profiles/work/CLAUDE.md
```

Do not print sensitive environment values.

### Isolated

```bash
claudectx run --isolated work
```

Should try to minimize ambient config. Proposed behavior:

- include profile settings with `--settings`
- include profile MCP config, or an empty generated MCP config
- include `--strict-mcp-config`
- include `--setting-sources project,local` if this reliably prevents user settings from loading while still allowing command-line `--settings`

This needs verification because `--settings` is a command-line settings input, and `--setting-sources` controls named settings sources. The expected goal is:

> load the explicit profile settings, but do not load normal user settings

If Claude Code does not support that exact combination, document the limitation and keep `--isolated` out of v1.

---

## Important behavioral caveats

### `run` is not `switch`

`run` should not update claudectx's current profile.

After:

```bash
claudectx -c
claudectx run personal
claudectx -c
```

the current profile should be unchanged.

### No auto-sync by default

Persistent switching currently auto-syncs changes from the active profile before switching away. `run` should not do that. It does not know which running Claude session owns changes, and it should not mutate profile files unexpectedly.

If future sync support is needed, make it explicit:

```bash
claudectx run work --sync-on-exit
```

Do not include this in v1.

### In-session `/config` changes may write elsewhere

If the user runs `/config` inside a `claudectx run work` session, Claude Code may write changes to the normal user or project config, not the profile's `settings.json`.

This must be documented. The command launches a session with profile settings; it does not necessarily redirect all future Claude Code writes.

### Session resume behavior may be shared

Claude Code session history and resume state may still be global or project-scoped. `claudectx run` should not promise isolated conversation history unless Claude Code provides flags for it.

Users can pass Claude flags such as `--name`, `--session-id`, `--no-session-persistence`, or `--fork-session` where appropriate.

### Auth is not profile-scoped

Authentication remains whatever Claude Code normally uses. `run` should not try to manage OAuth tokens or keychain entries.

If a profile uses API provider env vars in `settings.json`, that is fine. But claudectx should not attempt to switch login state.

---

## Implementation plan

### Phase 1: Add command parser support

Update `main.go`:

- add `case "run"`
- parse arguments
- show usage if profile missing
- preserve pass-through args after `--`
- support `--dry-run`
- optionally support `--isolated`

Suggested parser behavior:

```text
claudectx run <profile>
claudectx run <profile> -- <claude args...>
claudectx run --dry-run <profile>
claudectx run --isolated <profile>
claudectx run --isolated --dry-run <profile> -- <claude args...>
```

Keep parsing intentionally small. This project does not currently use a full CLI framework.

### Phase 2: Add `cmd/run.go`

Suggested public function:

```go
type RunOptions struct {
    ProfileName string
    ClaudeArgs  []string
    DryRun      bool
    Isolated    bool
}

func RunProfile(s *store.Store, opts RunOptions) (int, error)
```

Why return an exit code?

- if Claude exits with code 0, return 0
- if Claude exits with code 1, return 1 without wrapping it as an internal claudectx error
- if claudectx cannot build the command, return error

This lets `main.go` preserve child process status correctly.

### Phase 3: Build Claude command args

Create a helper:

```go
func BuildClaudeArgs(prof *profile.Profile, opts RunOptions) ([]string, *RunArtifacts, error)
```

`RunArtifacts` can track generated temporary files for cleanup:

```go
type RunArtifacts struct {
    TempDir string
    MCPConfigPath string
}
```

Argument order should be:

1. claudectx-generated profile args
2. user pass-through Claude args

This gives user args the best chance of overriding profile values.

### Phase 4: Settings flag

The settings file path should come from:

```go
paths.ProfileFile(profileName, "settings.json")
```

Add:

```go
--settings <path>
```

Do not write the settings file during `run`.

Validate settings before launching, using the same validator used by switch:

```go
validator.ValidateSettings(prof.Settings)
```

### Phase 5: CLAUDE.md flag

If `prof.ClaudeMD` is non-empty, add:

```go
--append-system-prompt-file <profile CLAUDE.md path>
```

Use the actual profile file path, not a temporary file, because `store.Load` already loaded the profile from disk and `store.Save` owns the profile file layout.

Validate with:

```go
validator.ValidateClaudeMD(prof.ClaudeMD)
```

If later testing shows `--append-system-prompt-file` is not available in the supported Claude Code version range, either:

- omit `CLAUDE.md` support in v1 and document it, or
- require a minimum Claude Code version.

### Phase 6: MCP config wrapper

Add a function in `internal/mcpconfig`, for example:

```go
func SaveClaudeMCPConfig(path string, servers MCPServers) error
```

It should write:

```json
{
  "mcpServers": {
    "...": {}
  }
}
```

Do not reuse `SaveToFile` for this unless its format is changed, because profile storage currently uses the raw server map.

In `run`:

- if profile has MCP servers, create a temp dir under `~/.claude/.claudectx-run/`
- write wrapper config
- add `--mcp-config <path>`
- add `--strict-mcp-config`

Temp dir naming:

```text
~/.claude/.claudectx-run/run-<timestamp>-<pid>/
```

Cleanup:

- for `--dry-run`, keep or print the temp path? Prefer not creating temp files in dry run; print what would be generated.
- for normal execution, remove temp dir after Claude exits.
- if cleanup fails, print a warning, not a fatal error.

Important: cleanup after Claude exits is safe because `claude` should read the config at startup. If Claude lazily reads MCP config later, cleanup may be too early. Verify this. If uncertain, leave generated run dirs and prune old dirs opportunistically.

### Phase 7: Execute child process

Use `os/exec`:

```go
command := exec.Command("claude", args...)
command.Stdin = os.Stdin
command.Stdout = os.Stdout
command.Stderr = os.Stderr
command.Env = os.Environ()
err := command.Run()
```

Exit code handling:

```go
if err == nil {
    return 0, nil
}
var exitErr *exec.ExitError
if errors.As(err, &exitErr) {
    return exitErr.ExitCode(), nil
}
return 1, fmt.Errorf("failed to run claude: %w", err)
```

This distinguishes "Claude ran and failed" from "claude executable not found."

### Phase 8: Signal handling

For v1, `exec.Command` with inherited stdio may be enough. But interactive CLI behavior can be sensitive to signals.

Risks:

- Ctrl+C may interrupt parent and child differently
- terminal state may not restore cleanly if parent handles signals incorrectly
- child may outlive parent in some cases

Recommended v1:

- keep parent minimal
- do not intercept signals unless needed
- let the OS deliver terminal signals to the foreground process group

If testing shows issues, consider:

- using `syscall.Exec` to replace the claudectx process with `claude`
- or setting process groups explicitly on Unix

`syscall.Exec` pros:

- perfect signal behavior
- no cleanup burden after child exits because there is no parent

`syscall.Exec` cons:

- cannot clean up temp MCP files afterward
- platform-specific
- harder to test

Recommendation: start with `exec.Command`; revisit if interactive signal behavior is poor.

### Phase 9: Help and documentation

Update `printHelp()` in `main.go`:

```text
claudectx run <NAME> [-- ARGS]   Run Claude with profile for this session only
```

Add examples:

```text
claudectx run work               Start Claude using 'work' without switching globally
claudectx run work -- --model opus
claudectx run review -- -p "Review this diff"
```

Update README:

- explain difference between `switch` and `run`
- add a "Concurrent Sessions" section
- document pass-through args
- document that `run` does not update current/previous profile
- document caveats around `/config`, resume history, auth, and `CLAUDE.md`

---

## Validation and error handling

### Missing profile

```text
Error: profile "work" does not exist
```

### Invalid settings

Reuse existing validation:

```text
Error: profile settings are invalid: ...
```

### Missing `claude` executable

```text
Error: failed to run claude: executable file not found in $PATH
```

Optionally add:

```text
Install Claude Code or ensure `claude` is available in PATH.
```

### Invalid pass-through usage

These should be accepted:

```bash
claudectx run work -- --model opus
claudectx run work --model opus
```

But the second form creates parser ambiguity. Recommendation for v1:

- require `--` before Claude args
- reject unknown claudectx flags before the profile
- document the rule clearly

Error:

```text
Error: unknown claudectx run flag "--model"; use `claudectx run work -- --model opus`
```

### Generated MCP config errors

If writing the temp config fails, fail before launching Claude:

```text
Error: failed to prepare MCP config for profile "work": ...
```

Do not fall back to global MCP silently. That would violate the user's profile expectation.

---

## Tests

### `cmd/run_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestRunProfile_MissingProfileFails` | clear error for unknown profile |
| `TestRunProfile_DoesNotModifyCurrentProfile` | `.claudectx-current` is unchanged |
| `TestRunProfile_DoesNotModifyActiveSettings` | `~/.claude/settings.json` is unchanged |
| `TestBuildClaudeArgs_IncludesSettings` | generated args include `--settings <profile/settings.json>` |
| `TestBuildClaudeArgs_IncludesClaudeMDWhenPresent` | generated args include `--append-system-prompt-file` |
| `TestBuildClaudeArgs_OmitsClaudeMDWhenEmpty` | no prompt file flag when profile has no instructions |
| `TestBuildClaudeArgs_AppendsPassThroughArgs` | args after `--` are preserved and ordered after generated args |
| `TestRunProfile_DryRunDoesNotExecuteClaude` | dry run prints/returns command without launching child |
| `TestRunProfile_ReturnsClaudeExitCode` | child exit code is propagated |
| `TestRunProfile_ClaudeExecutableMissingReturnsError` | missing executable is claudectx error, not child exit |

### `internal/mcpconfig`

| Test | What it verifies |
|------|-----------------|
| `TestSaveClaudeMCPConfig_WrapsServers` | writes `{ "mcpServers": ... }` |
| `TestSaveClaudeMCPConfig_EmptyServers` | writes valid empty config |
| `TestSaveClaudeMCPConfig_RoundTrip` | generated file can be parsed back |

### Parser tests

If main parsing is hard to test directly, extract:

```go
func ParseRunArgs(args []string) (RunOptions, error)
```

Tests:

| Test | What it verifies |
|------|-----------------|
| `TestParseRunArgs_ProfileOnly` | parses `run work` |
| `TestParseRunArgs_WithSeparator` | parses `run work -- --model opus` |
| `TestParseRunArgs_DryRunBeforeProfile` | parses `run --dry-run work` |
| `TestParseRunArgs_IsolatedBeforeProfile` | parses `run --isolated work` |
| `TestParseRunArgs_RejectsClaudeArgsWithoutSeparator` | rejects ambiguous flags |

---

## Manual verification

After implementation:

1. Create two profiles:

   ```bash
   claudectx -n work
   claudectx -n personal
   ```

2. Edit profile settings so they use visibly different models or permission defaults.

3. Run:

   ```bash
   claudectx run work
   ```

4. In another terminal, run:

   ```bash
   claudectx run personal
   ```

5. In both Claude sessions, run `/status` and verify the expected settings sources.

6. Verify global current profile did not change:

   ```bash
   claudectx -c
   ```

7. Verify global active config files did not change:

   ```bash
   git diff -- ~/.claude/settings.json
   ```

   Or compare checksums before and after if not in git.

8. Test MCP behavior by giving each profile a different MCP server and checking `/mcp`.

9. Test pass-through:

   ```bash
   claudectx run work -- --model opus
   ```

10. Test non-interactive mode:

    ```bash
    claudectx run work -- -p "Say which model and permission mode are active"
    ```

---

## Risks

### Incomplete parity with `switch`

Risk: `run` may not behave exactly like switching globally and then launching Claude.

Mitigation:

- document `run` as session-scoped
- be explicit about which profile components are applied
- avoid claiming full global-profile parity

### `CLAUDE.md` semantic mismatch

Risk: `--append-system-prompt-file` may not be equivalent to user-level `~/.claude/CLAUDE.md`.

Mitigation:

- document this as an approximation
- verify behavior manually with `/status` or known prompt effects
- update implementation if Claude adds a better flag

### Ambient user settings leakage

Risk: `--settings` may merge with normal user settings rather than replace them.

Mitigation:

- document precedence
- add `--isolated` only after verifying behavior
- use `--setting-sources` where appropriate
- add tests for generated args, and manual verification for actual Claude behavior

### Ambient MCP leakage

Risk: profile session also loads global MCP servers.

Mitigation:

- use `--strict-mcp-config` when profile MCP is provided
- consider empty strict MCP config in isolated mode

### Temporary file lifecycle

Risk: temp MCP config is deleted too early or left behind.

Mitigation:

- verify Claude reads MCP config at startup
- if uncertain, leave temp run dirs and prune old ones
- keep temp files under `~/.claude/.claudectx-run/`

### Signal handling

Risk: Ctrl+C or terminal resizing behaves poorly because claudectx is a parent process.

Mitigation:

- start with inherited stdio
- test interactive behavior
- switch to process replacement or process group handling only if needed

### Child exit codes

Risk: wrappers often collapse child failures into generic errors.

Mitigation:

- explicitly propagate Claude's exit code
- only return claudectx errors for claudectx setup failures

### User confusion between `switch` and `run`

Risk: users may not understand why `claudectx -c` does not change after `run`.

Mitigation:

- help text must say "for this session only"
- README must contrast `switch` vs `run`
- command output can say:

```text
Running Claude with profile "work" for this session only
```

### Reliance on Claude Code CLI compatibility

Risk: flags change across Claude Code versions.

Mitigation:

- rely only on documented flags
- keep unsupported optional behavior out of v1
- surface clear errors from Claude when flags are unsupported
- consider a `claude --version` diagnostic in `claudectx health`

---

## Pros and cons of implementing

### Pros

- solves a real concurrent workflow
- avoids global mutation
- reduces accidental profile leakage
- improves scripting and automation
- keeps existing `claudectx <name>` behavior intact
- likely small implementation surface compared with reworking profile storage
- aligns with Claude Code's documented session-scoped CLI flags

### Cons

- not a perfect equivalent to switching global files
- adds a second mental model: persistent switch vs session run
- requires careful docs
- MCP config format mismatch needs a small adapter
- `CLAUDE.md` behavior may be approximate
- actual isolation depends on Claude Code settings precedence behavior

---

## Suggested issue response

This is a good request and it is feasible, but I would frame it slightly differently. Rather than trying to make multiple claudectx profiles globally active at once, we should add a session-scoped launcher:

```bash
claudectx run <profile> [-- <claude args...>]
```

`claudectx <profile>` would keep its current meaning: switch the global active Claude Code config. `claudectx run <profile>` would launch one Claude Code process with that profile's settings and leave the global config/current profile untouched.

The implementation can use Claude Code's documented `--settings` flag, generate a temporary `--mcp-config` file for the profile's MCP servers, and pass the profile's `CLAUDE.md` via `--append-system-prompt-file` where available. We should be clear that this is session-scoped and may not be 100% identical to a global switch, especially for `CLAUDE.md` and settings-source precedence.

I think this is worth implementing because it unlocks concurrent work/personal/client sessions without races or global config churn.

---

## Acceptance criteria

1. `claudectx run <profile>` starts Claude Code using the named profile.
2. The command does not modify `~/.claude/settings.json`.
3. The command does not modify `~/.claude/CLAUDE.md`.
4. The command does not modify `~/.claude.json`.
5. The command does not update `.claudectx-current` or `.claudectx-previous`.
6. Profile settings are passed via `--settings`.
7. Profile MCP servers are passed via a generated Claude-compatible `--mcp-config` file.
8. Profile `CLAUDE.md` is passed via `--append-system-prompt-file` if supported.
9. Arguments after `--` are passed to `claude` unchanged.
10. The command returns Claude's exit code.
11. Missing profile, invalid profile, and missing `claude` executable produce clear errors.
12. README and help text explain the difference between `switch` and `run`.

---

## Recommended implementation order

1. Add parser and dry-run support.
2. Implement settings-only `run`.
3. Add pass-through args.
4. Add child process execution and exit-code propagation.
5. Add MCP wrapper generation.
6. Add `CLAUDE.md` prompt-file support.
7. Add tests.
8. Add README/help docs.
9. Manually verify with real `claude`.
10. Consider `--isolated` only after verifying `--setting-sources` behavior.

This keeps the first working version small while preserving a clear path to stronger isolation.
