# Review: issue21.md proposal

**Reviewed:** 2026-04-28  
**Proposal:** [docs/issue21.md](issue21.md)  
**Issue:** https://github.com/foxj77/claudectx/issues/21

---

## Overall verdict

Strong proposal. The author correctly reframes the request — "multiple profiles active simultaneously" is not a coherent implementation target, but "session-scoped profile launcher" is — and builds a clean design around it. The key decision (Approach B: `claudectx run`) is the right one, and the reasoning for rejecting the alternatives is sound. Recommend proceeding with the caveats below.

---

## Understanding the issue

Eduardo's request is simple and practical: run `claude` with a `work` profile in one terminal and a `personal` profile in another, at the same time. He even identifies the mechanism — Claude Code's `--settings` flag — and suggests a `claudectx run work` command.

The doc correctly identifies that the phrase "multiple profiles enabled at the same time" is misleading. Global state (`~/.claude/settings.json`, `.claudectx-current`) is inherently singular. Two concurrent sessions cannot own it without races. The right mental model is not multi-profile state; it is single-shot process launch with session-scoped settings.

---

## What the doc gets right

**Reframing the problem.** The distinction between persistent global switching (`claudectx <name>`) and session-scoped launch (`claudectx run <name>`) is the core insight, and it is explained well. Without that framing, any implementation would either race or confuse.

**Approach C rejection (temporary global swap).** The doc correctly calls this out as unsafe. Two concurrent `run` commands using a backup/restore approach would race, and a crash would leave the wrong profile active globally. Do not implement this.

**Approach D rejection (undocumented `CLAUDE_CONFIG_DIR`).** The project already has prior notes that this env var is undocumented and buggy. Relying on it would be fragile across Claude Code releases. Correct call.

**`--strict-mcp-config` default when MCP is present.** Without this, ambient global MCP servers would bleed into the profile session. The conditional logic (strict when profile has servers, silent when it does not) is the right default.

**Signal handling caution.** Recommending `exec.Command` with inherited stdio first, and reserving `syscall.Exec` for later, is pragmatic. The tradeoff (clean signals vs. temp file cleanup) is accurately described.

**Exit code propagation.** Explicitly distinguishing "Claude ran and failed" from "claudectx setup failed" and "claude binary missing" is the right level of care. This matters for scripting.

**Dry-run first.** Making `--dry-run` part of v1 is valuable because it lets users inspect the generated command without launching Claude. This also makes the feature easier to test in CI without needing a running `claude`.

---

## Concerns and pushback

### 1. `--settings` precedence — resolved ✓

**Investigated 2026-04-28. This concern is resolved and the outcome is favourable.**

`--settings` sits at command-line precedence, which is above user/project/local settings. The full order (highest to lowest):

1. Managed settings (MDM/server — cannot be overridden)
2. Command-line arguments — `--settings` lands here
3. Local project settings (`.claude/settings.local.json`)
4. Shared project settings (`.claude/settings.json`)
5. User settings (`~/.claude/settings.json`)

**Scalar settings (model, permission mode, env vars):** `--settings <profile/settings.json>` overrides the global `~/.claude/settings.json`. Confirmed by live test — a profile specifying `claude-haiku-4-5-20251001` overrides the default `claude-sonnet-4-6`. Profile settings win.

**Multiple `--settings` flags:** First-specified wins for conflicting scalar keys. Live test confirmed.

**`--settings` and `--setting-sources` are independent.** Passing `--setting-sources project,local` drops global user settings; `--settings <profile>` still loads cleanly on top.

**One genuine caveat to document:** Array-valued settings (`permissions.allow`, `permissions.deny`, `permissions.additionalDirectories`) are **merged and deduplicated** across all scopes — they are never replaced. A profile's allow list accumulates with the global allow list. This means `claudectx run` cannot use a profile to *restrict* permissions already granted globally. This must be documented clearly. It is not a blocker, but users who expect strict permission isolation will not get it from `--settings` alone — they would need `--isolated` mode (which remains a future feature pending `--setting-sources` verification).

### 2. `CLAUDE.md` via `--append-system-prompt-file` — resolved ✓

**Investigated 2026-04-28. Behaviour confirmed empirically.**

Five tests were run against a live Claude Code session:

| Test | Result |
|------|--------|
| Does the appended file content reach Claude? | Yes — a unique marker phrase was confirmed present |
| Does it persist to new sessions (leaked into memory)? | No — a fresh session had no knowledge of the marker |
| Does global `~/.claude/CLAUDE.md` still apply alongside it? | Yes — both active simultaneously |
| Does the appended file override global `CLAUDE.md` on conflict? | No — global wins |
| Does a missing file cause an error (exit 1)? | Yes — must guard with existence check |
| Does an empty file cause an error? | No — exits 0 safely |

**What this means for `claudectx run`:**

- `--append-system-prompt-file` is genuinely additive and session-scoped. It is the right mechanism for v1.
- The global `~/.claude/CLAUDE.md` always remains active. Profile CLAUDE.md content is layered on top, not a replacement.
- A profile cannot use its `CLAUDE.md` to override instructions in the user's global `CLAUDE.md`. This is expected and should be documented.
- Implementation must check that the profile's `CLAUDE.md` file exists **and is non-empty** before adding the flag. A missing file causes exit code 1; an empty file is harmless but passes a no-op flag that would mislead debugging.
- Do not write "profile CLAUDE.md is applied" in the README — write "profile CLAUDE.md is appended to the system prompt for this session."

### 3. MCP temp file cleanup — resolved ✓

**Investigated 2026-04-28. Cleanup after Claude exits is safe.**

Debug logging confirms Claude reads `--mcp-config` once at startup (`[STARTUP] Loading MCP configs...`) and immediately launches the server process. The file is not re-read during the session. Deleting the temp file after Claude exits is safe.

Strategy for v1: attempt cleanup after Claude exits, warn on failure, do not fail the process. On `--dry-run`, do not write temp files — print the path that would be generated.

**Additional finding:** The MCP wrapper must faithfully copy the `env` block from the profile's stored server config. The global `~/.claude.json` stores env vars (e.g. `KUBECONFIG`) alongside each server definition; if the generated temp file omits them, the server process will fail to start. The current `internal/mcpconfig` format already captures `env`, so `SaveClaudeMCPConfig` just needs to pass it through. This must be explicitly tested.

### 4. `ParseRunArgs` design — resolved ✓

**Investigated 2026-04-28. Design locked based on existing codebase patterns.**

`main.go` uses a flat `switch` on `os.Args[1]` with no CLI framework. The `run` case should follow the same pattern: minimal, hand-rolled, tested via an extracted `ParseRunArgs` function.

`--isolated` is cut from v1 (see item 5 below), which simplifies the parser considerably. The only flags needed in v1 are `--dry-run` and `--` separator.

**Recommended `ParseRunArgs` signature:**

```go
type RunOptions struct {
    ProfileName string
    ClaudeArgs  []string
    DryRun      bool
}

func ParseRunArgs(args []string) (RunOptions, error)
```

**Parsing rules (args = everything after `"run"`):**

1. Scan left to right for `--dry-run`; set `DryRun = true`, remove from slice.
2. The first non-flag token is `ProfileName`.
3. If `--` appears after the profile name, everything after it is `ClaudeArgs`.
4. Any unrecognised flag before `--` returns an error with a hint to use `--`.

**Valid forms in v1:**

```bash
claudectx run work
claudectx run --dry-run work
claudectx run work -- --model opus
claudectx run --dry-run work -- --model opus
```

**Error on ambiguous flags (no `--` separator):**

```
Error: unknown flag "--model"; use: claudectx run work -- --model opus
```

This is the same defensive pattern used elsewhere in the project. Write `ParseRunArgs` and its unit tests before implementing `cmd/run.go`. The parser is the highest-risk part of the feature to get wrong silently.

### 5. `--isolated` — cut from v1 ✓

`--setting-sources` interaction with command-line `--settings` is unverified. Shipping an option that may silently do nothing would damage trust. Cut from v1 explicitly. Document as a future addition once confirmed. No further investigation needed before v1 ships.

### 6. The `syscall.Exec` option deserves a clearer call

The doc presents `syscall.Exec` as a potential fallback for signal handling. There is a better framing: `syscall.Exec` replaces the claudectx process with the `claude` process entirely, which means temp MCP files written before exec cannot be cleaned up afterward.

This is not just a con — it is a blocker for using `syscall.Exec` when MCP config generation is in play. Start with `exec.Command`. Only revisit `syscall.Exec` if MCP temp files are moved to a persistent location that Claude Code reads on demand (which is unlikely).

---

## One gap not addressed: `claude` binary location

The doc assumes `claude` is in `$PATH`. For most Claude Code installations this is true, but some users install via the app and rely on the shell integration. A missing binary should produce a clear, actionable error:

```
Error: "claude" not found in PATH
Install Claude Code CLI or run: npm install -g @anthropic-ai/claude-code
```

The doc's error table mentions this but does not include the install hint. Add it.

---

## Recommendation

All open items resolved. Proceed with `claudectx run`. The v1 scope is now fully defined.

**Confirmed safe to implement:**
- `--settings <profile/settings.json>` overrides global scalar settings (model, permission mode, env). Permission arrays accumulate across scopes — document this clearly.
- `--append-system-prompt-file <profile/CLAUDE.md>` is session-scoped and additive. Global `~/.claude/CLAUDE.md` remains active alongside it. Only pass the flag if the file exists and is non-empty (missing file causes exit 1; empty file is a no-op).
- MCP temp file is read at startup only — safe to delete after Claude exits. Must copy the full server config including the `env` block, or servers requiring env vars will crash on connect.
- `--isolated` cut from v1.

**Implementation order:**
1. Write `ParseRunArgs(args []string) (RunOptions, error)` and its unit tests first.
2. Add `case "run"` in `main.go`.
3. Implement `cmd/run.go`: load profile → validate → build args → write MCP temp file → exec claude → cleanup.
4. Settings flag: always include. Validate with `validator.ValidateSettings` before launch.
5. CLAUDE.md flag: include only if file exists and is non-empty.
6. MCP flag: include only if profile has servers. Copy full server config including `env`. Add `--strict-mcp-config`. Attempt cleanup after exit, warn on failure, do not fail.
7. Pass-through args: append after generated args so user flags can override profile values.
8. Dry-run: print generated command, skip temp file creation, skip exec.
9. Update `printHelp()` in `main.go` and README. Missing `claude` binary error must include an install hint.

The feature unlocks a real concurrent workflow, keeps existing switch behaviour untouched, and has a well-bounded implementation surface. All empirical questions are now answered.
