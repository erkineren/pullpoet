# Release Process

This directory contains scripts for managing PullPoet releases.

## Release Script

The `release.sh` script automates the release process for PullPoet:

1. **Tagging**: Creates and pushes a git tag
2. **GitHub Release**: Waits for GitHub CI to automatically create a release
3. **Homebrew Update**: Updates the Homebrew formula with the new version

## Prerequisites

Before running the release script, ensure you have:

- `git` - for version control
- `gh` (GitHub CLI) - for GitHub operations
- `go` - for building and testing
- Access to the Homebrew tap repository

## Setup

1. **Install GitHub CLI and authenticate:**

   ```bash
   # Install GitHub CLI
   brew install gh

   # Authenticate
   gh auth login
   ```

2. **Clone the Homebrew tap repository:**

   ```bash
   git clone https://github.com/erkineren/homebrew-pullpoet.git /Users/erkineren/development/workspaces/personal/homebrew-pullpoet
   ```

3. **Make the script executable:**
   ```bash
   chmod +x scripts/release.sh
   ```

## Usage

### Basic Release

```bash
./scripts/release.sh
```

The script will:

1. Prompt for the new version number
2. Run tests to ensure everything works
3. Build the project locally
4. Create and push a git tag
5. Wait for GitHub CI to create the release
6. Update the Homebrew formula
7. Commit and push Homebrew changes

### Custom Homebrew Tap Path

If your Homebrew tap is in a different location, you can set the path:

```bash
HOMEBREW_TAP_PATH="/path/to/your/homebrew-tap" ./scripts/release.sh
```

## Release Process Flow

1. **Version Input**: Enter the new semantic version (e.g., 2.2.0)
2. **Validation**: Script validates the version format and checks prerequisites
3. **Testing**: Runs `go test ./...` to ensure code quality
4. **Building**: Builds the project locally to catch any build issues
5. **Tagging**: Creates and pushes a git tag (e.g., v2.2.0)
6. **CI Trigger**: GitHub CI automatically detects the tag and creates a release
7. **Wait for Release**: Script waits for the GitHub release to be available
8. **Homebrew Update**: Downloads the release tarball and calculates SHA256
9. **Formula Update**: Updates the Homebrew formula with new URL and SHA256
10. **Commit Changes**: Commits and pushes Homebrew changes

## Troubleshooting

### Homebrew Tap Not Found

If you get an error about the Homebrew tap directory not being found:

```bash
# Clone the Homebrew tap repository
git clone https://github.com/erkineren/homebrew-pullpoet.git /Users/erkineren/development/workspaces/personal/homebrew-pullpoet
```

### GitHub CLI Not Authenticated

If GitHub CLI is not authenticated:

```bash
gh auth login
```

### Tag Already Exists

If the tag already exists, the script will ask if you want to delete and recreate it. Choose 'y' to proceed or 'N' to cancel.

### Release Not Found

If the GitHub release is not found after waiting, check:

1. GitHub CI status for the tag
2. Ensure the CI workflow is configured correctly
3. Wait a bit longer and run the script again

## Manual Steps (if needed)

If the automated process fails, you can perform the steps manually:

1. **Create and push tag:**

   ```bash
   git tag v2.2.0
   git push origin v2.2.0
   ```

2. **Wait for GitHub CI to create release**

3. **Update Homebrew formula manually:**
   ```bash
   cd /Users/erkineren/development/workspaces/personal/homebrew-pullpoet
   # Edit pullpoet.rb with new URL and SHA256
   git add pullpoet.rb
   git commit -m "Update pullpoet to v2.2.0"
   git push origin main
   ```

## Configuration

The script uses the following configuration:

- **Repository**: `erkineren/pullpoet`
- **Homebrew Tap**: `erkineren/homebrew-pullpoet`
- **Default Homebrew Path**: `/Users/erkineren/development/workspaces/personal/homebrew-pullpoet`

You can override the Homebrew path using the `HOMEBREW_TAP_PATH` environment variable.
