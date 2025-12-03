# ABOUTME: Windows PowerShell installation script for Hex CLI
# ABOUTME: Downloads and installs the latest release for Windows

#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "harper/hex"
$BinaryName = "hex.exe"
$InstallDir = "$env:ProgramFiles\Hex"

# Colors for output
function Write-Info {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Green
}

function Write-ErrorMsg {
    param([string]$Message)
    Write-Host "Error: $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "Warning: $Message" -ForegroundColor Yellow
}

# Detect architecture
function Get-Architecture {
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")

    switch ($arch) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        default {
            Write-ErrorMsg "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get latest release version
function Get-LatestVersion {
    Write-Info "Fetching latest release..."

    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $version = $response.tag_name

        if ([string]::IsNullOrEmpty($version)) {
            throw "Version is empty"
        }

        Write-Info "Latest version: $version"
        return $version
    }
    catch {
        Write-ErrorMsg "Failed to fetch latest release: $_"
        exit 1
    }
}

# Download and extract binary
function Download-Binary {
    param(
        [string]$Version,
        [string]$Arch
    )

    # Remove 'v' prefix from version
    $versionNumber = $Version -replace '^v', ''

    Write-Info "Downloading Hex $versionNumber for Windows $Arch..."

    # Construct download URL
    $archiveName = "hex_${versionNumber}_Windows_${Arch}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"

    Write-Info "URL: $downloadUrl"

    # Create temporary directory
    $tempDir = Join-Path $env:TEMP "hex-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download archive
        $archivePath = Join-Path $tempDir $archiveName
        Write-Info "Downloading to $archivePath..."

        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing

        # Extract archive
        Write-Info "Extracting archive..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

        # Verify binary exists
        $binaryPath = Join-Path $tempDir $BinaryName
        if (-not (Test-Path $binaryPath)) {
            throw "Binary not found in archive: $BinaryName"
        }

        return $binaryPath
    }
    catch {
        Write-ErrorMsg "Failed to download or extract: $_"
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }
}

# Install binary
function Install-Binary {
    param([string]$BinaryPath)

    Write-Info "Installing to $InstallDir..."

    try {
        # Create install directory if it doesn't exist
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        # Copy binary
        $destination = Join-Path $InstallDir $BinaryName
        Copy-Item -Path $BinaryPath -Destination $destination -Force

        Write-Info "Binary installed to $destination"
    }
    catch {
        Write-ErrorMsg "Failed to install binary: $_"
        exit 1
    }
}

# Add to PATH
function Add-ToPath {
    Write-Info "Adding to system PATH..."

    try {
        # Get current PATH
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")

        # Check if already in PATH
        if ($currentPath -like "*$InstallDir*") {
            Write-Info "Already in PATH"
            return
        }

        # Add to PATH
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")

        # Update current session PATH
        $env:Path = "$env:Path;$InstallDir"

        Write-Info "Added to PATH"
    }
    catch {
        Write-Warning "Failed to add to PATH: $_"
        Write-Warning "You may need to add $InstallDir to your PATH manually"
    }
}

# Verify installation
function Test-Installation {
    Write-Info "Verifying installation..."

    # Refresh PATH for current session
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine")

    try {
        # Run version check
        $versionOutput = & "$InstallDir\$BinaryName" --version 2>&1

        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✅ Hex installed successfully!" -ForegroundColor Green
            Write-Host "   Version: $versionOutput" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Next steps:" -ForegroundColor Cyan
            Write-Host "  1. Restart your terminal to load the updated PATH" -ForegroundColor Gray
            Write-Host ""
            Write-Host "  2. Set up your API key:" -ForegroundColor Gray
            Write-Host "     `$env:ANTHROPIC_API_KEY = 'your-key-here'" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "  3. Start using Hex:" -ForegroundColor Gray
            Write-Host "     hex              # Interactive mode" -ForegroundColor Yellow
            Write-Host "     hex --help       # See all options" -ForegroundColor Yellow
            Write-Host "     hex --print `"Hello!`"  # One-shot query" -ForegroundColor Yellow
            Write-Host ""
        }
        else {
            throw "Version check failed with exit code $LASTEXITCODE"
        }
    }
    catch {
        Write-ErrorMsg "Installation verification failed: $_"
        Write-Warning "Try running 'hex --version' after restarting your terminal"
        exit 1
    }
}

# Main installation flow
function Main {
    Write-Host ""
    Write-Host "Hex Installation Script for Windows" -ForegroundColor Cyan
    Write-Host "=====================================" -ForegroundColor Cyan
    Write-Host ""

    # Check if running as administrator
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    $isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

    if (-not $isAdmin) {
        Write-ErrorMsg "This script requires administrator privileges"
        Write-Info "Please run PowerShell as Administrator and try again"
        exit 1
    }

    $arch = Get-Architecture
    Write-Info "Detected architecture: $arch"

    $version = Get-LatestVersion
    $binaryPath = Download-Binary -Version $version -Arch $arch
    Install-Binary -BinaryPath $binaryPath
    Add-ToPath
    Test-Installation

    # Cleanup
    $tempDir = Split-Path $binaryPath -Parent
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}

# Run main function
try {
    Main
}
catch {
    Write-ErrorMsg "Installation failed: $_"
    exit 1
}
