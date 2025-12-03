# Release Testing Guide

## Local Testing

Before creating a release tag, test the build locally:

```bash
# Install goreleaser
brew install goreleaser

# Test snapshot build
goreleaser build --snapshot --clean --single-target

# Verify binary works
dist/hex_darwin_*/hex --version

# Clean up
rm -rf dist/
```

## Creating a Release

1. Ensure all changes are committed and pushed
2. Create and push a tag:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```

3. GitHub Actions will automatically:
   - Build on ubuntu-latest using pure Go cross-compilation
   - Build for both Intel (amd64) and Apple Silicon (arm64)
   - Run tests with `-short` flag
   - Create GitHub Release
   - Upload both macOS binaries

4. Monitor the workflow:
   - https://github.com/harperreed/hex/actions

## Release Artifacts

Each release includes:
- `hex_vX.Y.Z_Darwin_x86_64.tar.gz` - Intel binary
- `hex_vX.Y.Z_Darwin_arm64.tar.gz` - Apple Silicon binary
- `checksums.txt` - SHA256 checksums
- Auto-generated changelog from commits

## Architecture Notes

**Pure Go Cross-Compilation:**
- Uses `modernc.org/sqlite` which is a pure Go SQLite implementation
- No CGO required (`CGO_ENABLED=0`)
- Single ubuntu-latest runner builds for both architectures
- Fast, simple, and reliable cross-compilation

**Why This Works:**
- `modernc.org/sqlite` is a transpiled version of SQLite in pure Go
- No C compiler or platform-specific tooling needed
- Consistent builds across all platforms
- Smaller binaries due to static linking

## Verifying a Release

```bash
# Download the appropriate binary
curl -L -O https://github.com/harperreed/hex/releases/download/v1.0.1/hex_v1.0.1_Darwin_arm64.tar.gz

# Verify checksum
curl -L -O https://github.com/harperreed/hex/releases/download/v1.0.1/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing

# Extract and test
tar xzf hex_v1.0.1_Darwin_arm64.tar.gz
./hex --version
```

## Troubleshooting

**Build fails:**
- Check the Actions log for the ubuntu-latest runner
- Verify Go version matches (1.24.x)
- Ensure `go.mod` is up to date

**Tests fail:**
- Check if VCR cassettes are causing issues
- Use `-short` flag to skip integration tests
- Verify database migrations are compatible

**Binary won't run on macOS:**
- Check macOS Gatekeeper: `xattr -d com.apple.quarantine hex`
- Verify architecture: `file hex` (should show Mach-O 64-bit executable)
- Check it's statically linked: `otool -L hex` (minimal dependencies)

**Cross-compilation issues:**
- Verify `CGO_ENABLED=0` in `.goreleaser.yml`
- Check that no CGO-requiring packages were introduced
- Ensure `modernc.org/sqlite` is the SQLite driver (not `mattn/go-sqlite3`)

## Future Enhancements

Not included yet (can be added later):

1. **Linux builds**
   - Add `linux` to `goos` in `.goreleaser.yml`
   - Pure Go cross-compilation makes this trivial

2. **Windows builds**
   - Add `windows` to `goos` in `.goreleaser.yml`
   - Pure Go eliminates traditional Windows CGO complexity

3. **Homebrew integration**
   - Requires HOMEBREW_TAP_TOKEN secret
   - Add brews: section to `.goreleaser.yml`
   - Add publish-homebrew job to workflow
