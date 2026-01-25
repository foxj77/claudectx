# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

### Option 1: GitHub Security Advisories (Preferred)

1. Navigate to the [Security Advisories](https://github.com/foxj77/claudectx/security/advisories) page
2. Click "Report a vulnerability"
3. Fill out the advisory form with details about the vulnerability

### Option 2: Email

Send an email to security@claudectx.dev with:

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested fixes (if available)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Investigation**: We will investigate and validate the reported vulnerability
- **Updates**: We will send you regular updates about our progress
- **Resolution**: Once resolved, we will publicly disclose the vulnerability (with credit to you, if desired)

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Varies based on severity (critical issues within 7 days, others within 30 days)

## Security Update Process

When a security vulnerability is fixed:

1. A security advisory will be published on GitHub
2. A new patch version will be released
3. The CHANGELOG will include security fix details
4. Users will be notified via GitHub releases

## Security Best Practices for Users

When using `claudectx`:

- **Verify downloads**: Always download from official sources (GitHub releases or Homebrew tap)
- **Check signatures**: Verify release checksums from the releases page
- **Keep updated**: Run the latest version to get security patches
- **Review permissions**: Understand what files `claudectx` accesses (`~/.claude/` directory)
- **Backup data**: Regular backups are created in `~/.claude/backups/`

## Scope

This security policy applies to:

- The `claudectx` CLI tool and its source code
- Official distribution channels (GitHub releases, Homebrew tap)
- CI/CD pipelines and release automation

## Security Features

`claudectx` implements several security features:

- **No network access**: All operations are local filesystem only
- **Atomic operations**: Configuration changes use atomic file operations
- **Automatic backups**: Every profile switch creates timestamped backups
- **Input validation**: Profile names and settings are validated before use
- **No credential storage**: No passwords, API keys, or tokens are stored by `claudectx`

## Attribution

We appreciate the security research community and will acknowledge reporters in our security advisories unless you prefer to remain anonymous.

---

Thank you for helping keep `claudectx` and its users safe!
