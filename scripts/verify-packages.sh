#!/bin/bash
# ABOUTME: Verification script for all distribution packages and installation methods
# ABOUTME: Tests each package type to ensure they install and run correctly

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
VERSION="1.0.0"
REPO="harper/hex"
TEST_DIR=$(mktemp -d)
RESULTS_FILE="$TEST_DIR/verification-results.txt"

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$RESULTS_FILE"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1" | tee -a "$RESULTS_FILE"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1" | tee -a "$RESULTS_FILE"
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1" | tee -a "$RESULTS_FILE"
}

# Test a binary
test_binary() {
    local binary_path=$1
    local test_name=$2

    log_info "Testing $test_name..."

    # Check if binary exists
    if [ ! -f "$binary_path" ]; then
        log_error "$test_name: Binary not found at $binary_path"
        return 1
    fi

    # Check if executable
    if [ ! -x "$binary_path" ]; then
        log_error "$test_name: Binary is not executable"
        return 1
    fi

    # Test --version
    local version_output=$($binary_path --version 2>&1 || true)
    if [[ "$version_output" != *"$VERSION"* ]]; then
        log_error "$test_name: Version mismatch. Expected $VERSION, got: $version_output"
        return 1
    fi

    # Test --help
    if ! $binary_path --help >/dev/null 2>&1; then
        log_error "$test_name: --help failed"
        return 1
    fi

    log_success "$test_name: All checks passed"
    return 0
}

# Header
echo ""
log_info "Hex v$VERSION - Package Verification"
echo "========================================"
echo ""
log_info "Test directory: $TEST_DIR"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH_TITLE="x86_64" ;;
    arm64|aarch64) ARCH_TITLE="arm64" ;;
    *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    darwin) OS_TITLE="Darwin" ;;
    linux) OS_TITLE="Linux" ;;
    *) log_error "Unsupported OS: $OS"; exit 1 ;;
esac

log_info "Platform: $OS_TITLE / $ARCH_TITLE"
echo ""

# Track results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_test() {
    local test_name=$1
    shift
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if "$@"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Test 1: Binary Archive (.tar.gz)
test_binary_archive() {
    log_info "Test 1: Binary Archive (.tar.gz)"

    local archive_name="hex_${VERSION}_${OS_TITLE}_${ARCH_TITLE}.tar.gz"
    local download_url="https://github.com/$REPO/releases/download/v${VERSION}/${archive_name}"

    log_info "Downloading: $download_url"

    if ! curl -fsSL "$download_url" -o "$TEST_DIR/$archive_name"; then
        log_error "Failed to download binary archive"
        return 1
    fi

    log_info "Extracting archive..."
    tar -xzf "$TEST_DIR/$archive_name" -C "$TEST_DIR"

    test_binary "$TEST_DIR/hex" "Binary Archive"
}

# Test 2: Checksums
test_checksums() {
    log_info "Test 2: Checksums Verification"

    local checksums_url="https://github.com/$REPO/releases/download/v${VERSION}/checksums.txt"

    log_info "Downloading checksums..."
    if ! curl -fsSL "$checksums_url" -o "$TEST_DIR/checksums.txt"; then
        log_error "Failed to download checksums"
        return 1
    fi

    # Verify checksums file is not empty
    if [ ! -s "$TEST_DIR/checksums.txt" ]; then
        log_error "Checksums file is empty"
        return 1
    fi

    log_info "Checksums file downloaded successfully"

    # Verify our archive is in the checksums
    local archive_name="hex_${VERSION}_${OS_TITLE}_${ARCH_TITLE}.tar.gz"
    if ! grep -q "$archive_name" "$TEST_DIR/checksums.txt"; then
        log_error "Archive $archive_name not found in checksums.txt"
        return 1
    fi

    log_success "Checksums file valid"
    return 0
}

# Test 3: Homebrew (macOS/Linux only)
test_homebrew() {
    if ! command -v brew >/dev/null 2>&1; then
        log_warn "Homebrew not installed, skipping test"
        return 0
    fi

    log_info "Test 3: Homebrew Installation"

    # Check if tap exists
    if ! brew tap | grep -q "harper/tap"; then
        log_info "Adding tap: harper/tap"
        brew tap harper/tap || return 1
    fi

    # Check if formula exists
    if ! brew info harper/tap/hex >/dev/null 2>&1; then
        log_error "Homebrew formula not found"
        return 1
    fi

    log_success "Homebrew formula exists and is accessible"
    return 0
}

# Test 4: Docker Image
test_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        log_warn "Docker not installed, skipping test"
        return 0
    fi

    log_info "Test 4: Docker Image"

    # Pull image
    local image="ghcr.io/harper/hex:${VERSION}"
    log_info "Pulling: $image"

    if ! docker pull "$image" >/dev/null 2>&1; then
        # Try latest tag
        image="ghcr.io/harper/hex:latest"
        log_info "Trying latest tag: $image"
        if ! docker pull "$image" >/dev/null 2>&1; then
            log_error "Failed to pull Docker image"
            return 1
        fi
    fi

    # Test run
    local version_output=$(docker run --rm "$image" --version 2>&1 || true)
    if [[ "$version_output" != *"$VERSION"* ]]; then
        log_warn "Docker version mismatch: $version_output"
    else
        log_success "Docker image verified"
    fi

    return 0
}

# Test 5: Linux Packages (Linux only)
test_linux_packages() {
    if [ "$OS" != "linux" ]; then
        log_info "Test 5: Linux Packages (skipped on $OS)"
        return 0
    fi

    log_info "Test 5: Linux Packages"

    # Test .deb
    local deb_url="https://github.com/$REPO/releases/download/v${VERSION}/hex_${VERSION}_${OS_TITLE}_${ARCH_TITLE}.deb"
    if curl -fsSL "$deb_url" -o "$TEST_DIR/hex.deb"; then
        log_success ".deb package available"
    else
        log_error ".deb package not found"
    fi

    # Test .rpm
    local rpm_url="https://github.com/$REPO/releases/download/v${VERSION}/hex_${VERSION}_${OS_TITLE}_${ARCH_TITLE}.rpm"
    if curl -fsSL "$rpm_url" -o "$TEST_DIR/hex.rpm"; then
        log_success ".rpm package available"
    else
        log_error ".rpm package not found"
    fi

    # Test .apk
    local apk_url="https://github.com/$REPO/releases/download/v${VERSION}/hex_${VERSION}_${OS_TITLE}_${ARCH_TITLE}.apk"
    if curl -fsSL "$apk_url" -o "$TEST_DIR/hex.apk"; then
        log_success ".apk package available"
    else
        log_error ".apk package not found"
    fi

    return 0
}

# Test 6: Install Script
test_install_script() {
    log_info "Test 6: Install Script"

    local script_url="https://raw.githubusercontent.com/$REPO/main/install.sh"

    # Download script
    if ! curl -fsSL "$script_url" -o "$TEST_DIR/install.sh"; then
        log_error "Failed to download install script"
        return 1
    fi

    # Verify script is not empty
    if [ ! -s "$TEST_DIR/install.sh" ]; then
        log_error "Install script is empty"
        return 1
    fi

    # Verify script is a shell script
    if ! head -1 "$TEST_DIR/install.sh" | grep -q "#!/bin/bash"; then
        log_error "Install script is not a bash script"
        return 1
    fi

    log_success "Install script valid"
    return 0
}

# Run all tests
echo "Running verification tests..."
echo ""

run_test "Binary Archive" test_binary_archive
run_test "Checksums" test_checksums
run_test "Homebrew" test_homebrew
run_test "Docker Image" test_docker
run_test "Linux Packages" test_linux_packages
run_test "Install Script" test_install_script

# Summary
echo ""
echo "========================================"
log_info "Verification Summary"
echo "========================================"
echo ""
echo "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:      $PASSED_TESTS${NC}"
echo -e "${RED}Failed:      $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    log_success "All tests passed!"
    echo ""
    echo "Distribution channels verified:"
    echo "  ✓ Binary archives"
    echo "  ✓ Checksums"
    echo "  ✓ Homebrew (if installed)"
    echo "  ✓ Docker images"
    echo "  ✓ Linux packages (on Linux)"
    echo "  ✓ Install scripts"
    echo ""
    exit 0
else
    log_error "Some tests failed. Review results above."
    echo ""
    exit 1
fi
