#!/bin/bash

# Docker Run Wrapper for PullPoet
# This script automatically mounts current git repo and passes environment variables

set -e

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

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

# Default image name
IMAGE_NAME="erkineren/pullpoet"

# Check if we're in a git repository
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    print_error "Current directory is not a git repository!"
    echo "Please run this script from within a git repository."
    exit 1
fi

# Get current directory (git repo root)
GIT_ROOT=$(git rev-parse --show-toplevel)
print_info "Git repository root: $GIT_ROOT"

# Build environment variables to pass
ENV_VARS=""

# List of PullPoet environment variables to pass through
PULLPOET_ENVS=(
    "PULLPOET_PROVIDER"
    "PULLPOET_PROVIDER_BASE_URL"
    "PULLPOET_MODEL"
    "PULLPOET_API_KEY"
    "PULLPOET_CLICKUP_PAT"
)

for env_var in "${PULLPOET_ENVS[@]}"; do
    if [ ! -z "${!env_var}" ]; then
        ENV_VARS="$ENV_VARS -e $env_var=${!env_var}"
        print_info "Passing environment variable: $env_var"
    fi
done

# Check if any environment variables are set
if [ -z "$ENV_VARS" ]; then
    print_warning "No PullPoet environment variables found!"
    echo "You may need to set PULLPOET_PROVIDER, PULLPOET_MODEL, etc."
fi

# Parse arguments to check for help
if [[ "$*" == *"--help"* ]] || [[ "$*" == *"-h"* ]]; then
    print_info "Running PullPoet with --help"
    docker run --rm "$IMAGE_NAME" "$@"
    exit 0
fi

# Parse arguments to check for version
if [[ "$*" == *"--version"* ]] || [[ "$*" == *"-v"* ]]; then
    print_info "Running PullPoet with version flag"
    docker run --rm "$IMAGE_NAME" "$@"
    exit 0
fi

print_info "Running PullPoet with auto-detected git repository..."
print_success "Repository will be mounted at: /workspace"
print_success "Working directory will be: /workspace"

# Run Docker with:
# - Git repository mounted as volume
# - Working directory set to mounted repo
# - Environment variables passed through
# - All arguments forwarded
docker run --rm \
    -v "$GIT_ROOT:/workspace" \
    -w /workspace \
    $ENV_VARS \
    "$IMAGE_NAME" \
    "$@"
