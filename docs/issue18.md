# Issue #18: Profile-aware global skill switching

**Issue:** https://github.com/foxj77/claudectx/issues/18  
**Status:** Proposed  
**Recommendation:** Implement, but with safer semantics than the original proposal

---

## Summary

Issue #18 asks claudectx to switch Claude Code global skills along with profiles. The request is sound: skills can materially change Claude Code behavior, and users who maintain separate work, personal, client, or security-sensitive profiles may reasonably expect skill state to be part of that context.

The implementation should not directly treat `~/.claude/skills/` as ordinary profile content to freely replace. Skills are installed assets, not just preferences. Some may come from plugins, curated repos, private repos, or manual local changes. Blindly swapping directory contents would create data-loss and supply-chain risk.

The recommended approach is to add an opt-in, manifest-based feature first:

- profiles can declare the desired active global skills
- claudectx can enable or disable installed skills to match that manifest
- missing skills produce warnings, not automatic installs
- exports include skill names only, not skill source contents
- profiles with no skill manifest leave skills untouched

---

## Is this worth implementing?

**Yes, but as a controlled feature.**

This fits claudectx's purpose: switching complete Claude Code operating contexts. The existing tool already manages:

- `~/.claude/settings.json`
- `~/.claude/CLAUDE.md`
- MCP servers from `~/.claude.json`
- profile import/export
- auto-sync on switch
- backup and rollback

Global skills are another context input. If a work profile expects code review, deployment, or compliance skills, and a personal profile expects writing or brainstorming skills, leaving skills unchanged makes profile switching incomplete.

The issue is strong enough to warrant implementation after core safety/correctness work, but the first version should be conservative. It should manage active skill selection, not skill installation or skill packaging.

---

## Primary use cases

### Work vs personal

A work profile may enable skills for code review, tests, deployment, and incident response. A personal profile may enable writing, brainstorming, or learning-oriented skills. Switching profiles should not leave a work-only skill active in a personal context, or vice versa.

### Client isolation

Client-specific profiles may have domain skills, terminology, checklists, or workflow instructions. Keeping those active while switching to another client risks context leakage and bad guidance.

### Security-sensitive contexts

Some profiles may intentionally minimize extra skills. For example, a production operations profile might enable only tightly reviewed operational skills, while a sandbox profile may allow experimental skills.

### Team standardization

Exported profiles can communicate "this profile expects these skills" without bundling executable or prompt content. A teammate can import the profile, see missing skill warnings, and install approved skills through the normal mechanism.

---

## Important distinction: installed vs active

The implementation must distinguish these concepts:

| Concept | Meaning | Should claudectx manage it? |
|---------|---------|-----------------------------|
| Installed skill | The actual skill directory/content on disk | No, not in v1 |
| Active skill | A skill that Claude Code should currently see/use | Yes, if profile opts in |
| Desired skill manifest | The profile's list of active skill names | Yes |
| Skill installation source | Git repo, plugin, curated source, local files | No, not in v1 |

This distinction keeps claudectx from becoming a package manager and avoids copying untrusted or private skill content into exports.

---

## Proposed profile format

Add an optional per-profile file:

```text
~/.claude/profiles/<name>/skills.json
```

Recommended v1 schema:

```json
{
  "enabled": [
    "code-reviewer",
    "tdd",
    "deployment-helper"
  ]
}
```

Do not use the original issue's top-level `"skills"` key unless there is a compatibility reason. `"enabled"` is clearer because it says these are active skill names, not bundled skill definitions.

Behavior:

- If `skills.json` is missing, claudectx leaves global skills untouched.
- If `skills.json` exists with an empty `enabled` list, claudectx disables all currently enabled global skills it manages.
- Unknown fields in `skills.json` should be ignored or preserved where practical, so the format can evolve.

---

## Recommended filesystem model

This depends on how Claude Code recognizes active skills. The implementation should verify this before coding. Assuming global skills are active when present under `~/.claude/skills/`, use a claudectx-owned disabled directory:

```text
~/.claude/
├── skills/
│   ├── code-reviewer/
│   └── tdd/
└── .claudectx-disabled-skills/
    └── creative-writing/
```

Switching to a manifest should:

1. list installed/available skill directories from both active and disabled locations
2. move desired skills into `~/.claude/skills/`
3. move undesired currently active skills into `~/.claude/.claudectx-disabled-skills/`
4. warn when a desired skill is missing from both locations

Use directory rename operations where possible. Avoid copying skill contents during normal switching, because copying can duplicate state and make ownership unclear.

If Claude Code has an official enable/disable registry in `settings.json` rather than directory presence, prefer changing that registry instead of moving directories. The same manifest model still applies.

---

## Safety rules

### Opt-in only

Profiles created before this feature must not unexpectedly change skills. Missing `skills.json` means "do not manage skills for this profile."

### Backup before mutation

The backup manager should include both active and disabled skill state when a skill-aware switch will mutate skills. Rollback must restore the skill directories or registry to the pre-switch state.

### No automatic installs

If `skills.json` references a skill that is not installed, print a warning:

```text
Warning: Skill "deployment-helper" is listed in profile "work" but is not installed
```

Do not fetch from the network. Do not infer source repos. Do not create placeholder skills.

### Do not export skill contents

`claudectx export` should include names only:

```json
{
  "skills": {
    "enabled": ["code-reviewer", "tdd"]
  }
}
```

This avoids leaking private skill instructions or bundling executable files. A future explicit `--bundle-skills` feature could be considered separately, but should not be part of issue #18.

### Avoid deleting skills

Switching should move skills between active and disabled locations, not delete them. Deletion should remain a user-initiated operation outside this feature.

### Validate names

Skill names must reject path separators, empty names, `.` and `..`, and hidden path tricks. Reuse the same defensive spirit as `profile.ValidateProfileName`.

---

## User-facing behavior

### Create

Default recommendation:

```text
claudectx -n work
```

For backward compatibility, this should probably continue creating a normal profile without skill management unless the user opts in.

Add a flag:

```text
claudectx -n work --with-skills
```

This snapshots currently active global skill names into the profile's `skills.json`.

If adding flags is too much for the first cut, document manual creation of `skills.json` and keep automatic snapshotting out of v1.

### Switch

```text
claudectx work
```

Behavior:

- if `work/skills.json` exists, apply the manifest
- if it does not exist, leave skills unchanged
- print a short summary only when skills are managed:

```text
Skills: enabled 2, disabled 1, missing 1
```

### Sync

`claudectx sync` should preserve the current profile's existing skill manifest unless skill sync is explicitly requested.

Recommended flag:

```text
claudectx sync --skills
```

This updates the current profile's `skills.json` from currently active skills.

### Export/import

Export should include skill names only when the profile has a `skills.json` manifest.

Import should restore the manifest into the profile directory, but should not install missing skills. Switching to the imported profile will warn for missing skills.

### List/health

`claudectx health <profile>` should report missing skills when the profile has a manifest.

Interactive list can show a small skill count later, but this should be lower priority than correct switching and rollback.

---

## Implementation plan

### Phase 1: Discovery and explicit model

Confirm how Claude Code discovers active global skills:

- directory presence under `~/.claude/skills/`
- a settings key
- plugin-managed registry
- another file

Document the finding in the PR. The rest of the implementation should be based on that actual behavior, not assumption.

### Phase 2: Add internal skill package

Add `internal/skills`.

Suggested types:

```go
type Manifest struct {
    Enabled []string `json:"enabled"`
}

type State struct {
    Active   []string
    Disabled []string
}
```

Suggested functions:

- `LoadManifest(path string) (*Manifest, error)`
- `SaveManifest(path string, manifest *Manifest) error`
- `ListActive() ([]string, error)`
- `ListDisabled() ([]string, error)`
- `SnapshotActive() (*Manifest, error)`
- `ApplyManifest(manifest *Manifest) (Result, error)`
- `ValidateSkillName(name string) error`

`Result` should include enabled, disabled, unchanged, and missing counts/names for output and tests.

### Phase 3: Extend paths

Add helpers in `internal/paths`:

- `SkillsDir()`
- `DisabledSkillsDir()`
- `ProfileSkillsFile(profileName string)`

Keep path logic centralized. Do not construct `~/.claude/skills` directly in command code.

### Phase 4: Extend profile loading/storage

Add an optional field to `profile.Profile`:

```go
Skills *skills.Manifest
```

Use a pointer so the code can distinguish:

- `nil`: profile does not manage skills
- non-nil with empty `Enabled`: profile manages skills and wants none active

Update `store.Save` and `store.Load`:

- save `skills.json` only when `prof.Skills != nil`
- load it when present
- do not require it for `Exists`
- preserve existing profiles without migration

Update `Profile.IsEmpty` to treat a non-nil skills manifest as meaningful.

### Phase 5: Backup and rollback

Extend `internal/backup` only for skill-aware operations.

Practical approach:

- backup active `~/.claude/skills/`
- backup disabled `~/.claude/.claudectx-disabled-skills/`
- restore both directories on rollback

If the directories are large, consider a later optimization. Correctness matters more for v1.

### Phase 6: Switch integration

In `cmd/switch.go`, after settings, `CLAUDE.md`, and MCP validation but before mutating skill state:

1. create backup as currently done
2. apply settings, `CLAUDE.md`, and MCP config
3. if `prof.Skills != nil`, apply skill manifest
4. on any failure, rollback the whole backup
5. update current/previous profile only after all mutations succeed

This keeps profile state honest: a profile should not be marked current if skill switching failed.

### Phase 7: Create and sync flags

Add opt-in flags only after the core internals are tested.

Recommended:

- `claudectx -n <name> --with-skills`
- `claudectx sync --skills`

If the CLI parser is too simple for this change, defer flags and support only manually authored `skills.json` in v1. That would still satisfy the core switch use case safely.

### Phase 8: Export/import

Update `internal/exporter`:

- add optional `Skills *skills.Manifest 'json:"skills,omitempty"'`
- keep export version compatible if possible
- if version bump is needed, accept old exports gracefully

Import behavior:

- save manifest into imported profile
- validate skill names
- do not install missing skills

### Phase 9: Health and docs

Update README:

- "What Gets Switched?"
- "What stays the same?"
- "How It Works" tree
- explain opt-in skill manifests
- explain missing skill warnings
- state that export/import includes skill names, not skill contents

Update health checks:

- valid manifest JSON
- invalid skill names
- missing desired skills

---

## Test plan

### `internal/skills`

| Test | What it verifies |
|------|-----------------|
| `TestValidateSkillName` | rejects empty names, separators, `.`, `..`, and hidden traversal |
| `TestLoadManifestMissingFile` | missing manifest can be represented distinctly from empty enabled list |
| `TestLoadManifestInvalidJSON` | invalid JSON returns a useful error |
| `TestSaveLoadManifestRoundTrip` | enabled skill names survive save/load |
| `TestListActiveSkills` | only directories are treated as skills |
| `TestApplyManifestMovesSkills` | desired skills are active and undesired active skills are disabled |
| `TestApplyManifestMissingSkillWarns` | missing skills appear in result but do not fail the switch |
| `TestApplyManifestEmptyEnabledDisablesAll` | empty manifest intentionally disables active skills |

### `internal/store`

| Test | What it verifies |
|------|-----------------|
| `TestSaveLoadProfileWithSkills` | `skills.json` is persisted and loaded |
| `TestLoadProfileWithoutSkillsLeavesNil` | old profiles do not manage skills |
| `TestProfileIsEmptyWithSkillsManifest` | a skills-only profile is not considered empty |

### `cmd/switch`

| Test | What it verifies |
|------|-----------------|
| `TestSwitchProfileWithoutSkillsManifestLeavesSkillsUntouched` | backward compatibility |
| `TestSwitchProfileWithSkillsManifestAppliesSkills` | manifest drives active skill state |
| `TestSwitchProfileSkillFailureRollsBack` | failed moves restore prior state |
| `TestSwitchProfileMissingSkillWarnsButSucceeds` | missing installed skill does not block switch |

### `internal/backup`

| Test | What it verifies |
|------|-----------------|
| `TestBackupRestoreSkills` | active and disabled skill dirs restore correctly |
| `TestBackupRestoreWithoutSkillsDirs` | absent dirs remain absent or empty after restore |

### `internal/exporter`

| Test | What it verifies |
|------|-----------------|
| `TestExportProfileIncludesSkillsManifest` | exports skill names when present |
| `TestExportProfileWithoutSkillsOmitsSkills` | old profiles export cleanly |
| `TestImportProfileWithSkillsManifest` | imported profile gets `skills.json` |
| `TestImportProfileDoesNotRequireInstalledSkills` | import remains portable |

---

## Open questions

1. How exactly does Claude Code determine which global skills are active?
2. Are skills always directories, or can they be files/symlinks?
3. Should claudectx manage symlinked skills, or warn and leave them untouched?
4. Should the disabled directory live at `~/.claude/.claudectx-disabled-skills/` or inside `~/.claude/profiles/`?
5. Should `--with-skills` be added in the first PR, or should v1 support only manually authored manifests?

---

## Non-goals for v1

- installing skills
- updating skills
- exporting full skill contents
- importing full skill contents
- resolving skill sources from names
- syncing plugin-managed skills unless Claude Code documents that mechanism
- deleting skills
- silently changing skills for profiles that do not have `skills.json`

---

## Suggested issue response

This is a good feature request and fits claudectx's goal of switching complete Claude Code contexts. I agree with the need, but I would implement it as an opt-in skills manifest rather than directly bundling or replacing `~/.claude/skills/`.

The safer v1 would store desired active skill names in `profiles/<name>/skills.json`. Profiles without that file would leave skills unchanged for backward compatibility. Switching to a profile with a skills manifest would enable installed matching skills, disable active skills not in the manifest, and warn for missing skills. Export/import would include skill names only, not skill contents.

That gives users profile-aware skills without turning claudectx into a skill package manager or risking accidental deletion/leakage of local skill content.

---

## Recommendation

Keep the issue open and implement it, but narrow the acceptance criteria:

1. Add opt-in `skills.json` manifests.
2. Leave skills untouched for profiles without manifests.
3. Move or registry-toggle active installed skills only after confirming Claude Code's actual skill activation mechanism.
4. Warn for missing skills.
5. Include skill names, not contents, in export/import.
6. Add backup/rollback coverage before enabling switch-time mutation.

This gives the feature real value while preserving claudectx's safety model.
