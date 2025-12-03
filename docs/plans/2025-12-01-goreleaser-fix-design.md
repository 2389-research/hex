# GoReleaser Configuration Fix Design

**Date:** 2025-12-01
**Status:** Approved
**Platforms:** macOS only (Intel + Apple Silicon)

## Overview

Fix the GoReleaser configuration which currently references the wrong project ("toki" instead of "hex") and cannot build binaries due to misconfigured GitHub Actions workflow.

## Current Problems

1. `.goreleaser.yml` configured for "toki" project (wrong name)
2. Builds only for macOS but runs on Ubuntu runner
3. Cannot cross-compile CGO binaries from Linux to macOS
4. Homebrew integration not ready yet
5. References `cmd/toki` which doesn't exist

## Solution: Native macOS Builds via Matrix

Use GitHub Actions matrix strategy to build on actual macOS runners for each architecture.

### Platforms Supported

- macOS Intel (x86_64) - via `macos-13` runner
- macOS Apple Silicon (arm64) - via `macos-14` runner

### Build Strategy

Each runner builds natively for its architecture:
- No cross-compilation needed
- CGO works naturally
- Both runners upload artifacts to same GitHub Release
- GoReleaser merges artifacts automatically

## Configuration Changes

### 1. `.goreleaser.yml`

**Project naming:**
- Change `id: toki` → `id: hex`
- Change `binary: toki` → `binary: hex`
- Change `main: ./cmd/toki` → `main: ./cmd/hex`
- Update archive name template from `toki_` to `hex_`

**Build settings:**
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

**Remove:**
- Entire `brews:` section (Homebrew integration - save for later)

**Keep unchanged:**
- Checksum generation
- Changelog generation from commits
- Archive structure with README/LICENSE/docs

### 2. `.github/workflows/release.yml`

**Add matrix strategy:**
```yaml
jobs:
  release:
    strategy:
      matrix:
        include:
          - os: macos-13
            goarch: amd64
          - os: macos-14
            goarch: arm64

    runs-on: ${{ matrix.os }}
```

**Update steps:**
- Change tests: `go test ./... -race` → `go test -short ./...`
  - Faster execution
  - Skips slow VCR tests
  - More reliable in CI

**Remove:**
- Entire `publish-homebrew` job

### 3. `cmd/hex/root.go`

**Add version variables:**
```go
var (
    version = "1.0.0"    // existing
    commit  = "dev"      // new - filled by goreleaser
    date    = "unknown"  // new - filled by goreleaser
)
```

These get populated at build time via ldflags.

## Release Process

1. Developer creates git tag: `git tag v1.2.3 && git push --tags`
2. GitHub Actions triggers on tag push
3. Two jobs run in parallel:
   - `macos-13` builds Intel binary
   - `macos-14` builds Apple Silicon binary
4. Each job runs:
   - Checkout code
   - Setup Go 1.24.x
   - Run tests (short mode)
   - Run GoReleaser
5. GoReleaser on each runner:
   - Builds native binary with CGO
   - Creates tar.gz archive
   - Uploads to GitHub Release
6. Result: Single release with two artifacts:
   - `hex_v1.2.3_Darwin_x86_64.tar.gz`
   - `hex_v1.2.3_Darwin_arm64.tar.gz`
   - `checksums.txt`
   - Auto-generated changelog

## Future Enhancements

**Not included in this design:**
- Homebrew tap integration (needs HOMEBREW_TAP_TOKEN secret)
- Linux builds (would need separate matrix entries)
- Windows builds (CGO on Windows is complex)

These can be added incrementally after macOS releases are working.

## Testing Plan

1. Create test tag in fork or branch
2. Verify both matrix jobs run
3. Check that binaries are built for both architectures
4. Download and test each binary
5. Verify `hex --version` shows correct version/commit/date

## Files to Modify

1. `.goreleaser.yml` - Update project name and build config
2. `.github/workflows/release.yml` - Add matrix, remove Homebrew job
3. `cmd/hex/root.go` - Add commit and date variables
