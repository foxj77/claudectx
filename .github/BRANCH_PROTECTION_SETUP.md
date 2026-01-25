# Branch Protection Setup Guide

This guide will help you enable branch protection on the `main` branch to improve your OpenSSF Scorecard.

## Steps to Enable Branch Protection

1. **Navigate to Repository Settings**
   - Go to https://github.com/foxj77/claudectx
   - Click on "Settings" tab
   - Click on "Branches" in the left sidebar

2. **Add Branch Protection Rule**
   - Click "Add branch protection rule"
   - In "Branch name pattern", enter: `main`

3. **Required Settings for OpenSSF Scorecard**

   Enable these options:

   **Protect matching branches:**
   - ✅ **Require a pull request before merging**
     - ✅ Require approvals: `1`
     - ✅ Dismiss stale pull request approvals when new commits are pushed
     - ✅ Require review from Code Owners (if you create a CODEOWNERS file)

   - ✅ **Require status checks to pass before merging**
     - ✅ Require branches to be up to date before merging
     - Add required status checks:
       - `test (ubuntu-latest, 1.21)`
       - `test (macos-latest, 1.21)`
       - `test (windows-latest, 1.21)`
       - `lint`

   - ✅ **Require conversation resolution before merging**

   - ✅ **Require signed commits** (recommended)

   - ✅ **Require linear history** (optional, keeps history clean)

   - ✅ **Do not allow bypassing the above settings**

   **Rules applied to everyone including administrators:**
   - ✅ **Include administrators**

4. **Click "Create" or "Save changes"**

## Optional: Set Up CODEOWNERS

Create a `.github/CODEOWNERS` file to automatically request reviews:

```
# Default code owners
* @foxj77

# Workflows require maintainer approval
/.github/ @foxj77

# Security-related files
/SECURITY.md @foxj77
/.github/dependabot.yml @foxj77
```

## Optional: Enable Signed Commits

To require signed commits:

1. Set up GPG signing locally:
   ```bash
   # Generate GPG key
   gpg --full-generate-key

   # List keys and copy the key ID
   gpg --list-secret-keys --keyid-format=long

   # Add to Git config
   git config --global user.signingkey YOUR_KEY_ID
   git config --global commit.gpgsign true

   # Add GPG key to GitHub
   gpg --armor --export YOUR_KEY_ID
   # Copy output and add at https://github.com/settings/keys
   ```

2. Or use GitHub's web-based commit signing (automatically enabled for web UI commits)

## Testing Branch Protection

After enabling, test by:

1. Creating a new branch:
   ```bash
   git checkout -b test-branch-protection
   echo "test" >> README.md
   git add README.md
   git commit -m "test: verify branch protection"
   git push -u origin test-branch-protection
   ```

2. Create a PR and verify you cannot merge without:
   - Passing CI checks
   - Required review approval

## Expected OpenSSF Scorecard Improvements

After enabling these settings, your scorecard should show improvements in:

- **Branch-Protection** (0/10 → 8-10/10)
- **Code-Review** (0/10 → 10/10)
- **Token-Permissions** (already addressed with `persist-credentials: false`)

## Allowing Admin Bypass (Not Recommended)

If you need to bypass occasionally (not recommended for OpenSSF score):

- Uncheck "Include administrators"
- You can then use "Merge without waiting for requirements" as an admin

**Note:** This will lower your Branch-Protection score.
