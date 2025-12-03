# GoReleaser Configuration Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix GoReleaser configuration to build macOS binaries (Intel + Apple Silicon) using native GitHub Actions runners.

**Architecture:** Replace cross-compilation approach with GitHub Actions matrix strategy using native macOS runners (macos-13 for Intel, macos-14 for Apple Silicon). Each runner builds natively with CGO support.

**Tech Stack:** GoReleaser v2, GitHub Actions, Go 1.24.x, modernc.org/sqlite (CGO)

---

## Task 1: Add version variables to cmd/hex/root.go

**Files:**
- Modify: `cmd/hex/root.go:22-24`

**Step 1: Read the current version variable location**

Run: `grep -n "var (" cmd/hex/root.go | head -1`
Expected: Line 22 shows `var (`

**Step 2: Add commit and date variables**

Edit `cmd/hex/root.go` at lines 22-24:

```go
var (
	// Version information
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"

	// Global flags
	printMode    bool
```

**Step 3: Verify syntax**

Run: `go build ./cmd/hex`
Expected: Successful build

**Step 4: Commit**

```bash
git add cmd/hex/root.go
git commit -m "feat: add commit and date variables for build-time injection

- Add commit and date vars alongside existing version var
- Default to dev/unknown for local builds
- Will be populated by goreleaser ldflags

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Update .goreleaser.yml project references

**Files:**
- Modify: `.goreleaser.yml`

**Step 1: Update builds section (lines 11-31)**

Replace the entire `builds:` section:

```yaml
builds:
  - id: hex
    binary: hex
    main: ./cmd/hex
    env:
      - CGO_ENABLED=1
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
    tags:
      - sqlite_omit_load_extension
```

**Step 2: Update archives section (lines 33-49)**

Replace the entire `archives:` section:

```yaml
archives:
  - id: hex
    formats: [tar.gz]
    name_template: >-
      hex_
      {{- .Version }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - README.md
      - LICENSE*
      - docs/**/*
```

**Step 3: Remove Homebrew section (lines 79-95)**

Delete the entire `brews:` section from line 79 to the end of file.

**Step 4: Verify YAML syntax**

Run: `yamllint .goreleaser.yml 2>/dev/null || echo "YAML is valid"`
Expected: No errors

**Step 5: Update header comment (lines 1-3)**

Replace:

```yaml
# GoReleaser configuration for toki
# Note: Due to CGO requirements for SQLite, this config builds for the current platform only
# Multi-platform releases require running goreleaser on each platform (handled by CI matrix)
```

With:

```yaml
# GoReleaser configuration for hex
# Builds for macOS (Intel + Apple Silicon) using native GitHub Actions runners
# CGO is required for SQLite support via modernc.org/sqlite
```

**Step 6: Commit**

```bash
git add .goreleaser.yml
git commit -m "fix: update goreleaser config for hex project

- Change project name from toki to hex
- Update binary name and main path
- Remove Homebrew integration (not ready yet)
- Update header comments
- Keep macOS-only builds for now

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: Update GitHub Actions release workflow

**Files:**
- Modify: `.github/workflows/release.yml`

**Step 1: Add matrix strategy to release job (after line 14)**

Replace lines 13-15:

```yaml
jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
```

With:

```yaml
jobs:
  release:
    name: Release
    strategy:
      matrix:
        include:
          - os: macos-13
            goarch: amd64
          - os: macos-14
            goarch: arm64
    runs-on: ${{ matrix.os }}
```

**Step 2: Update test command (line 30)**

Replace:

```yaml
      - name: Run tests
        run: go test ./... -race
```

With:

```yaml
      - name: Run tests
        run: go test -short ./...
```

**Step 3: Add GOARCH environment variable (line 38)**

Add to the `env:` section of the GoReleaser step:

```yaml
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GOARCH: ${{ matrix.goarch }}
```

**Step 4: Remove publish-homebrew job (lines 41-59)**

Delete the entire `publish-homebrew:` job section.

**Step 5: Verify workflow syntax**

Run: `cat .github/workflows/release.yml | grep "runs-on:"`
Expected: Shows `runs-on: ${{ matrix.os }}`

**Step 6: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "fix: use matrix strategy for native macOS builds

- Add matrix with macos-13 (Intel) and macos-14 (Apple Silicon)
- Change from ubuntu-latest to native macOS runners
- Update tests to use -short flag (skip slow VCR tests)
- Remove Homebrew publishing job
- Pass GOARCH via environment

This enables proper CGO compilation on each platform without
cross-compilation complexity.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Test local goreleaser build

**Files:**
- None (verification only)

**Step 1: Install goreleaser if needed**

Run: `which goreleaser || brew install goreleaser`
Expected: Path to goreleaser binary

**Step 2: Run goreleaser in snapshot mode**

Run: `goreleaser build --snapshot --clean --single-target`
Expected: Builds successfully for current platform

**Step 3: Check output binary**

Run: `ls -lh dist/hex_darwin_*/hex`
Expected: Binary file exists

**Step 4: Test binary runs**

Run: `dist/hex_darwin_*/hex --version`
Expected: Shows version output (dev build)

**Step 5: Clean up**

Run: `rm -rf dist/`
Expected: dist directory removed

**Step 6: Document verification**

No commit needed - this is verification only.

---

## Task 5: Create test documentation

**Files:**
- Create: `docs/release-testing.md`

**Step 1: Create release testing guide**

```markdown
# Release Testing Guide

## Local Testing

Before creating a release tag, test the build locally:

\`\`\`bash
# Install goreleaser
brew install goreleaser

# Test snapshot build
goreleaser build --snapshot --clean --single-target

# Verify binary works
dist/hex_darwin_*/hex --version

# Clean up
rm -rf dist/
\`\`\`

## Creating a Release

1. Ensure all changes are committed and pushed
2. Create and push a tag:
   \`\`\`bash
   git tag v1.0.1
   git push origin v1.0.1
   \`\`\`

3. GitHub Actions will automatically:
   - Build on macos-13 (Intel)
   - Build on macos-14 (Apple Silicon)
   - Run tests on both
   - Create GitHub Release
   - Upload both binaries

4. Monitor the workflow:
   - https://github.com/harperreed/hex/actions

## Release Artifacts

Each release includes:
- `hex_vX.Y.Z_Darwin_x86_64.tar.gz` - Intel binary
- `hex_vX.Y.Z_Darwin_arm64.tar.gz` - Apple Silicon binary
- `checksums.txt` - SHA256 checksums
- Auto-generated changelog from commits

## Verifying a Release

\`\`\`bash
# Download the appropriate binary
curl -L -O https://github.com/harperreed/hex/releases/download/v1.0.1/hex_v1.0.1_Darwin_arm64.tar.gz

# Verify checksum
curl -L -O https://github.com/harperreed/hex/releases/download/v1.0.1/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing

# Extract and test
tar xzf hex_v1.0.1_Darwin_arm64.tar.gz
./hex --version
\`\`\`

## Troubleshooting

**Build fails on one platform:**
- Check the specific platform's Actions log
- CGO issues usually appear as "undefined reference" errors
- Verify Go version matches both runners

**Tests fail:**
- Check if VCR cassettes are causing issues
- Use `-short` flag to skip integration tests
- Verify database migrations are compatible

**Binary won't run:**
- Check macOS Gatekeeper: `xattr -d com.apple.quarantine hex`
- Verify architecture: `file hex`
- Check dynamic libraries: `otool -L hex`
```

**Step 2: Commit documentation**

```bash
git add docs/release-testing.md
git commit -m "docs: add release testing guide

- Document local testing with goreleaser
- Explain release creation process
- Add verification steps
- Include troubleshooting tips

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: Final verification

**Files:**
- None (verification only)

**Step 1: Verify all commits**

Run: `git log --oneline -5`
Expected: Shows 5 commits for this work

**Step 2: Verify files changed**

Run: `git diff HEAD~5 --stat`
Expected: Shows changes to:
- cmd/hex/root.go
- .goreleaser.yml
- .github/workflows/release.yml
- docs/release-testing.md

**Step 3: Run tests**

Run: `go test -short ./...`
Expected: All tests pass

**Step 4: Run pre-commit hooks**

Run: `pre-commit run --all-files`
Expected: All hooks pass

**Step 5: Verify goreleaser config**

Run: `goreleaser check`
Expected: Configuration is valid

**Step 6: Ready to push**

All verification complete. Ready to push to GitHub.

---

## Post-Implementation

After pushing these changes:

1. **Do NOT create a release tag yet**
2. Wait for confirmation that the workflow changes look correct
3. Consider testing with a pre-release tag first (v0.0.0-test)
4. Document any issues found during first release

## Future Enhancements

Not included in this plan (can be added later):

1. **Homebrew integration**
   - Requires HOMEBREW_TAP_TOKEN secret
   - Re-add brews: section to .goreleaser.yml
   - Re-add publish-homebrew job to workflow

2. **Linux builds**
   - Add ubuntu-latest with GOARCH matrix
   - Test CGO compilation on Linux
   - May need additional build tags

3. **Windows builds**
   - Add windows-latest runner
   - CGO on Windows is complex
   - May need MinGW toolchain
