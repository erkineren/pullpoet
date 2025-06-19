#!/bin/bash

# PullPoet Release Script
# This script handles tagging. GitHub CI automatically creates releases and updates Homebrew

set -e # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_NAME="erkineren/pullpoet"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Function to print colored output
print_status() {
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

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."

    local missing_tools=()

    if ! command_exists git; then
        missing_tools+=("git")
    fi

    if ! command_exists go; then
        missing_tools+=("go")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_status "Please install the missing tools and try again."
        exit 1
    fi

    print_success "All prerequisites are satisfied"
}

# Function to get current version
get_current_version() {
    local version

    # Try to get version from git tags first
    version=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "")

    # If no git tags found, default to 0.0.0
    if [[ -z "$version" ]]; then
        version="0.0.0"
    fi

    echo "$version"
}

# Function to prompt for version
prompt_for_version() {
    local current_version
    current_version=$(get_current_version)

    echo "[INFO] Current version: $current_version" >&2
    read -p "Enter new version (e.g., 2.2.0): " new_version

    if [[ -z "$new_version" ]]; then
        echo "[ERROR] Version cannot be empty" >&2
        exit 1
    fi

    # Remove 'v' prefix if present
    new_version="${new_version#v}"

    # Validate version format (semantic versioning)
    if [[ ! $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "[ERROR] Invalid version format. Please use semantic versioning (e.g., 2.2.0)" >&2
        exit 1
    fi

    echo "$new_version"
}

# Function to run tests
run_tests() {
    print_status "Running tests..."

    if ! go test ./...; then
        print_error "Tests failed"
        exit 1
    fi

    print_success "All tests passed"
}

# Function to build locally
build_local() {
    print_status "Building locally..."

    if ! go build -o pullpoet cmd/main.go; then
        print_error "Build failed"
        exit 1
    fi

    print_success "Local build successful"

    # Clean up the binary
    rm -f pullpoet
}

# Function to create and push git tag
create_git_tag() {
    local version=$1
    local tag="v$version"

    echo "[INFO] Creating git tag: $tag" >&2

    # Check if tag already exists
    if git tag -l | grep -q "^$tag$"; then
        echo "[WARNING] Tag $tag already exists" >&2
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "$tag"
            git push origin ":refs/tags/$tag" 2>/dev/null || true
        else
            echo "[ERROR] Tag already exists. Please use a different version or delete the existing tag." >&2
            exit 1
        fi
    fi

    # Create and push tag
    git tag "$tag"
    git push origin "$tag"

    echo "[SUCCESS] Git tag $tag created and pushed" >&2
    echo "[INFO] GitHub CI will automatically create a release and update Homebrew" >&2
}

# Function to display release summary
display_summary() {
    local version=$1

    echo
    echo "=========================================="
    echo "🎉 Release Summary"
    echo "=========================================="
    echo "Version: $version"
    echo "Tag: v$version"
    echo "GitHub Release: https://github.com/$REPO_NAME/releases/tag/v$version"
    echo
    echo "Automated processes:"
    echo "✅ GitHub Release will be created automatically"
    echo "✅ Homebrew formula will be updated automatically"
    echo
    echo "Next steps:"
    echo "1. Check GitHub Actions for release and homebrew update status"
    echo "2. Verify the GitHub release assets"
    echo "3. Test the Homebrew installation: brew install erkineren/pullpoet/pullpoet"
    echo "=========================================="
}

# Main release function
main() {
    echo "🚀 PullPoet Release Script"
    echo "=========================="
    echo

    # Check prerequisites
    check_prerequisites

    # Change to project root
    cd "$PROJECT_ROOT"

    # Ensure we're on main branch
    if [[ $(git branch --show-current) != "main" ]]; then
        print_error "Please switch to the main branch before releasing"
        exit 1
    fi

    # Ensure working directory is clean
    if [[ -n $(git status --porcelain) ]]; then
        print_error "Working directory is not clean. Please commit or stash your changes."
        exit 1
    fi

    # Get version
    local version
    version=$(prompt_for_version)

    # Clear any extra output
    echo

    # Run tests
    run_tests

    # Build locally to verify
    build_local

    # Create git tag (this triggers GitHub CI to create release and update Homebrew)
    create_git_tag "$version"

    # Display summary
    display_summary "$version"

    print_success "Release process initiated successfully! 🎉"
    print_status "Check GitHub Actions for automated release and Homebrew update progress."
}

# Run main function
main "$@"
