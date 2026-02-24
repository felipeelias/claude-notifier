---
name: release
description: Use when the user asks to create a new release, bump the version, publish, or ship a new version
disable-model-invocation: true
allowed-tools: Bash(gh *), Bash(git log *), Bash(git describe *)
---

# Creating a Release

Releases are created via GitHub Releases. Pushing a tag triggers goreleaser, which builds cross-platform binaries and updates the Homebrew tap.

## Current state

- Latest release: !`gh release list --limit 1 --json tagName,publishedAt --template '{{range .}}{{.tagName}} ({{.publishedAt}}){{end}}'`
- Commits since last release:

!`git log $(git describe --tags --abbrev=0)..HEAD --oneline`

## Steps

### 1. Verify CI is green on main

```bash
gh run list --branch main --limit 3
```

If any run is in progress, wait with `gh run watch <id> --exit-status`.

### 2. Determine next version

Apply conventional commit rules to the commits listed above:

- `feat:` commit present -> **minor** bump (e.g. v0.1.1 -> v0.2.0)
- Only `fix:`, `refactor:`, `perf:`, etc. -> **patch** bump (e.g. v0.1.1 -> v0.1.2)
- Breaking change (`BREAKING CHANGE:` or `!:`) -> **major** bump

Confirm the version with the user before proceeding.

### 3. Write release notes

Group commits by type into a short human-readable summary. Goreleaser auto-generates a detailed changelog from commits (excluding `docs:`, `test:`, `ci:` prefixes), so keep notes high-level.

### 4. Create the release

```bash
gh release create v<VERSION> --title "v<VERSION>" --notes "<NOTES>"
```

This pushes a tag, which triggers `.github/workflows/release.yml` -> goreleaser -> builds binaries + updates Homebrew tap.

### 5. Verify the release workflow

```bash
gh run list --limit 1
gh run watch <id> --exit-status
```

Confirm binaries appear on the release page:

```bash
gh release view v<VERSION>
```
