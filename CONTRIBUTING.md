# Contributing to rekap

Thank you for your interest in contributing to rekap! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Issue Guidelines](#issue-guidelines)

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to make rekap better!

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally: `git clone https://github.com/YOUR-USERNAME/rekap.git`
3. Add the upstream repository: `git remote add upstream https://github.com/alexinslc/rekap.git`

## Development Setup

### Prerequisites

- Go 1.21 or later
- macOS (rekap is macOS-specific)
- Make

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Running Locally

```bash
./rekap          # Run with real data
./rekap demo     # Run with sample data
./rekap doctor   # Check permissions
./rekap init     # Set up permissions
```

## How to Contribute

### Finding an Issue

- Look for issues labeled `good first issue` for beginner-friendly tasks
- Check issues labeled `help wanted` for more challenging work
- Comment on an issue to let others know you're working on it

### Creating a New Issue

Before creating a new issue:
- Search existing issues to avoid duplicates
- Use issue templates when available
- Provide clear descriptions and examples
- For bugs, include steps to reproduce

## Coding Standards

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format your code
- Run `go vet` to check for common errors
- Keep functions small and focused
- Write clear, descriptive variable and function names

### UI/UX Guidelines

- Use the existing color palette (defined in `internal/ui/renderer.go`)
- Maintain consistent spacing and indentation (2 spaces)
- Test output in both TTY and non-TTY environments
- Ensure animations are smooth and don't slow down the app
- Icons should be meaningful and universally understood

### Commit Messages

Follow the conventional commits format:

```
<type>: <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Example:
```
feat: add weekly summary report

Add a new --week flag that shows activity summary for the past 7 days
including daily breakdowns and trends.

Closes #42
```

## Testing

### Unit Tests

- Write tests for new functionality
- Ensure existing tests pass: `make test`
- Aim for meaningful test coverage, especially for core logic

### Manual Testing

Before submitting a PR, test:
- `./rekap` - Main command
- `./rekap demo` - Demo mode
- `./rekap doctor` - Permissions check
- `./rekap init` - Permission setup
- `./rekap --help` - Help output
- `./rekap --quiet` - Machine-parsable output

Test in different scenarios:
- With and without Full Disk Access
- With different battery states
- With various apps running
- With and without media playing

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, well-documented code
   - Add tests for new functionality
   - Update documentation as needed

3. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: your feature description"
   ```

4. **Keep your branch updated**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Use a clear, descriptive title
   - Reference related issues (e.g., "Closes #123")
   - Describe what changed and why
   - Include screenshots for UI changes
   - Check all PR template boxes

### PR Review Process

- Maintainers will review your PR
- Address any requested changes
- Once approved, a maintainer will merge your PR
- Your contribution will be credited in the release notes

## Releasing

Releases are automated using GitHub Actions and GoReleaser.

### Creating a Release

1. **Tag the release:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Build binaries for Intel Mac (amd64) and Apple Silicon (arm64)
   - Create a universal binary
   - Generate checksums
   - Create release archives (.tar.gz)
   - Generate release notes from commits
   - Publish to GitHub Releases
   - Update Homebrew formula (if configured)

### Release Artifacts

Each release includes:
- `rekap-darwin-amd64` - Standalone binary for Intel Mac
- `rekap-darwin-arm64` - Standalone binary for Apple Silicon
- `rekap-darwin-all` - Universal binary (works on both architectures)
- `rekap-v1.0.0-darwin-amd64.tar.gz` - Archive with binary and documentation
- `rekap-v1.0.0-darwin-arm64.tar.gz` - Archive with binary and documentation
- `checksums.txt` - SHA256 checksums for verification
- Auto-generated release notes

### Testing Release Process

To test the release process without publishing:

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Test with snapshot (doesn't publish)
goreleaser release --snapshot --clean --skip=publish

# Check generated artifacts
ls -lh dist/
```

## Issue Guidelines

### Good Issue Titles

✅ Good:
- "Add daily notification feature"
- "Battery percentage not updating correctly"
- "Improve error message for missing permissions"

❌ Bad:
- "Bug"
- "Feature request"
- "It doesn't work"

### Issue Descriptions

Include:
- **What**: What is the issue or feature?
- **Why**: Why is this important?
- **How**: (Optional) Suggest implementation approach
- **Examples**: Code, screenshots, or examples

## Questions?

- Open a discussion on GitHub
- Comment on relevant issues
- Reach out to maintainers

Thank you for contributing to rekap! 🎉
