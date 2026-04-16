# Issue #16: "plugins are gone when switching profiles"

**Reported by:** e-minguez  
**Date:** 2026-04-15  
**Status:** In progress  
**Branch:** `fix/settings-unknown-fields-issue-16`

---

## What the user reported

When a Claude Code plugin (specifically obra/superpowers) is installed, and then the user switches to a different claudectx profile and back to the original, the plugin no longer appears.

---

## Is this a genuine bug?

**Yes.** The root cause is confirmed by reading the source code. It is not a user error or a misunderstanding of the tool. Every profile switch or sync with claudectx silently destroys fields in `~/.claude/settings.json` that the tool does not know about.

---

## Root cause

### The Settings struct is incomplete

`internal/config/config.go` defines the data structure for `settings.json` as:

```go
type Settings struct {
    Env         map[string]string `json:"env,omitempty"`
    Model       string            `json:"model,omitempty"`
    Permissions *Permissions      `json:"permissions,omitempty"`
}

type Permissions struct {
    Allow []string `json:"allow,omitempty"`
    Deny  []string `json:"deny,omitempty"`
}
```

But a real `~/.claude/settings.json` on disk looks like this:

```json
{
    "permissions": {
        "allow": ["mcp__flux-operator-mcp__*"],
        "defaultMode": "bypassPermissions"
    },
    "effortLevel": "medium",
    "autoDreamEnabled": true,
    "skipDangerousModePermissionPrompt": true
}
```

The struct does not model `effortLevel`, `autoDreamEnabled`, `skipDangerousModePermissionPrompt`, or `permissions.defaultMode`. Go's `encoding/json` silently ignores any JSON key that has no matching struct field.

### The destructive write cycle

Every time claudectx switches or syncs a profile, it runs this cycle:

1. **`LoadSettings(path)`** — reads `settings.json` into `Settings`. Unknown fields are dropped from memory.
2. **`SaveSettings(path, settings)`** — marshals the struct back to JSON and overwrites the file. Only the 3 known fields are written.

Result: after any profile switch or sync, `effortLevel`, `autoDreamEnabled`, `skipDangerousModePermissionPrompt`, and `permissions.defaultMode` are **permanently erased** from `settings.json`.

### How this causes plugins to disappear

Claude Code stores plugin-related configuration in `settings.json`. If it tracks active plugins via a field that is not in claudectx's `Settings` struct (for example, a `plugins` key, an `enabledExtensions` key, or any other key the tool does not know about), that field is wiped on every profile switch. The plugin appears installed (the binary/files are still in `~/.claude/plugins/`) but Claude Code no longer has its activation record in settings.json, so the plugin appears absent.

The same destruction applies to any current or future Claude Code settings field that claudectx doesn't explicitly model.

---

## Affected code paths

| File | Function | Problem |
|------|----------|---------|
| `internal/config/config.go:53` | `SaveSettings` | Marshals only known struct fields, erasing all others |
| `internal/config/config.go:25` | `LoadSettings` | Drops unknown fields on parse |
| `cmd/switch.go:95` | `SwitchProfile` | Calls `SaveSettings` on every switch |
| `cmd/sync.go:136` | `syncCurrentProfile` | Calls `LoadSettings` + `SaveSettings` on every auto-sync |
| `cmd/sync.go:58` | `hasConfigChanged` | Compares active settings to stored profile; both sides have already had unknowns stripped |

---

## What claudectx does NOT manage (and should not)

`~/.claude/plugins/` — the plugin install registry — is never touched by claudectx. This is correct: installed plugins are machine-scoped, not profile-scoped (they are binaries and cached files, not just configuration). The problem is not that claudectx needs to manage this directory; it is that claudectx destroys the settings.json fields that reference those plugins.

---

## Scenario walkthrough (concrete reproduction)

1. User is on profile `work`. Active `settings.json` contains `{"model":"opus","plugins":["superpowers"],"effortLevel":"high"}`.
2. User runs `claudectx personal` to switch.
3. claudectx auto-syncs `work`: calls `syncCurrentProfile` → `LoadSettings` → `settings.Settings{Model:"opus"}` (plugins and effortLevel are gone from memory) → `SaveSettings` → writes `{"model":"opus"}` to `work`'s stored profile. **`plugins` and `effortLevel` are permanently deleted from the stored profile.**
4. claudectx applies `personal` profile to active `settings.json`.
5. User runs `claudectx work` to switch back.
6. claudectx restores `work`'s stored settings: `{"model":"opus"}`. No `plugins` key. Plugin is gone.

---

## The fix

Apply the same "unknown-field preservation" pattern already used in `internal/mcpconfig/mcpconfig.go` to the `Settings` and `Permissions` structs.

`mcpconfig.SaveMCPServers` already does this correctly for `~/.claude.json`: it reads the file as `map[string]json.RawMessage`, modifies only the `mcpServers` key, and writes all other keys back unchanged. The identical approach should be applied to `settings.json`.

**Implementation:** add custom `UnmarshalJSON`/`MarshalJSON` methods to both `Settings` and `Permissions`, storing unrecognised keys in an unexported `extras map[string]json.RawMessage` field. `LoadSettings` and `SaveSettings` require no changes — they call `json.Unmarshal`/`json.MarshalIndent` as before, and the custom methods are invoked automatically.

---

## Test plan (TDD)

Tests are written before the fix to demonstrate the bug, then the fix is applied to make them pass.

### `internal/config/config_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestLoadSaveSettings_PreservesUnknownTopLevelFields` | `effortLevel`, `autoDreamEnabled` survive a `LoadSettings` → `SaveSettings` roundtrip |
| `TestLoadSaveSettings_PreservesUnknownPermissionsFields` | `permissions.defaultMode` survives a roundtrip |
| `TestLoadSettings_KnownFieldsStillAccessible` | After loading a rich file, `Model`, `Env`, `Permissions.Allow` are correctly populated (regression guard) |
| `TestSaveSettings_ModifyKnownField_PreservesUnknownFields` | Modifying `settings.Model` in memory and saving leaves unknown fields intact |
| `TestSaveSettings_NilPermissions_PreservesExistingPermissionsUnknownFields` | When `Permissions == nil` in the struct, existing `permissions.defaultMode` survives |

### `cmd/switch_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestSwitchProfile_PreservesProfileStoredUnknownFields` | Switching to a profile whose stored settings has `effortLevel:"high"` results in active settings.json containing `effortLevel:"high"` |
| `TestSwitchProfile_AutoSync_PreservesUnknownFieldsInProfile` | Auto-sync triggered during a switch saves unknown fields (including a `plugins` key) from active settings.json into the stored profile |

### `cmd/sync_test.go` (new file)

| Test | What it verifies |
|------|-----------------|
| `TestSyncCurrentProfile_PreservesUnknownSettingsFields` | Explicit sync captures `effortLevel` and `skipDangerousModePermissionPrompt` into the stored profile |
| `TestHashSettings_StableAcrossUnknownFields` | `hashSettings` produces the same hash before and after a roundtrip — no spurious change detection |

---

## Files changed

| File | Change |
|------|--------|
| `internal/config/config.go` | Add `extras` field + custom `UnmarshalJSON`/`MarshalJSON` to `Settings` and `Permissions` |
| `internal/config/config_test.go` | Add 5 new tests |
| `cmd/switch_test.go` | Add 2 new tests |
| `cmd/sync_test.go` | New file with 2 tests |
| `CLAUDE.md` | Note the unknown-field preservation guarantee |
| `item 16.md` | This file |
