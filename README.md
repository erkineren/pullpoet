# PullPoet

[![CI](https://github.com/erkineren/pullpoet/workflows/CI/badge.svg)](https://github.com/erkineren/pullpoet/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/erkineren/pullpoet)](https://goreportcard.com/report/github.com/erkineren/pullpoet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/erkineren/pullpoet.svg)](https://pkg.go.dev/github.com/erkineren/pullpoet)

A Go CLI tool that generates AI-powered pull request titles and descriptions by analyzing git diffs.

## Features

- **🍺 Homebrew Support**: Easy installation via Homebrew on macOS and Linux
- **🐳 Docker Support**: Run as a container without local installation
- **🤖 Multi-Provider Support**: Works with OpenAI, Ollama, Google Gemini, and OpenWebUI
- **🔍 Auto-Detection**: Automatically detects git repository, current branch, and default branch
- **🌍 Environment Variables**: Support for configuration via environment variables
- **🔑 SSH & HTTPS Support**: Works with both public and private repositories
- **📦 Git Integration**: Automatically clones repositories and generates diffs between branches
- **🧠 Smart Analysis**: Combines git diffs with optional manual descriptions for context
- **✨ Professional Output**: Generates beautiful PR descriptions with emojis and structured markdown
- **⚡ Performance Optimized**: Shallow cloning and fast mode for large repositories
- **📊 Progress Tracking**: Real-time logging of each operation step
- **🎨 Visual Design**: Modern PR descriptions with emojis, checkboxes, and proper formatting
- **📋 Structured Content**: Problem statements, technical changes, acceptance criteria, and testing notes
- **💾 File Output**: Save generated PR descriptions to markdown files
- **🔧 Flexible Options**: Configurable AI providers, models, and output formats

## Installation

### Universal Install Script (Recommended) 🚀

**One-liner installation for all platforms:**

```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash

# Update to latest version
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash -s -- --update

# Install to custom directory
INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash

# Uninstall
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.sh | bash -s -- --uninstall
```

**Features:**

- ✅ **Auto-detects OS/Architecture** (Linux, macOS, Windows)
- ✅ **Downloads latest release** from GitHub automatically
- ✅ **Update functionality** with `--update` flag
- ✅ **Custom installation directory** support
- ✅ **Uninstall option** with `--uninstall` flag
- ✅ **Version verification** after installation
- ✅ **Progress indicators** and colored output

### Windows Native Installation 🪟

**PowerShell (Recommended):**

```powershell
# Install latest version
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -UseBasicParsing).Content

# Update to latest version (use this method for parameters)
$script = Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -UseBasicParsing
Invoke-Expression $script.Content -Update

# Install to custom directory
$script = Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -UseBasicParsing
Invoke-Expression $script.Content -InstallDir "C:\Tools\pullpoet"

# Uninstall
$script = Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -UseBasicParsing
Invoke-Expression $script.Content -Uninstall

# Alternative: Download and run locally (recommended for parameters)
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.ps1" -OutFile "install.ps1"
.\install.ps1 -Uninstall
```

**Batch File (Easiest):**

```cmd
# Download and run the batch installer
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.bat -o install.bat
install.bat

# Or run directly
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/install.bat | cmd

# With parameters (update, uninstall, etc.)
install.bat -Update
install.bat -Uninstall
install.bat -InstallDir "C:\Tools\pullpoet"
```

**Windows Features:**

- ✅ **Auto-detects Windows architecture** (x86_64, i386)
- ✅ **Automatic PATH configuration** for user environment
- ✅ **Current session PATH refresh** - no terminal restart needed
- ✅ **Installation to user directory** (no admin required)
- ✅ **PowerShell 5.0+ compatibility** check
- ✅ **Progress indicators** and colored output
- ✅ **Update and uninstall** functionality

**Prerequisites:**

- PowerShell 5.0 or higher (Windows 10+ includes this by default)
- Internet connection for downloading

### Homebrew (macOS/Linux) 🍺

```bash
# Install directly from tap
brew install erkineren/pullpoet/pullpoet

# Or add the tap first and then install
brew tap erkineren/pullpoet
brew install pullpoet
```

### Docker 🐳

```bash
# Run with Docker Hub image
docker run --rm erkineren/pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key
```

### From Source

```bash
git clone https://github.com/erkineren/pullpoet.git
cd pullpoet
go build -o pullpoet cmd/main.go
```

### Using Go Install

```bash
go install ./cmd
```

## Usage

### Auto-Detection Features 🔍

PullPoet can automatically detect git repository information when run from within a git repository:

```bash
# Minimal usage - auto-detects repo, source branch, and target branch
pullpoet --provider openai --model gpt-3.5-turbo --api-key your-key

# Auto-detects repo and source, specify target
pullpoet --target main --provider openai --model gpt-3.5-turbo --api-key your-key

# Manual override of auto-detected values
pullpoet \
  --repo https://github.com/custom/repo.git \
  --source custom-branch \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-key
```

**Auto-Detection Features:**

- **Repository URL**: Automatically uses `git remote get-url origin`
- **Source Branch**: Uses current branch (`git rev-parse --abbrev-ref HEAD`)
- **Target Branch**: Uses default branch (typically `main` or `master`)
- **SSH Support**: Supports both SSH (`git@github.com:user/repo.git`) and HTTPS URLs
- **Private Repositories**: Works with private repos using SSH keys or git credentials

### Environment Variables Support 🌍

Configure PullPoet using environment variables to avoid repeating common parameters:

**Linux/macOS:**

```bash
# Set environment variables
export PULLPOET_PROVIDER=openai
export PULLPOET_MODEL=gpt-3.5-turbo
export PULLPOET_API_KEY=your-openai-api-key
export PULLPOET_PROVIDER_BASE_URL=https://api.openai.com  # optional
export PULLPOET_CLICKUP_PAT=pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ  # optional

# Now run with minimal flags
pullpoet --target main

# Environment variables + CLI flags (CLI flags take precedence)
pullpoet --target main --model gpt-4  # Uses gpt-4 instead of env var
```

**Windows (PowerShell):**

```powershell
# Set environment variables
$env:PULLPOET_PROVIDER="openai"
$env:PULLPOET_MODEL="gpt-3.5-turbo"
$env:PULLPOET_API_KEY="your-openai-api-key"
$env:PULLPOET_PROVIDER_BASE_URL="https://api.openai.com"  # optional
$env:PULLPOET_CLICKUP_PAT="pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ"  # optional

# Now run with minimal flags
pullpoet --target main
```

**Windows (Command Prompt):**

```cmd
# Set environment variables
set PULLPOET_PROVIDER=openai
set PULLPOET_MODEL=gpt-3.5-turbo
set PULLPOET_API_KEY=your-openai-api-key
set PULLPOET_PROVIDER_BASE_URL=https://api.openai.com
set PULLPOET_CLICKUP_PAT=pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ

# Now run with minimal flags
pullpoet --target main
```

**Supported Environment Variables:**

- `PULLPOET_PROVIDER` - AI provider (`openai`, `ollama`, `gemini`, `openwebui`)
- `PULLPOET_MODEL` - AI model name
- `PULLPOET_API_KEY` - API key for OpenAI/Gemini/OpenWebUI
- `PULLPOET_PROVIDER_BASE_URL` - Base URL for AI provider
- `PULLPOET_CLICKUP_PAT` - ClickUp Personal Access Token

**Priority Order:** CLI flags > Environment variables > Default values

### Docker Usage 🐳

You can run PullPoet directly using Docker without installing it locally.

#### **🎯 Smart Docker Usage (Recommended)**

For the best experience with auto-detection, use the Docker wrapper script:

```bash
# Download and use the Docker wrapper script
curl -sSL https://raw.githubusercontent.com/erkineren/pullpoet/main/scripts/docker-run.sh > pullpoet-docker
chmod +x pullpoet-docker

# Set your environment variables
export PULLPOET_PROVIDER=gemini
export PULLPOET_MODEL=gemini-2.5-flash-preview-05-20
export PULLPOET_API_KEY=your-gemini-api-key

# Run from within your git repository - auto-detects everything!
cd /path/to/your/git/repo
./pullpoet-docker
```

**What the wrapper script does:**

- ✅ **Auto-mounts current git repository** → Git info detection works
- ✅ **Passes environment variables** → No need to repeat API keys
- ✅ **Sets working directory correctly** → Auto-detection works perfectly
- ✅ **Validates git repository** → Prevents common errors

#### **⚡ Manual Docker Usage**

If you prefer manual control or can't use the wrapper script:

```bash
# Option A: Mount current git repository for auto-detection
docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  -e PULLPOET_PROVIDER=gemini \
  -e PULLPOET_MODEL=gemini-2.5-flash-preview-05-20 \
  -e PULLPOET_API_KEY=your-gemini-api-key \
  erkineren/pullpoet

# Option B: Manual parameters (no auto-detection)
docker run --rm erkineren/pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider gemini \
  --model gemini-2.5-flash-preview-05-20 \
  --api-key your-gemini-api-key

# Option C: Save output to file
docker run --rm \
  -v "$(pwd):/workspace" \
  -w /workspace \
  -e PULLPOET_PROVIDER=gemini \
  -e PULLPOET_MODEL=gemini-2.5-flash-preview-05-20 \
  -e PULLPOET_API_KEY=your-gemini-api-key \
  erkineren/pullpoet \
  --output pr-description.md
```

#### **🐛 Docker Troubleshooting**

**Problem:** Auto-detection doesn't work in Docker  
**Solution:** Mount your git repository as volume:

```bash
# ❌ Wrong - no git repository access
docker run --rm erkineren/pullpoet

# ✅ Correct - git repository mounted
docker run --rm -v "$(pwd):/workspace" -w /workspace erkineren/pullpoet
```

**Problem:** Environment variables not passed  
**Solution:** Use `-e` flag or wrapper script:

```bash
# ❌ Wrong - environment variables not accessible
docker run --rm erkineren/pullpoet

# ✅ Correct - environment variables passed
docker run --rm \
  -e PULLPOET_PROVIDER=gemini \
  -e PULLPOET_API_KEY=your-key \
  erkineren/pullpoet
```

### Basic Usage

**Traditional Usage (Manual Parameters):**

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key
```

**Smart Usage (Auto-Detection + Environment Variables):**

```bash
# Set environment variables once
export PULLPOET_PROVIDER=openai
export PULLPOET_MODEL=gpt-3.5-turbo
export PULLPOET_API_KEY=your-openai-api-key

# Run from within your git repository
cd /path/to/your/repo
pullpoet --target main  # Auto-detects repo URL and current branch
```

**Minimal Usage (Maximum Auto-Detection):**

```bash
# If target branch is default branch (main/master)
export PULLPOET_PROVIDER=openai
export PULLPOET_MODEL=gpt-3.5-turbo
export PULLPOET_API_KEY=your-openai-api-key

cd /path/to/your/repo
pullpoet  # Auto-detects everything!
```

### Quick Start with Gemini (Recommended)

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider gemini \
  --model gemini-2.5-flash-preview-05-20 \
  --api-key your-gemini-api-key
```

### With Optional Description

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --description "User story: As a user, I want to login with email and password so I can access my account securely" \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key
```

### Using Ollama

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider ollama \
  --model llama2 \
  --provider-base-url https://user:password@ollama.home.example.com \
  --description "JIRA-123: Implement OAuth2 login flow with Google and GitHub providers"
```

### Using OpenWebUI

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openwebui \
  --model llama3.1 \
  --provider-base-url http://localhost:3000 \
  --api-key your-openwebui-api-key \
  --description "Implement secure user authentication with JWT tokens"
```

### Using Google Gemini

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider gemini \
  --model gemini-2.5-flash-preview-05-20 \
  --api-key your-gemini-api-key \
  --description "Implement secure user authentication with JWT tokens"
```

### Fast Mode (Recommended for Large Repositories)

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key \
  --fast
```

### ClickUp Integration

Automatically fetch task descriptions from ClickUp using Personal Access Token (PAT):

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/new-feature \
  --target main \
  --provider gemini \
  --model gemini-2.5-flash-preview-05-20 \
  --api-key your-gemini-api-key \
  --clickup-pat pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ \
  --clickup-task-id 86c2dbq35
```

When both ClickUp PAT and task ID are provided, the tool will automatically fetch the task description from ClickUp, including task name, status, creator, and full description. This eliminates the need to manually copy and paste task details.

### Save to File

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key \
  --output pr-description.md
```

### Using Custom System Prompt

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key \
  --system-prompt /path/to/custom-prompt.md
```

## Configuration Options

| Flag                  | Description                                                                          | Required                          | Environment Variable         | Example                                                     |
| --------------------- | ------------------------------------------------------------------------------------ | --------------------------------- | ---------------------------- | ----------------------------------------------------------- |
| `--repo`              | Git repository URL (auto-detected if running in git repo)                            | No\*                              | N/A                          | `https://github.com/example/repo.git`                       |
| `--source`            | Source branch name (auto-detected as current branch)                                 | No\*                              | N/A                          | `feature/new-feature`                                       |
| `--target`            | Target branch name (auto-detected as default branch)                                 | No\*                              | N/A                          | `main`                                                      |
| `--description`       | Optional issue/task description from ClickUp, Jira, etc.                             | No                                | N/A                          | `"JIRA-123: Add user authentication feature"`               |
| `--provider`          | AI provider (`openai`, `ollama`, `gemini`, or `openwebui`)                           | Yes\*\*                           | `PULLPOET_PROVIDER`          | `openai`                                                    |
| `--model`             | AI model to use                                                                      | Yes\*\*                           | `PULLPOET_MODEL`             | `gpt-3.5-turbo`, `llama2`, `gemini-2.5-flash-preview-05-20` |
| `--api-key`           | OpenAI, Gemini, or OpenWebUI API key                                                 | Yes (for OpenAI/Gemini/OpenWebUI) | `PULLPOET_API_KEY`           | `sk-...` or `AIza...`                                       |
| `--provider-base-url` | Base URL for AI provider (required for Ollama/OpenWebUI, optional for OpenAI/Gemini) | Yes (for Ollama/OpenWebUI)        | `PULLPOET_PROVIDER_BASE_URL` | `https://user:pass@host:port` or `http://localhost:3000`    |
| `--system-prompt`     | Custom system prompt file path to override default                                   | No                                | N/A                          | `/path/to/custom-prompt.md`                                 |
| `--clickup-pat`       | ClickUp Personal Access Token                                                        | No                                | `PULLPOET_CLICKUP_PAT`       | `pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ`                   |
| `--clickup-task-id`   | ClickUp Task ID to fetch description from                                            | No                                | N/A\*\*\*                    | `86c2dbq35`                                                 |
| `--fast`              | Use fast native git commands                                                         | No                                | N/A                          | `--fast`                                                    |
| `--output`            | Output file path                                                                     | No                                | N/A                          | `output.md`                                                 |

**Notes:**

- `*` Auto-detected when running from within a git repository
- `**` Required unless set via environment variable
- `***` Task ID is not supported via environment variable as it's typically unique per PR

## Custom System Prompts

PullPoet allows you to customize the system prompt used to instruct the AI model. By default, it uses an embedded prompt optimized for generating professional PR descriptions.

### Default Behavior

When no `--system-prompt` is provided, PullPoet uses its built-in system prompt that:

- Instructs the AI to generate structured PR descriptions
- Requests JSON format with 'title' and 'body' fields
- Provides guidelines for professional formatting
- Ensures consistent output across different AI providers

### Custom System Prompt

You can override the default system prompt by providing a custom markdown file:

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key \
  --system-prompt /path/to/custom-prompt.md
```

### Custom Prompt File Format

Your custom system prompt file should contain the instructions you want to send to the AI model. For example:

```markdown
# Custom System Prompt

You are an expert software developer and technical writer. Your task is to analyze git changes and generate pull request descriptions.

## Requirements:

- Generate a concise title (max 80 characters)
- Create a detailed description in markdown format
- Focus on technical changes and business impact
- Include testing considerations
- Use professional language and emojis for visual appeal

## Output Format:

Respond with a JSON object containing:

- "title": A concise PR title
- "body": A detailed markdown description

## Guidelines:

- Be specific about what changed and why
- Highlight security implications if any
- Mention performance impacts
- Include relevant file paths
```

### Use Cases for Custom Prompts

- **Team-specific formatting**: Match your team's PR template style
- **Domain-specific requirements**: Add industry-specific guidelines
- **Compliance requirements**: Include security or regulatory considerations
- **Language preferences**: Customize tone and terminology
- **Integration requirements**: Add specific formatting for external tools

## AI Providers

### OpenAI

- Uses the ChatGPT API
- Requires an OpenAI API key
- Set the `--api-key` flag with your OpenAI API key
- Supported models: `gpt-3.5-turbo`, `gpt-4`, `gpt-4-turbo-preview`, etc.

### Ollama

- Uses a self-hosted Ollama instance
- Supports basic authentication via URL credentials
- Set the `--provider-base-url` flag with your Ollama endpoint
- URL format: `https://username:password@your-ollama-host.com`
- Supported models: `llama2`, `codellama`, `mistral`, `neural-chat`, etc.

### OpenWebUI

- Uses OpenWebUI as a unified LLM provider interface
- Supports OpenAI-compatible API endpoints
- Set the `--provider-base-url` flag with your OpenWebUI endpoint (default: `http://localhost:3000`)
- Set the `--api-key` flag with your OpenWebUI API key (optional, can be obtained from Settings > Account)
- Compatible with any model available in your OpenWebUI instance (Ollama models, OpenAI models, etc.)
- **Benefits**: Unified interface for multiple LLM providers, local deployment, cost control

### Google Gemini

- Uses Google's Gemini AI models via the official API
- Requires a Google AI Studio API key ([Get your API key here](https://makersuite.google.com/app/apikey))
- Set the `--api-key` flag with your Gemini API key
- **Advanced Features**: Structured JSON output with schema validation for consistent responses
- **Models Supported** (for text generation):
  - `gemini-2.5-flash-preview-05-20` ⭐ **Recommended** - newest, adaptive thinking, cost efficient
  - `gemini-2.5-pro-preview-06-05` - enhanced reasoning, advanced coding, complex diffs
  - `gemini-2.0-flash` - next generation features, speed, thinking
  - `gemini-2.0-flash-lite` - cost efficiency, low latency
  - `gemini-1.5-flash` - fast and versatile, stable release
  - `gemini-1.5-flash-8b` - high volume, lower intelligence tasks
  - `gemini-1.5-pro` - complex reasoning tasks
- **Model Selection Guide**:
  - **For most users**: `gemini-2.5-flash-preview-05-20` (best balance of quality, speed, and cost)
  - **For complex repositories**: `gemini-2.5-pro-preview-06-05` (advanced reasoning)
  - **For high volume usage**: `gemini-2.0-flash-lite` or `gemini-1.5-flash-8b` (cost optimization)
- **Benefits**:
  - Guaranteed JSON format responses
  - Built-in response validation
  - Enhanced reliability over prompt-based approaches
  - Free tier available with generous quotas

## ClickUp Integration

PullPoet supports automatic task description fetching from ClickUp using the ClickUp API v2.

### Setup

1. **Get ClickUp Personal Access Token (PAT)**:

   - Log in to your ClickUp account
   - Go to Settings → Apps
   - Click "Generate" under API Token
   - Copy the generated token (starts with `pk_`)

2. **Find Task ID**:
   - Open any ClickUp task
   - The task ID is in the URL: `https://app.clickup.com/t/86c2dbq35`
   - Or copy it from the task's share options

### Usage

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/new-feature \
  --target main \
  --provider gemini \
  --model gemini-2.5-flash-preview-05-20 \
  --api-key your-gemini-api-key \
  --clickup-pat pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ \
  --clickup-task-id 86c2dbq35
```

### What Gets Fetched

When you provide both ClickUp PAT and task ID, PullPoet automatically fetches:

- **Task Name**: Used for context and PR title generation
- **Task Description**: Full task description or text content
- **Task Status**: Current status (e.g., "in progress", "back log")
- **Creator**: Task creator information
- **Task URL**: Direct link to the ClickUp task

The fetched information is formatted and used as the description input for AI analysis, providing rich context for generating relevant PR descriptions.

### Benefits

- **No Manual Copy-Paste**: Eliminates the need to manually copy task descriptions
- **Always Up-to-Date**: Fetches the latest task information
- **Rich Context**: Includes task metadata for better AI analysis
- **Standardized Format**: Consistent task information formatting

### Error Handling

- Invalid PAT: Clear error message with authentication failure
- Task Not Found: Helpful error when task ID doesn't exist
- Network Issues: Retry logic and timeout handling
- Partial Data: Graceful handling when some task fields are empty

## Performance Optimization

PullPoet includes several performance optimizations to handle large repositories efficiently:

### Default Mode (go-git library)

- **Shallow cloning**: Only fetches recent commits (depth: 50)
- **No checkout**: Skips working directory checkout for faster operation
- **Branch-specific fetch**: Only downloads required branches

### Fast Mode (`--fast` flag)

- **Native git commands**: Uses system git for maximum performance
- **Minimal data transfer**: Fetches only necessary commits
- **Optimized for large repositories**: Recommended for repositories >100MB

### Performance Comparison

- **Large repositories**: Use `--fast` for 3-5x faster cloning
- **Small repositories**: Default mode is sufficient
- **Network limited**: Both modes benefit from shallow cloning

## Output Format

The tool generates professional, visually appealing PR descriptions with emojis and structured content:

```
🎉 Generated PR Description
════════════════════════════════════════════════════════════

📋 **Title:**
🚀 Add user authentication feature

────────────────────────────────────────────────────────────

📝 **Description:**
# 🚀 Add user authentication feature

## 📋 Problem Statement / Overview
Users need secure authentication to access protected resources in the application.

## 🎯 Solution Overview
Implemented comprehensive authentication system with JWT tokens, secure session management, and role-based access control.

## 🔧 Technical Changes

### **Authentication Components**
- **AuthController**: Added login/logout endpoints with validation
- **JWTMiddleware**: Implemented token verification and refresh logic
- **UserService**: Created user session management with security features

### **Security Enhancements**
- **Password Hashing**: Added bcrypt hashing for secure password storage
- **Token Management**: JWT tokens with configurable expiration times

## ✅ Key Features / Acceptance Criteria

- [x] **Secure Login**: Users can authenticate with email/password
- [x] **Token Management**: JWT tokens for stateless authentication
- [x] **Session Security**: Proper logout and token invalidation
- [x] **Role-based Access**: Different permissions for user roles

## 🧪 Testing Considerations

- **Unit Tests**: Authentication logic and token validation
- **Integration Tests**: Login/logout endpoint functionality
- **Security Tests**: Password hashing and token security

## 📋 Files Changed
- `app/Http/Controllers/AuthController.php`
- `app/Http/Middleware/JWTMiddleware.php`
- `app/Services/UserService.php`

════════════════════════════════════════════════════════════
✅ PR description generated successfully!
💡 You can now copy this content to your pull request.
```

## Project Structure

```
/cmd
  main.go          # CLI entry point
/config
  config.go        # Configuration management
/internal
  /git
    diff.go        # Git operations (clone, diff)
    fast_diff.go   # Fast native git operations
  /ai
    client.go      # AI client interface
    openai.go      # OpenAI implementation
    ollama.go      # Ollama implementation
    openwebui.go   # OpenWebUI implementation
    gemini.go      # Google Gemini implementation
  /pr
    generate.go    # PR generation logic
```

## Development

### Prerequisites

- Go 1.21 or later
- Git

### Building

```bash
go mod tidy
go build -o pullpoet cmd/main.go
```

### Testing

```bash
go test ./...
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on how to get started.

### Quick Start for Contributors

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Recent Updates

- **🔍 Auto-Detection Support**: Automatically detects git repository URL, current branch, and default branch when running from within a git repository
- **🌍 Environment Variables**: Full support for configuration via environment variables (`PULLPOET_PROVIDER`, `PULLPOET_MODEL`, etc.)
- **🔑 SSH & Private Repository Support**: Enhanced support for private repositories using SSH authentication
- **⚡ Simplified Usage**: Minimal command line usage with smart defaults and auto-detection
- **🆕 OpenWebUI Support**: Added full support for OpenWebUI as a unified LLM provider interface
- **🔧 Unified Provider URLs**: Replaced provider-specific URL flags with a single `--provider-base-url` parameter
- **⚡ Default URLs**: Added default URLs for all providers (OpenAI: `https://api.openai.com`, Ollama: `http://localhost:11434`, OpenWebUI: `http://localhost:3000`)
- **🎯 Flexible Configuration**: Provider base URLs are required for Ollama/OpenWebUI, optional for OpenAI/Gemini to override defaults
- **🆕 Google Gemini Support**: Added full support for Google's Gemini AI models
- **📋 Structured Output**: Implemented schema-validated JSON responses for enhanced reliability
- **⚡ Performance**: Gemini 2.0 Flash provides fastest response times
- **🔧 Multi-Provider**: Now supports OpenAI, Ollama, and Google Gemini

## Future Features

- Support for GitLab and GitHub API integration
- Automatic PR creation
- Custom AI model selection for Ollama
- Configuration file support
- Template customization
- Additional AI providers (Anthropic Claude, Azure OpenAI)
- Enhanced OpenWebUI features (RAG, file uploads, knowledge collections)

## AI Analysis Features

PullPoet provides comprehensive analysis of your changes:

### Git Information Analysis

- **Diff Analysis**: Complete diff between source and target branches
- **Commit History**: Individual commit messages, authors, and timestamps
- **File Changes**: Detailed breakdown of modified, added, and deleted files

### Issue Context Integration

- **Task Management Integration**: Include original issue descriptions from ClickUp, Jira, Asana, etc.
- **Requirements Mapping**: AI understands the original requirements and maps them to code changes
- **Business Context**: Maintains connection between business needs and technical implementation

### Professional Output

- **Structured Format**: Consistent markdown format with emojis and clear sections
- **Repository Links**: Real GitHub/GitLab links (no placeholder URLs)
- **Technical Documentation**: Detailed technical changes and testing considerations
