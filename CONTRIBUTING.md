# Contributing to lockr

Thank you for considering contributing to lockr! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and constructive in all interactions. We're all here to build something useful together.

## How to Contribute

### Reporting Bugs

1. Check existing [issues](https://github.com/devops-chris/lockr/issues) to avoid duplicates
2. Use the bug report template
3. Include:
   - lockr version (`lockr version`)
   - Operating system and version
   - Steps to reproduce
   - Expected vs actual behavior
   - Any error messages

### Suggesting Features

1. Check existing issues and discussions
2. Open a new issue with the "feature request" label
3. Describe the use case and expected behavior

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Run linter (`golangci-lint run`)
6. Commit with clear messages
7. Push to your fork
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21+
- AWS credentials configured (for testing)

### Building

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/lockr.git
cd lockr

# Build
go build -o lockr .

# Run tests
go test ./...

# Run linter
golangci-lint run
```

### Project Structure

```
lockr/
├── cmd/              # CLI commands
│   ├── root.go       # Root command and config
│   ├── write.go      # Write command
│   ├── read.go       # Read command
│   ├── list.go       # List command
│   ├── delete.go     # Delete command
│   └── version.go    # Version command
├── internal/
│   ├── config/       # Configuration handling
│   └── ssm/          # AWS SSM client
├── main.go           # Entry point
├── go.mod
└── go.sum
```

### Coding Standards

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write clear, concise comments
- Add tests for new functionality
- Keep functions small and focused

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for AWS Secrets Manager
fix: handle empty values in write command
docs: update installation instructions
refactor: simplify path building logic
```

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GoReleaser handles:
- Building binaries for all platforms
- Creating GitHub releases
- Updating Homebrew tap

## Questions?

Open an issue or start a discussion. We're happy to help!

