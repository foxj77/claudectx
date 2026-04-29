# Review: issue18.md proposal

**Reviewed:** 2026-04-28  
**Proposal:** [docs/issue18.md](issue18.md)  
**Issue:** https://github.com/foxj77/claudectx/issues/18

---

## Overall verdict

Sound proposal. The author understood the risks the original issue glossed over. The opt-in manifest model and the refusal to copy skill file contents into exports are the right calls. Recommend proceeding, with the caveats below.

---

## What the doc gets right

**Installed vs active distinction.** The original issue's simple `"skills": ["code-reviewer", "tdd"]` list on `Profile` would have led to either bundling arbitrary executable/prompt content in exports (a supply-chain risk) or silently doing nothing when a skill isn't installed. The manifest + move-not-copy model sidesteps both.

**`nil` vs empty `Enabled`.** Distinguishing a missing `skills.json` (do not manage skills) from one with an empty `enabled` list (actively disable all managed skills) is the right level of care. It means old profiles stay untouched automatically — correct backward-compatibility behavior for a safety-oriented tool.

**No automatic installs, no content in exports.** Both of these are correct non-goals for v1.

---

## Concerns and pushback

### 1. Phase 1 (discovery) is load-bearing and not yet answered

The entire filesystem model — move directories vs. toggle a registry key in `settings.json` — depends on how Claude Code actually discovers active skills. The doc says "verify this before coding" but then writes out a full directory-move implementation as if that's established fact.

**This must be answered empirically before any code is written.** If skills are toggled via a `settings.json` key, the disabled-skills directory model is unnecessary complexity.

### 2. `~/.claude/.claudectx-disabled-skills/` is an awkward location

Stashing disabled skills outside the profile directory means uninstalling claudectx or switching tools leaves orphaned content with no obvious owner. A per-profile subfolder (e.g. `~/.claude/profiles/work/disabled-skills/`) would give clearer ownership — though this is also contingent on the discovery answer above.

### 3. Cut `--with-skills` from v1

The doc wavers on this but the right call is to ship only manually authored `skills.json` in v1. Adding the flag introduces CLI surface area before the core behavior is validated. Document manual creation and defer the flag.

### 4. Backup of skill directories could be large

The doc acknowledges this but defers it. Worth at minimum adding a size guard or a `--dry-run` flag before v1 ships, since some skill directories could contain large assets.

---

## Recommendation

Keep the issue open. Before writing any code:

1. Confirm empirically how Claude Code discovers active global skills (directory presence, `settings.json` key, plugin registry, or other).
2. Document the finding in the PR description.
3. Adjust the filesystem model in `issue18.md` to match reality.

Then proceed with Phases 2–9 as written, minus the `--with-skills` flag for v1.
