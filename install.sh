#!/bin/bash
# ABOUTME: Installation script for Clem CLI
# ABOUTME: Downloads latest release and installs to system or user bin directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="harper/clem"
BINARY_NAME="clem"
INSTALL_DIR=""
USE_SUDO=false

# Function to print colored output
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "darwin"
            ;;
        Linux*)
            echo "linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            error "Unsupported operating system: $(uname -s)"
            ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "x86_64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        *)
            error "Unsupported architecture: $(uname -m)"
            ;;
    esac
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release version..."

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version"
    fi

    info "Latest version: $VERSION"
    echo "$VERSION"
}

# Download binary
download_binary() {
    local os=$1
    local arch=$2
    local version=$3

    # Construct filename based on OS
    local os_name
    case "$os" in
        darwin) os_name="Darwin" ;;
        linux) os_name="Linux" ;;
        windows) os_name="Windows" ;;
    esac

    local archive_ext="tar.gz"
    if [ "$os" = "windows" ]; then
        archive_ext="zip"
    fi

    local filename="${BINARY_NAME}_${version#v}_${os_name}_${arch}.${archive_ext}"
    local url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    local checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    info "Downloading from: $url"

    local tmp_dir=$(mktemp -d)
    cd "$tmp_dir"

    # Download archive
    if command -v curl >/dev/null 2>&1; then
        curl -sL -o "$filename" "$url" || error "Failed to download binary"
        curl -sL -o "checksums.txt" "$checksum_url" || warn "Failed to download checksums"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "$filename" "$url" || error "Failed to download binary"
        wget -q -O "checksums.txt" "$checksum_url" || warn "Failed to download checksums"
    fi

    # Verify checksum if available
    if [ -f "checksums.txt" ]; then
        info "Verifying checksum..."
        if command -v sha256sum >/dev/null 2>&1; then
            grep "$filename" checksums.txt | sha256sum -c - || error "Checksum verification failed"
        elif command -v shasum >/dev/null 2>&1; then
            grep "$filename" checksums.txt | shasum -a 256 -c - || error "Checksum verification failed"
        else
            warn "sha256sum/shasum not found, skipping checksum verification"
        fi
    fi

    # Extract binary
    info "Extracting binary..."
    if [ "$archive_ext" = "zip" ]; then
        unzip -q "$filename" || error "Failed to extract archive"
    else
        tar -xzf "$filename" || error "Failed to extract archive"
    fi

    echo "$tmp_dir"
}

# Determine install directory
determine_install_dir() {
    # Try user local bin first
    if [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
        USE_SUDO=false
        return
    fi

    # Check if user has write permission to /usr/local/bin
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        USE_SUDO=false
        return
    fi

    # Fall back to system directory with sudo
    if [ -d "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        USE_SUDO=true
        warn "Will install to $INSTALL_DIR (requires sudo)"
        return
    fi

    # Create user local bin as last resort
    info "Creating $HOME/.local/bin directory"
    mkdir -p "$HOME/.local/bin"
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=false
}

# Install binary
install_binary() {
    local tmp_dir=$1

    determine_install_dir

    info "Installing to: $INSTALL_DIR"

    local binary_path="$tmp_dir/$BINARY_NAME"

    # Handle Windows executable extension
    if [ "$(detect_os)" = "windows" ]; then
        binary_path="${binary_path}.exe"
    fi

    if [ ! -f "$binary_path" ]; then
        error "Binary not found at $binary_path"
    fi

    # Make binary executable
    chmod +x "$binary_path"

    # Install binary
    if [ "$USE_SUDO" = true ]; then
        sudo mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME" || error "Failed to install binary"
    else
        mv "$binary_path" "$INSTALL_DIR/$BINARY_NAME" || error "Failed to install binary"
    fi

    success "Binary installed to $INSTALL_DIR/$BINARY_NAME"
}

# Check if directory is in PATH
check_path() {
    local dir=$1
    if [[ ":$PATH:" != *":$dir:"* ]]; then
        warn "$dir is not in your PATH"
        info "Add it to your PATH by adding this to your shell config:"
        echo ""
        echo "  export PATH=\"$dir:\$PATH\""
        echo ""
        return 1
    fi
    return 0
}

# Main installation flow
main() {
    echo ""
    info "Clem CLI Installer"
    echo ""

    # Detect system
    local os=$(detect_os)
    local arch=$(detect_arch)
    info "Detected OS: $os"
    info "Detected Architecture: $arch"

    # Get latest version
    local version=$(get_latest_version)

    # Download binary
    local tmp_dir=$(download_binary "$os" "$arch" "$version")

    # Install binary
    install_binary "$tmp_dir"

    # Cleanup
    rm -rf "$tmp_dir"

    # Verify installation
    echo ""
    info "Verifying installation..."
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        success "$BINARY_NAME installed successfully!"
        echo ""
        info "Version: $($BINARY_NAME --version 2>&1 || echo 'unknown')"
    else
        warn "$BINARY_NAME installed but not found in PATH"
        check_path "$INSTALL_DIR"
    fi

    # Print next steps
    echo ""
    success "Installation complete!"
    echo ""
    info "Next steps:"
    echo "  1. Configure your Anthropic API key:"
    echo "     ${BINARY_NAME} setup-token sk-ant-api03-..."
    echo ""
    echo "  2. Verify configuration:"
    echo "     ${BINARY_NAME} doctor"
    echo ""
    echo "  3. Start using Clem:"
    echo "     ${BINARY_NAME}"
    echo ""
    info "Documentation: https://github.com/${REPO}/blob/main/docs/USER_GUIDE.md"
    echo ""
}

# Run main function
main "$@"
