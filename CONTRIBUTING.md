# Contributing to PullPoet

Thank you for your interest in contributing to PullPoet! We welcome contributions from everyone.

## Code of Conduct

By participating in this project, you are expected to uphold our code of conduct. Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Issues

- Before creating an issue, please search existing issues to avoid duplicates
- Use the issue templates when available
- Provide clear steps to reproduce the problem
- Include relevant system information (Go version, OS, etc.)

### Submitting Changes

1. **Fork the repository**
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** with clear, descriptive commit messages
4. **Add tests** for new functionality
5. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   ```
6. **Update documentation** if needed
7. **Submit a pull request**

### Pull Request Guidelines

- Follow the existing code style and conventions
- Write clear, descriptive commit messages
- Include tests for new features
- Update documentation for any new functionality
- Keep pull requests focused on a single feature or fix
- Reference related issues in your PR description

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Local Development

1. Clone your fork:

   ```bash
   git clone https://github.com/yourusername/pullpoet.git
   cd pullpoet
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Build the project:

   ```bash
   go build -o pullpoet cmd/main.go
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

## Code Style

- Follow standard Go conventions and formatting
- Use `gofmt` to format your code
- Run `go vet` to check for common issues
- Write meaningful variable and function names
- Add comments for complex logic

## Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage of new code
- Test both success and error scenarios

## Documentation

- Update README.md for new features or changes to usage
- Add inline code comments for complex logic
- Update CLI help text if adding new flags or commands

## Questions?

Feel free to open an issue with the "question" label if you need help or clarification.

Thank you for contributing to PullPoet! 🎉
