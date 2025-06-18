#!/bin/bash

# PullPoet Release Script
# This script handles tagging and Homebrew updates
# GitHub CI automatically creates releases when tags are pushed

set -e # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_NAME="erkineren/pullpoet"
HOMEBREW_TAP="erkineren/homebrew-pullpoet"
HOMEBREW_TAP_PATH="${HOMEBREW_TAP_PATH:-/Users/erkineren/development/workspaces/personal/homebrew-pullpoet}"
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

    if ! command_exists gh; then
        missing_tools+=("gh (GitHub CLI)")
    fi

    if ! command_exists go; then
        missing_tools+=("go")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_status "Please install the missing tools and try again."
        exit 1
    fi

    # Check if GitHub CLI is authenticated
    if ! gh auth status >/dev/null 2>&1; then
        print_error "GitHub CLI is not authenticated. Please run 'gh auth login' first."
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

    print_status "Current version: $current_version"
    read -p "Enter new version (e.g., 2.2.0): " new_version

    if [[ -z "$new_version" ]]; then
        print_error "Version cannot be empty"
        exit 1
    fi

    # Remove 'v' prefix if present
    new_version="${new_version#v}"

    # Validate version format (semantic versioning)
    if [[ ! $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format. Please use semantic versioning (e.g., 2.2.0)"
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
}

# Function to create and push git tag
create_git_tag() {
    local version=$1
    local tag="v$version"

    print_status "Creating git tag: $tag"

    # Check if tag already exists
    if git tag -l | grep -q "^$tag$"; then
        print_warning "Tag $tag already exists"
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "$tag"
            git push origin ":refs/tags/$tag" 2>/dev/null || true
        else
            print_error "Tag already exists. Please use a different version or delete the existing tag."
            exit 1
        fi
    fi

    # Create and push tag
    git tag "$tag"
    git push origin "$tag"

    print_success "Git tag $tag created and pushed"
    print_status "GitHub CI will automatically create a release for this tag"
}

# Function to wait for GitHub release
wait_for_github_release() {
    local version=$1
    local tag="v$version"
    local max_attempts=30
    local attempt=1

    print_status "Waiting for GitHub release to be created by CI..."

    while [ $attempt -le $max_attempts ]; do
        if gh release view "$tag" >/dev/null 2>&1; then
            print_success "GitHub release found!"
            return 0
        fi

        print_status "Attempt $attempt/$max_attempts: Release not ready yet, waiting 10 seconds..."
        sleep 10
        ((attempt++))
    done

    print_warning "GitHub release not found after $max_attempts attempts"
    print_status "You may need to check the CI status manually"
    return 1
}

# Function to update Homebrew formula
update_homebrew() {
    local version=$1
    local formula_file="$HOMEBREW_TAP_PATH/pullpoet.rb"

    print_status "Updating Homebrew formula..."

    # Check if Homebrew tap directory exists
    if [[ ! -d "$HOMEBREW_TAP_PATH" ]]; then
        print_error "Homebrew tap directory not found: $HOMEBREW_TAP_PATH"
        print_status "Please clone the Homebrew tap repository first:"
        print_status "git clone https://github.com/$HOMEBREW_TAP.git $HOMEBREW_TAP_PATH"
        return 1
    fi

    # Get the release URL and SHA256
    local release_url="https://github.com/$REPO_NAME/archive/v$version.tar.gz"
    local temp_file="/tmp/pullpoet-v$version.tar.gz"

    # Download the release tarball
    print_status "Downloading release tarball..."
    if ! curl -L -o "$temp_file" "$release_url"; then
        print_error "Failed to download release tarball"
        print_status "Make sure the GitHub release exists and is accessible"
        return 1
    fi

    # Calculate SHA256
    local sha256_hash
    sha256_hash=$(shasum -a 256 "$temp_file" | cut -d' ' -f1)

    # Clean up temp file
    rm "$temp_file"

    # Update the formula
    sed -i.bak "s|url \".*\"|url \"$release_url\"|" "$formula_file"
    sed -i.bak "s|sha256 \".*\"|sha256 \"$sha256_hash\"|" "$formula_file"

    # Remove backup file
    rm "${formula_file}.bak"

    print_success "Homebrew formula updated"
    print_status "New URL: $release_url"
    print_status "New SHA256: $sha256_hash"
}

# Function to commit and push Homebrew changes
commit_homebrew_changes() {
    local version=$1

    print_status "Committing Homebrew changes..."

    # Change to Homebrew tap directory
    cd "$HOMEBREW_TAP_PATH"

    # Check if there are changes to commit
    if git diff --quiet pullpoet.rb; then
        print_warning "No changes to commit in Homebrew formula"
        return
    fi

    git add pullpoet.rb
    git commit -m "Update pullpoet to v$version"
    git push origin main

    print_success "Homebrew changes committed and pushed"
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
    echo "Homebrew Formula: Updated"
    echo
    echo "Next steps:"
    echo "1. Verify the GitHub release assets"
    echo "2. Test the Homebrew installation: brew install $HOMEBREW_TAP/pullpoet"
    echo "3. Update documentation if needed"
    echo "=========================================="
}

# Main release function
main() {
    echo "🚀 PullPoet Release Script"
    echo "=========================="
    echo "Homebrew Tap Path: $HOMEBREW_TAP_PATH"
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

    # Run tests
    run_tests

    # Build locally
    build_local

    # Create git tag (this triggers GitHub CI to create release)
    create_git_tag "$version"

    # Wait for GitHub release to be created
    if wait_for_github_release "$version"; then
        # Update Homebrew formula
        update_homebrew "$version"

        # Commit Homebrew changes
        commit_homebrew_changes "$version"
    else
        print_warning "Skipping Homebrew update due to release not being ready"
        print_status "You can run the script again later to update Homebrew"
    fi

    # Display summary
    display_summary "$version"

    print_success "Release process completed successfully! 🎉"
}

# Run main function
main "$@"
