# PullPoet

[![CI](https://github.com/yourusername/pullpoet/workflows/CI/badge.svg)](https://github.com/yourusername/pullpoet/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/pullpoet)](https://goreportcard.com/report/github.com/yourusername/pullpoet)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/pullpoet.svg)](https://pkg.go.dev/github.com/yourusername/pullpoet)

A Go CLI tool that generates AI-powered pull request titles and descriptions by analyzing git diffs.

## Features

- **🤖 Multi-Provider Support**: Works with OpenAI, Ollama, and Google Gemini
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

### From Source

```bash
git clone https://github.com/yourusername/pullpoet.git
cd pullpoet
go build -o pullpoet cmd/main.go
```

### Using Go Install

```bash
go install ./cmd
```

## Usage

### Basic Usage

```bash
pullpoet \
  --repo https://github.com/example/repo.git \
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key
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
  --ollama-url https://user:password@ollama.home.example.com \
  --description "JIRA-123: Implement OAuth2 login flow with Google and GitHub providers"
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
  --source feature/login \
  --target main \
  --provider openai \
  --model gpt-3.5-turbo \
  --api-key your-openai-api-key \
  --clickup-pat your-clickup-pat-token \
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

## Configuration Options

| Flag                | Description                                              | Required                | Example                                                     |
| ------------------- | -------------------------------------------------------- | ----------------------- | ----------------------------------------------------------- |
| `--repo`            | Git repository URL                                       | Yes                     | `https://github.com/owner/repo.git`                         |
| `--source`          | Source branch name                                       | Yes                     | `feature/new-feature`                                       |
| `--target`          | Target branch name                                       | Yes                     | `main`                                                      |
| `--description`     | Optional issue/task description from ClickUp, Jira, etc. | No                      | `"JIRA-123: Add user authentication feature"`               |
| `--provider`        | AI provider (`openai`, `ollama`, or `gemini`)            | Yes                     | `openai`                                                    |
| `--model`           | AI model to use                                          | Yes                     | `gpt-3.5-turbo`, `llama2`, `gemini-2.5-flash-preview-05-20` |
| `--api-key`         | OpenAI or Gemini API key                                 | Yes (for OpenAI/Gemini) | `sk-...` or `AIza...`                                       |
| `--ollama-url`      | Ollama endpoint with credentials                         | Yes (for Ollama)        | `https://user:pass@host:port`                               |
| `--clickup-pat`     | ClickUp Personal Access Token                            | No                      | `pk_123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ`                   |
| `--clickup-task-id` | ClickUp Task ID to fetch description from                | No                      | `86c2dbq35`                                                 |
| `--fast`            | Use fast native git commands                             | No                      | `--fast`                                                    |
| `--output`          | Output file path                                         | No                      | `output.md`                                                 |

## AI Providers

### OpenAI

- Uses the ChatGPT API
- Requires an OpenAI API key
- Set the `--api-key` flag with your OpenAI API key
- Supported models: `gpt-3.5-turbo`, `gpt-4`, `gpt-4-turbo-preview`, etc.

### Ollama

- Uses a self-hosted Ollama instance
- Supports basic authentication via URL credentials
- Set the `--ollama-url` flag with your Ollama endpoint
- URL format: `https://username:password@your-ollama-host.com`
- Supported models: `llama2`, `codellama`, `mistral`, `neural-chat`, etc.

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

### v2.1.0 - Gemini Integration

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
