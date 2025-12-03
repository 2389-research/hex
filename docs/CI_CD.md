# CI/CD Infrastructure

This document describes the continuous integration and deployment infrastructure for Hex.

## Overview

Hex uses GitHub Actions for CI/CD with the following workflows:

1. **Test Workflow** - Runs on every push and PR
2. **Release Workflow** - Runs on version tags
3. **GoReleaser** - Handles cross-platform builds

## Test Workflow

**File**: `.github/workflows/test.yml`

**Triggers**:
- Push to `main` branch
- Pull requests to `main` branch

**Jobs**:

### 1. Test Job

Runs tests across multiple platforms:

- **Matrix**: Ubuntu, macOS
- **Go Version**: 1.24.x
- **Steps**:
  1. Checkout code
  2. Set up Go with module caching
  3. Download and verify dependencies
  4. Run tests with race detection and coverage
  5. Upload coverage to Codecov
  6. Run `go vet` for static analysis
  7. Run `staticcheck` for additional linting

### 2. Lint Job

Runs code quality checks:

- **Platform**: Ubuntu
- **Steps**:
  1. Checkout code
  2. Set up Go
  3. Run golangci-lint with 5-minute timeout

### 3. Build Job

Verifies the binary builds and installs:

- **Platform**: Ubuntu
- **Steps**:
  1. Checkout code
  2. Set up Go
  3. Build binary
  4. Test installation with `go install`
  5. Verify version command works

## Release Workflow

**File**: `.github/workflows/release.yml`

**Triggers**:
- Tags matching `v*.*.*` (e.g., `v1.2.3`)

**Jobs**:

### 1. Release Job

Creates multi-platform release:

- **Steps**:
  1. Checkout with full history
  2. Set up Go 1.24.x
  3. Run full test suite
  4. Run GoReleaser to build and publish

**Artifacts Created**:
- Binaries for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64
- Archives: tar.gz (Unix), zip (Windows)
- Checksums file (SHA256)
- Debian/RPM/APK packages
- Docker images (ghcr.io/harper/hex)
- GitHub release with notes

### 2. Publish Homebrew Job

Updates Homebrew tap:

- **Depends on**: Release job completion
- **Steps**:
  1. Extract version from tag
  2. Trigger repository dispatch to `harper/homebrew-tap`

**Note**: Requires `HOMEBREW_TAP_TOKEN` secret in repository settings.

## GoReleaser Configuration

**File**: `.goreleaser.yml`

### Build Configuration

```yaml
builds:
  - id: hex
    main: ./cmd/hex
    binary: hex
    env:
      - CGO_ENABLED=0  # Static binaries
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w  # Strip debug info
      - -X github.com/2389-research/hex/internal/core.Version={{.Version}}
      - -X github.com/2389-research/hex/internal/core.Commit={{.ShortCommit}}
      - -X github.com/2389-research/hex/internal/core.Date={{.Date}}
```

### Archives

- **Naming**: `hex_v1.2.3_Darwin_x86_64.tar.gz`
- **Includes**: Binary, LICENSE, README.md, CHANGELOG.md, docs/
- **Format**: tar.gz (Unix), zip (Windows)

### Changelog

Automatic changelog generation with sections:
- Features (`feat:` commits)
- Bug Fixes (`fix:` commits)
- Other Changes

Excludes: docs, test, ci, chore commits

### Docker Images

Published to GitHub Container Registry (ghcr.io):
- `ghcr.io/harper/hex:latest`
- `ghcr.io/harper/hex:v1`
- `ghcr.io/harper/hex:v1.2`
- `ghcr.io/harper/hex:v1.2.3`

### Package Formats

- **Debian** (.deb)
- **RPM** (.rpm)
- **Alpine** (.apk)

Installed to: `/usr/bin/hex`

## Local Development

### Testing Workflows Locally

Install [act](https://github.com/nektos/act) to run GitHub Actions locally:

```bash
# Install act
brew install act

# Run test workflow
act -j test

# Run specific job
act -j lint
```

### Testing Release Build

```bash
# Install goreleaser
brew install goreleaser

# Test release build (no publish)
make snapshot

# Output in: dist/
```

### Building Binaries

```bash
# Current platform
make build

# All platforms (via goreleaser)
goreleaser build --snapshot --clean
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build for current platform |
| `make test` | Run all tests with race detection |
| `make test-coverage` | Generate coverage report |
| `make lint` | Run golangci-lint |
| `make verify` | Run all checks (fmt, vet, lint, test) |
| `make snapshot` | Build release snapshot locally |
| `make release` | Test full release build locally |

## Installation Methods

The CI/CD infrastructure supports multiple installation methods:

### 1. Install Script

**URL**: `https://raw.githubusercontent.com/harper/hex/main/install.sh`

**Features**:
- Auto-detects OS and architecture
- Downloads latest release
- Verifies SHA256 checksum
- Installs to appropriate directory
- Works on macOS, Linux, Windows (WSL)

**Usage**:
```bash
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash
```

### 2. Homebrew

**Tap**: `harper/tap`

**Maintained by**: GoReleaser (automatic updates)

**Usage**:
```bash
brew install harper/tap/hex
```

### 3. Go Install

**Usage**:
```bash
go install github.com/2389-research/hex/cmd/hex@latest
```

### 4. Pre-built Binaries

**Download from**: GitHub Releases page

**Steps**:
1. Download archive for your platform
2. Extract binary
3. Move to PATH directory
4. Verify: `hex --version`

### 5. Docker

**Registry**: GitHub Container Registry (ghcr.io)

**Usage**:
```bash
# Pull latest
docker pull ghcr.io/harper/hex:latest

# Run
docker run -it --rm ghcr.io/harper/hex:latest --help
```

### 6. Package Managers

**Debian/Ubuntu**:
```bash
wget https://github.com/2389-research/hex/releases/download/v1.2.3/hex_1.2.3_Linux_x86_64.deb
sudo dpkg -i hex_1.2.3_Linux_x86_64.deb
```

**Fedora/RHEL**:
```bash
wget https://github.com/2389-research/hex/releases/download/v1.2.3/hex_1.2.3_Linux_x86_64.rpm
sudo rpm -i hex_1.2.3_Linux_x86_64.rpm
```

**Alpine**:
```bash
wget https://github.com/2389-research/hex/releases/download/v1.2.3/hex_1.2.3_Linux_x86_64.apk
apk add --allow-untrusted hex_1.2.3_Linux_x86_64.apk
```

## Release Process

### Creating a Release

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with release notes
3. **Commit changes**:
   ```bash
   git add .
   git commit -m "chore: prepare v1.2.3 release"
   ```

4. **Create and push tag**:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin main
   git push origin v1.2.3
   ```

5. **Monitor GitHub Actions**:
   - Test workflow should pass
   - Release workflow will trigger automatically
   - Builds typically take 5-10 minutes

6. **Verify release**:
   - Check GitHub Releases page
   - Verify all platform binaries are present
   - Test install script
   - Verify Homebrew tap updated

### Pre-release Checklist

See [.github/RELEASE_CHECKLIST.md](.github/RELEASE_CHECKLIST.md) for detailed steps.

**Summary**:
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tag created and pushed

## Secrets Configuration

Required GitHub repository secrets:

| Secret | Purpose | Required For |
|--------|---------|-------------|
| `GITHUB_TOKEN` | GitHub API access | Release creation (auto-provided) |
| `HOMEBREW_TAP_TOKEN` | Update Homebrew tap | Homebrew formula updates |

### Setting Up Secrets

1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add required secrets

**Generating Homebrew Tap Token**:
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token (classic)
3. Scopes: `repo` (full control)
4. Add as `HOMEBREW_TAP_TOKEN` secret

## Monitoring

### GitHub Actions Dashboard

View workflow runs:
- Go to repository → Actions tab
- Filter by workflow name
- Check logs for failures

### Badges

Add to README.md:

```markdown
[![Test](https://github.com/2389-research/hex/workflows/Test/badge.svg)](https://github.com/2389-research/hex/actions/workflows/test.yml)
[![Release](https://github.com/2389-research/hex/workflows/Release/badge.svg)](https://github.com/2389-research/hex/actions/workflows/release.yml)
```

### Coverage Reports

Coverage uploaded to Codecov.io:
- View at: https://codecov.io/gh/harper/hex
- Badge: `[![codecov](https://codecov.io/gh/harper/hex/branch/main/graph/badge.svg)](https://codecov.io/gh/harper/hex)`

## Troubleshooting

### Test Failures

**Issue**: Tests fail on CI but pass locally

**Solutions**:
1. Ensure Go version matches CI (1.24.x)
2. Run with race detector: `go test -race ./...`
3. Check for time zone or platform-specific issues
4. Review CI logs for specific errors

### Release Failures

**Issue**: GoReleaser fails to build

**Solutions**:
1. Test locally: `make snapshot`
2. Verify `.goreleaser.yml` syntax
3. Check Go version compatibility
4. Review GoReleaser logs

### Install Script Issues

**Issue**: Install script fails on specific platform

**Solutions**:
1. Test locally: `bash -x install.sh`
2. Verify OS/arch detection logic
3. Check GitHub API rate limits
4. Test checksum verification

## Performance Optimization

### Caching

Go modules are cached between runs:
```yaml
- uses: actions/setup-go@v5
  with:
    cache: true
    cache-dependency-path: go.sum
```

### Matrix Strategy

Tests run in parallel across platforms:
```yaml
strategy:
  fail-fast: false
  matrix:
    os: [ubuntu-latest, macos-latest]
```

### Artifact Caching

GoReleaser caches build artifacts for faster rebuilds.

## Future Improvements

- [ ] Add Windows native runner (currently only WSL)
- [ ] Implement automated security scanning
- [ ] Add performance benchmarking
- [ ] Set up automated dependency updates (Dependabot)
- [ ] Add end-to-end integration tests
- [ ] Implement canary releases
- [ ] Add release notes automation

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Go Release Best Practices](https://golang.org/doc/install/source)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
