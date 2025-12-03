# Package Verification Plan - Hex v1.0

**Date:** 2025-11-28
**Purpose:** Verify all 6 distribution channels work correctly before v1.0 release

---

## Distribution Channels

Hex v1.0 ships through 6 distribution channels. Each must be verified independently.

| # | Channel | Platform | Priority | Verification Status |
|---|---------|----------|----------|---------------------|
| 1 | Homebrew | macOS/Linux | High | ⏳ Pending |
| 2 | Install Scripts | All | High | ⏳ Pending |
| 3 | Docker Images | All | Medium | ⏳ Pending |
| 4 | Binary Releases | All | High | ⏳ Pending |
| 5 | Linux Packages | Linux | Medium | ⏳ Pending |
| 6 | Go Install | All | Low | ⏳ Pending |

---

## 1. Homebrew Verification

**Platform:** macOS, Linux
**Repository:** harper/homebrew-tap
**Formula:** hex.rb

### Prerequisites
```bash
brew update
brew tap harper/tap
```

### Installation Test
```bash
# Clean install
brew uninstall harper/tap/hex 2>/dev/null || true
brew install harper/tap/hex

# Verify version
hex --version | grep "1.0.0"

# Verify binary location
which hex
# Expected: /opt/homebrew/bin/hex (Apple Silicon) or /usr/local/bin/hex (Intel)
```

### Functional Test
```bash
# Test print mode
hex --print "What is 2+2?" 2>&1 | head -5

# Test setup (skip if API key present)
hex doctor

# Test interactive mode (manual)
hex "Hello Claude"
# Should launch TUI, type ctrl+c to exit
```

### Upgrade Test
```bash
# Simulate upgrade
brew upgrade harper/tap/hex

# Verify new version
hex --version
```

### Uninstall Test
```bash
# Clean removal
brew uninstall harper/tap/hex

# Verify binary removed
which hex
# Expected: hex not found
```

**Success Criteria:**
- ✅ Clean install works
- ✅ Version is 1.0.0
- ✅ Binary in correct PATH location
- ✅ Print mode works
- ✅ Interactive TUI launches
- ✅ Upgrade preserves config
- ✅ Uninstall removes binary

---

## 2. Install Scripts Verification

**Platforms:** macOS, Linux (curl), Windows (PowerShell)

### Unix/Linux Install Script

```bash
# Download and inspect
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh > /tmp/install.sh
cat /tmp/install.sh | head -20

# Execute install
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash

# Verify installation
hex --version | grep "1.0.0"
which hex
# Expected: /usr/local/bin/hex or ~/.local/bin/hex
```

### Windows PowerShell Script

```powershell
# Run as Administrator
iwr -useb https://raw.githubusercontent.com/harper/hex/main/install.ps1 | iex

# Verify installation
hex --version
# Expected output: hex version 1.0.0

# Check PATH
where.exe hex
# Expected: C:\Program Files\hex\hex.exe or similar
```

### Functional Test (Both Platforms)
```bash
# Test basic functionality
hex --print "Hello world"

# Verify config directory created
ls ~/.hex/
# Expected: config.yaml or empty directory
```

**Success Criteria:**
- ✅ Script downloads without errors
- ✅ Binary installed to correct location
- ✅ Version is 1.0.0
- ✅ Binary is executable
- ✅ Config directory created
- ✅ Script idempotent (can run twice safely)

---

## 3. Docker Images Verification

**Registry:** ghcr.io/harper/hex
**Tags:** latest, 1.0.0, 1.0

### Pull and Inspect
```bash
# Pull latest image
docker pull ghcr.io/harper/hex:latest

# Inspect image
docker inspect ghcr.io/harper/hex:latest | jq '.[0].Config.Labels'

# Check size
docker images ghcr.io/harper/hex:latest
# Expected: < 50MB
```

### Run Tests
```bash
# Test version
docker run --rm ghcr.io/harper/hex:latest --version
# Expected: hex version 1.0.0

# Test print mode (needs API key)
docker run --rm \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  ghcr.io/harper/hex:latest \
  --print "What is 2+2?"

# Test with mounted config
docker run --rm \
  -v ~/.hex:/root/.hex \
  ghcr.io/harper/hex:latest \
  doctor
```

### Tag Verification
```bash
# Verify all tags point to same image
docker pull ghcr.io/harper/hex:1.0.0
docker pull ghcr.io/harper/hex:1.0
docker pull ghcr.io/harper/hex:latest

# Compare image IDs (should be identical)
docker images ghcr.io/harper/hex --format "{{.Tag}}\t{{.ID}}"
```

**Success Criteria:**
- ✅ Image pulls successfully
- ✅ Version is 1.0.0
- ✅ Image size reasonable (< 50MB)
- ✅ All tags exist and point to same image
- ✅ Print mode works with API key
- ✅ Config mounting works
- ✅ Labels include version and commit SHA

---

## 4. Binary Releases Verification

**Location:** GitHub Releases page
**URL:** https://github.com/harper/hex/releases/tag/v1.0.0

### Platforms to Test
- macOS Intel (darwin_amd64)
- macOS Apple Silicon (darwin_arm64)
- Linux x86_64 (linux_amd64)
- Linux ARM64 (linux_arm64)
- Windows x86_64 (windows_amd64.exe)

### Download and Extract (Example: macOS ARM64)
```bash
# Download
VERSION=1.0.0
PLATFORM=darwin_arm64
curl -L -o hex.tar.gz \
  "https://github.com/harper/hex/releases/download/v${VERSION}/hex_${VERSION}_${PLATFORM}.tar.gz"

# Extract
tar -xzf hex.tar.gz
chmod +x hex

# Verify
./hex --version
# Expected: hex version 1.0.0
```

### Checksum Verification
```bash
# Download checksums
curl -L -o checksums.txt \
  "https://github.com/harper/hex/releases/download/v${VERSION}/checksums.txt"

# Verify binary
shasum -a 256 -c checksums.txt --ignore-missing
# Expected: hex_1.0.0_darwin_arm64.tar.gz: OK
```

### Signature Verification (if GPG signing enabled)
```bash
# Download signature
curl -L -o hex.tar.gz.sig \
  "https://github.com/harper/hex/releases/download/v${VERSION}/hex_${VERSION}_${PLATFORM}.tar.gz.sig"

# Verify signature
gpg --verify hex.tar.gz.sig hex.tar.gz
```

### Test Each Platform
```bash
# Test binary functionality
./hex --print "Test"
./hex doctor
./hex --version
```

**Success Criteria:**
- ✅ All platform binaries present
- ✅ Checksums.txt present and valid
- ✅ Each binary runs and shows version 1.0.0
- ✅ Binaries are statically linked (no missing .so errors)
- ✅ File permissions correct (executable)
- ✅ Archive format correct (.tar.gz for Unix, .zip for Windows)

---

## 5. Linux Packages Verification

**Formats:** .deb (Debian/Ubuntu), .rpm (RHEL/Fedora), .apk (Alpine)

### Debian/Ubuntu (.deb)

```bash
# Download package
wget https://github.com/harper/hex/releases/download/v1.0.0/hex_1.0.0_amd64.deb

# Inspect package
dpkg -I hex_1.0.0_amd64.deb
dpkg -c hex_1.0.0_amd64.deb

# Install
sudo dpkg -i hex_1.0.0_amd64.deb

# Verify installation
which hex
hex --version
dpkg -l | grep hex

# Test functionality
hex --print "Test"

# Uninstall
sudo dpkg -r hex
```

### RHEL/Fedora (.rpm)

```bash
# Download package
wget https://github.com/harper/hex/releases/download/v1.0.0/hex_1.0.0_x86_64.rpm

# Inspect package
rpm -qip hex_1.0.0_x86_64.rpm
rpm -qlp hex_1.0.0_x86_64.rpm

# Install
sudo rpm -i hex_1.0.0_x86_64.rpm

# Verify installation
which hex
hex --version
rpm -qa | grep hex

# Test functionality
hex --print "Test"

# Uninstall
sudo rpm -e hex
```

### Alpine (.apk)

```bash
# Download package
wget https://github.com/harper/hex/releases/download/v1.0.0/hex_1.0.0_x86_64.apk

# Install
sudo apk add --allow-untrusted hex_1.0.0_x86_64.apk

# Verify
which hex
hex --version

# Uninstall
sudo apk del hex
```

**Success Criteria:**
- ✅ Package metadata correct (version, maintainer, description)
- ✅ Binary installed to /usr/bin or /usr/local/bin
- ✅ Package manager recognizes installed package
- ✅ Version command shows 1.0.0
- ✅ Package uninstalls cleanly
- ✅ Dependencies declared correctly

---

## 6. Go Install Verification

**Method:** `go install`
**Requires:** Go 1.24.9+

### Installation Test
```bash
# Clean environment
rm -f $(go env GOPATH)/bin/hex

# Install from source
go install github.com/harper/hex/cmd/hex@v1.0.0

# Verify installation
hex --version
# Expected: hex version 1.0.0 (or may show (devel) - acceptable)

which hex
# Expected: $GOPATH/bin/hex
```

### Build from Source
```bash
# Clone repository
git clone https://github.com/harper/hex.git
cd hex
git checkout v1.0.0

# Build
go build -o hex ./cmd/hex

# Verify
./hex --version
```

### Functional Test
```bash
# Test binary
hex --print "Test"
hex doctor
```

**Success Criteria:**
- ✅ `go install` succeeds without errors
- ✅ Binary placed in $GOPATH/bin
- ✅ Version is 1.0.0 (or (devel) acceptable)
- ✅ Build from source succeeds
- ✅ Tests pass: `go test ./...`

---

## Common Verification Tests

Run these tests for **every distribution channel** after installation:

### 1. Version Check
```bash
hex --version
# Expected: hex version 1.0.0
```

### 2. Help Command
```bash
hex --help
# Should display usage information
```

### 3. Doctor Command
```bash
hex doctor
# Should check configuration and dependencies
```

### 4. Print Mode (requires API key)
```bash
export ANTHROPIC_API_KEY=sk-ant-...
hex --print "What is 2+2?"
# Should return response about 4
```

### 5. Interactive Mode (manual)
```bash
hex "Test prompt"
# Should launch TUI
# Press Ctrl+C to exit
```

### 6. Config Directory
```bash
ls -la ~/.hex/
# Should exist after first run
```

### 7. Database Creation
```bash
hex "Test" # Run once
ls ~/.hex/hex.db
# Database should exist
```

---

## Automated Verification Script

Create `scripts/verify-packages.sh`:

```bash
#!/bin/bash
set -e

echo "=== Hex v1.0 Package Verification ==="

# 1. Homebrew
echo "Testing Homebrew..."
brew install harper/tap/hex
hex --version | grep "1.0.0" || exit 1
brew uninstall harper/tap/hex

# 2. Install Script
echo "Testing install script..."
curl -sSL https://raw.githubusercontent.com/harper/hex/main/install.sh | bash
hex --version | grep "1.0.0" || exit 1

# 3. Docker
echo "Testing Docker..."
docker run --rm ghcr.io/harper/hex:latest --version | grep "1.0.0" || exit 1

# 4. Binary Release
echo "Testing binary release..."
VERSION=1.0.0
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m)
curl -L -o /tmp/hex.tar.gz \
  "https://github.com/harper/hex/releases/download/v${VERSION}/hex_${VERSION}_${PLATFORM}.tar.gz"
tar -xzf /tmp/hex.tar.gz -C /tmp
/tmp/hex --version | grep "1.0.0" || exit 1

# 5. Go Install
echo "Testing go install..."
go install github.com/harper/hex/cmd/hex@v1.0.0
$(go env GOPATH)/bin/hex --version || exit 1

echo "✅ All packages verified successfully!"
```

---

## Pre-Release Checklist

Before tagging v1.0.0 and creating release:

### Code
- [ ] All tests pass: `go test ./...`
- [ ] No golangci-lint errors (warnings acceptable)
- [ ] Security audit clean (Go 1.24.9+)
- [ ] Version bumped in all files

### Documentation
- [ ] README.md updated for v1.0
- [ ] CHANGELOG.md includes v1.0 changes
- [ ] RELEASE_NOTES.md written
- [ ] SECURITY_AUDIT.md reviewed

### GitHub Release
- [ ] Tag created: `git tag v1.0.0`
- [ ] Tag pushed: `git push origin v1.0.0`
- [ ] GitHub Actions build completes
- [ ] Release notes added to GitHub release
- [ ] All artifacts attached to release

### Distribution Channels
- [ ] Homebrew formula updated
- [ ] Install scripts tested
- [ ] Docker images built and pushed
- [ ] Binary releases generated
- [ ] Linux packages built
- [ ] `go install` path verified

### Manual Verification
- [ ] At least 2 platforms tested (e.g., macOS + Linux)
- [ ] Interactive mode works
- [ ] Tool execution works (Read, Write, Bash)
- [ ] MCP integration works
- [ ] Config persists across runs

---

## Post-Release Verification

After v1.0.0 release:

1. **Wait 1 hour** for package propagation
2. **Run verification script** on clean VMs
3. **Test user-reported installation methods**
4. **Monitor GitHub issues** for installation problems
5. **Update documentation** if issues found

---

## Rollback Plan

If critical issues discovered:

1. **Unpublish release** (if possible)
2. **Revert Homebrew formula** to previous version
3. **Delete Docker tags** (or add `:broken` suffix)
4. **Mark GitHub release as pre-release**
5. **Post issue** explaining rollback
6. **Fix issues** and re-release as v1.0.1

---

## Success Metrics

v1.0 release is successful if:

- ✅ All 6 distribution channels install correctly
- ✅ Version shows 1.0.0 on all platforms
- ✅ < 5 installation-related GitHub issues in first week
- ✅ Docker image pulls without errors
- ✅ Homebrew formula audit passes
- ✅ No security vulnerabilities (govulncheck clean)

---

**Last Updated:** 2025-11-28
**Next Review:** Before v1.0.0 git tag
