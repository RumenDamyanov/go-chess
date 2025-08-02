# Contributing to go-chess

Thank you for your interest in contributing to go-chess! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a Code of Conduct to ensure a welcoming environment for all contributors. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- A GitHub account

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-chess.git
   cd go-chess
   ```

3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/rumendamyanov/go-chess.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Run tests to ensure everything works:
   ```bash
   go test ./...
   ```

## Development Workflow

### Branch Naming

- `feature/description` - for new features
- `bugfix/description` - for bug fixes
- `docs/description` - for documentation updates
- `refactor/description` - for code refactoring

### Making Changes

1. Create a new branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the coding standards below

3. Write or update tests for your changes

4. Run the test suite:
   ```bash
   go test ./...
   ```

5. Run linting:
   ```bash
   golangci-lint run
   ```

6. Commit your changes with a descriptive commit message:
   ```bash
   git commit -m "feat: add new chess piece validation"
   ```

7. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

8. Create a Pull Request on GitHub

## Coding Standards

### Go Code Style

- Follow standard Go formatting with `gofmt`
- Use `goimports` for import organization
- Follow effective Go guidelines
- Write clear, self-documenting code
- Use meaningful variable and function names

### Code Organization

- Keep functions small and focused
- Use clear interfaces
- Separate concerns appropriately
- Follow the existing project structure

### Documentation

- Add godoc comments for all exported functions, types, and constants
- Include examples in documentation where helpful
- Update README.md if adding new features
- Add or update relevant wiki pages

### Testing

- Write unit tests for all new functionality
- Aim for high test coverage (>80%)
- Include table-driven tests where appropriate
- Add benchmarks for performance-critical code
- Test edge cases and error conditions

Example test structure:
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        hasError bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Commit Messages

Use conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

## Pull Request Process

### Before Submitting

- Ensure all tests pass
- Run static analysis tools
- Update documentation if needed
- Add changelog entry if appropriate
- Rebase on latest upstream/master

### Pull Request Description

Include:
- Clear description of changes
- Motivation for the changes
- Any breaking changes
- Testing performed
- Screenshots for UI changes (if applicable)

### Review Process

- All PRs require at least one review
- Address reviewer feedback promptly
- Keep PRs focused and reasonably sized
- Be open to suggestions and improvements

## Types of Contributions

### Bug Reports

When reporting bugs, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant code snippets
- Error messages

### Feature Requests

For new features:
- Describe the use case
- Explain why it's valuable
- Consider implementation complexity
- Discuss potential breaking changes

### Code Contributions

Areas where contributions are especially welcome:
- Chess engine improvements
- AI algorithm enhancements
- Performance optimizations
- Additional API endpoints
- Frontend integration examples
- Documentation improvements
- Test coverage expansion

### Documentation

- API documentation improvements
- Tutorial additions
- Code examples
- Wiki page updates
- README enhancements

## Project Structure

```
go-chess/
├── ai/              # AI engine implementations
├── api/             # HTTP API handlers
├── config/          # Configuration management
├── engine/          # Core chess engine
├── examples/        # Example applications
├── docs/            # Additional documentation
├── .github/         # GitHub workflows and templates
└── tests/           # Integration tests
```

## Performance Considerations

- Chess engines are performance-critical
- Profile code before optimizing
- Use benchmarks to measure improvements
- Consider memory allocation patterns
- Cache expensive computations when appropriate

## Security

- Report security vulnerabilities privately
- Follow secure coding practices
- Validate all inputs
- Avoid exposing sensitive information in logs
- Use context for timeouts and cancellation

## Release Process

1. Update version numbers
2. Update CHANGELOG.md
3. Create release notes
4. Tag the release
5. Update documentation

## Getting Help

- Check existing issues and PRs
- Look at the documentation and examples
- Ask questions in issue discussions
- Join community discussions

## Recognition

Contributors will be recognized in:
- CHANGELOG.md
- GitHub contributors list
- Release notes for significant contributions

Thank you for contributing to go-chess!
