#!/bin/bash

# PullPoet Universal Install Script
# Automatically detects OS/architecture and installs latest release from GitHub
# Usage: curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
REPO_OWNER="erkineren"
REPO_NAME="pullpoet"
REPO_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pullpoet"

# Script behavior
FORCE_INSTALL=false
UPDATE_MODE=false

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect OS and architecture
detect_platform() {
    local os
    local arch

    # Detect OS
    case "$(uname -s)" in
    Linux*) os="linux" ;;
    Darwin*) os="darwin" ;;
    CYGWIN* | MINGW* | MSYS*) os="windows" ;;
    *)
        print_error "Unsupported operating system: $(uname -s)"
        exit 1
        ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    i386 | i686) arch="386" ;;
    *)
        print_error "Unsupported architecture: $(uname -m)"
        exit 1
        ;;
    esac

    echo "${os}-${arch}"
}

# Function to get latest release version from GitHub
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"

    if command_exists curl; then
        curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//'
    elif command_exists wget; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//'
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
}

# Function to get currently installed version
get_installed_version() {
    if command_exists "$BINARY_NAME"; then
        "$BINARY_NAME" --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown"
    else
        echo "not_installed"
    fi
}

# Function to compare versions
version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"
}

# Function to download and install binary
install_binary() {
    local version="$1"
    local platform="$2"

    # Parse platform into OS and arch
    local os="${platform%-*}"
    local arch="${platform#*-}"

    # Convert to proper format for GitHub releases
    local os_formatted
    case "$os" in
    linux) os_formatted="Linux" ;;
    darwin) os_formatted="Darwin" ;;
    windows) os_formatted="Windows" ;;
    *)
        print_error "Unsupported OS: $os"
        exit 1
        ;;
    esac

    local arch_formatted
    case "$arch" in
    amd64) arch_formatted="x86_64" ;;
    arm64) arch_formatted="arm64" ;;
    386) arch_formatted="i386" ;;
    *)
        print_error "Unsupported architecture: $arch"
        exit 1
        ;;
    esac

    # Determine file extension
    local extension
    if [[ "$os" == "windows" ]]; then
        extension=".zip"
    else
        extension=".tar.gz"
    fi

    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${version}/pullpoet_${os_formatted}_${arch_formatted}${extension}"

    print_info "Downloading ${BINARY_NAME} v${version} for ${platform}..."
    print_info "Download URL: ${download_url}"

    # Create temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    local archive_file="${tmp_dir}/archive"

    # Download the release archive
    if command_exists curl; then
        if ! curl -L --fail --progress-bar "$download_url" -o "$archive_file"; then
            print_error "Failed to download release archive"
            rm -rf "$tmp_dir"
            exit 1
        fi
    elif command_exists wget; then
        if ! wget --progress=bar "$download_url" -O "$archive_file"; then
            print_error "Failed to download release archive"
            rm -rf "$tmp_dir"
            exit 1
        fi
    else
        print_error "Neither curl nor wget is available"
        rm -rf "$tmp_dir"
        exit 1
    fi

    # Extract the archive
    print_info "Extracting archive..."
    local extract_dir="${tmp_dir}/extract"
    mkdir -p "$extract_dir"

    if [[ "$download_url" == *.zip ]]; then
        if command_exists unzip; then
            unzip -q "$archive_file" -d "$extract_dir"
        else
            print_error "unzip command not found. Please install unzip."
            rm -rf "$tmp_dir"
            exit 1
        fi
    else
        if command_exists tar; then
            tar -xzf "$archive_file" -C "$extract_dir"
        else
            print_error "tar command not found. Please install tar."
            rm -rf "$tmp_dir"
            exit 1
        fi
    fi

    # Find the binary
    local binary_path
    binary_path=$(find "$extract_dir" -name "$BINARY_NAME" -type f | head -n 1)
    if [[ -z "$binary_path" ]]; then
        binary_path=$(find "$extract_dir" -name "${BINARY_NAME}.exe" -type f | head -n 1)
    fi

    if [[ -z "$binary_path" ]]; then
        print_error "Binary not found in the downloaded archive"
        rm -rf "$tmp_dir"
        exit 1
    fi

    # Install the binary
    print_info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."

    # Check if we need sudo
    if [[ ! -w "$INSTALL_DIR" ]]; then
        if command_exists sudo; then
            sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
            sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
        else
            print_error "No write permission to ${INSTALL_DIR} and sudo not available"
            print_info "Please run: cp $binary_path ${INSTALL_DIR}/${BINARY_NAME} as root"
            rm -rf "$tmp_dir"
            exit 1
        fi
    else
        cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Cleanup
    rm -rf "$tmp_dir"

    print_success "${BINARY_NAME} v${version} installed successfully!"
}

# Function to uninstall
uninstall() {
    print_info "Uninstalling ${BINARY_NAME}..."

    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"

    if [[ ! -f "$binary_path" ]]; then
        print_warning "${BINARY_NAME} is not installed in ${INSTALL_DIR}"
        exit 0
    fi

    # Check if we need sudo
    if [[ ! -w "$binary_path" ]]; then
        if command_exists sudo; then
            sudo rm "$binary_path"
        else
            print_error "No write permission and sudo not available"
            print_info "Please run: rm $binary_path as root"
            exit 1
        fi
    else
        rm "$binary_path"
    fi

    print_success "${BINARY_NAME} uninstalled successfully!"
}

# Function to show usage
show_usage() {
    echo "PullPoet Universal Install Script"
    echo ""
    echo "Usage:"
    echo "  $0                     Install latest version"
    echo "  $0 --update           Update to latest version"
    echo "  $0 --force            Force reinstall current version"
    echo "  $0 --uninstall        Uninstall pullpoet"
    echo "  $0 --help             Show this help"
    echo ""
    echo "Environment Variables:"
    echo "  INSTALL_DIR           Installation directory (default: /usr/local/bin)"
    echo ""
    echo "Examples:"
    echo "  # Install latest version"
    echo "  curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash"
    echo ""
    echo "  # Update to latest version"
    echo "  curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash -s -- --update"
    echo ""
    echo "  # Install to custom directory"
    echo "  INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash"
}

# Main installation function
main() {
    echo "🚀 PullPoet Universal Install Script"
    echo "======================================"
    echo ""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
        --force)
            FORCE_INSTALL=true
            shift
            ;;
        --update)
            UPDATE_MODE=true
            shift
            ;;
        --uninstall)
            uninstall
            exit 0
            ;;
        --help | -h)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        esac
    done

    # Use custom install directory if provided
    if [[ -n "$INSTALL_DIR_ENV" ]]; then
        INSTALL_DIR="$INSTALL_DIR_ENV"
    fi

    print_info "Target installation directory: $INSTALL_DIR"

    # Detect platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"

    # Get latest version
    print_info "Fetching latest release information..."
    local latest_version
    latest_version=$(get_latest_version)

    if [[ -z "$latest_version" ]]; then
        print_error "Failed to fetch latest version from GitHub"
        exit 1
    fi

    print_info "Latest version available: v$latest_version"

    # Check currently installed version
    local installed_version
    installed_version=$(get_installed_version)

    if [[ "$installed_version" != "not_installed" ]]; then
        print_info "Currently installed version: v$installed_version"

        if [[ "$UPDATE_MODE" == "true" ]]; then
            if [[ "$installed_version" == "$latest_version" ]]; then
                print_success "Already running the latest version (v$latest_version)"
                exit 0
            elif version_gt "$latest_version" "$installed_version"; then
                print_info "Updating from v$installed_version to v$latest_version"
            else
                print_warning "Installed version (v$installed_version) is newer than latest release (v$latest_version)"
                if [[ "$FORCE_INSTALL" != "true" ]]; then
                    read -p "Do you want to downgrade? (y/N): " -n 1 -r
                    echo
                    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                        print_info "Installation cancelled"
                        exit 0
                    fi
                fi
            fi
        elif [[ "$FORCE_INSTALL" != "true" ]]; then
            if [[ "$installed_version" == "$latest_version" ]]; then
                print_success "PullPoet v$latest_version is already installed"
                print_info "Use --force to reinstall or --update to check for updates"
                exit 0
            else
                print_warning "PullPoet is already installed (v$installed_version)"
                read -p "Do you want to install v$latest_version? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    print_info "Installation cancelled"
                    exit 0
                fi
            fi
        fi
    fi

    # Install the binary
    install_binary "$latest_version" "$platform"

    # Verify installation
    echo ""
    print_info "Verifying installation..."

    if command_exists "$BINARY_NAME"; then
        local final_version
        final_version=$("$BINARY_NAME" --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
        if [[ "$final_version" == "$latest_version" ]]; then
            print_success "Installation verified successfully!"
            echo ""
            echo "🎉 PullPoet v$latest_version is now installed!"
            echo ""
            echo "Quick start:"
            echo "  pullpoet --help                    # Show help"
            echo "  pullpoet --version                 # Show version"
            echo ""
            echo "Example usage:"
            echo "  export PULLPOET_PROVIDER=openai"
            echo "  export PULLPOET_MODEL=gpt-3.5-turbo"
            echo "  export PULLPOET_API_KEY=your-api-key"
            echo "  pullpoet --target main"
            echo ""
            echo "Learn more: $REPO_URL"
        else
            print_warning "Installation completed but version verification failed"
            print_info "Expected: v$latest_version, Got: v$final_version"
        fi
    else
        print_error "Installation completed but binary not found in PATH"
        print_info "You may need to add $INSTALL_DIR to your PATH"
    fi
}

# Check prerequisites
if ! command_exists curl && ! command_exists wget; then
    print_error "Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Run main function
main "$@"
