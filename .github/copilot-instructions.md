# Copilot Instructions for rekap

## Project Overview

rekap is a single-binary macOS CLI application that provides a daily activity summary with a friendly, animated terminal UI. The project focuses on privacy-first design with no historical database, cloud sync, or telemetry—all data stays local and only today's activity is analyzed.

## Technology Stack

- **Language**: Go 1.21+
- **Platform**: macOS-specific (uses macOS APIs for system info, Screen Time, battery, etc.)
- **Dependencies**: 
  - `github.com/charmbracelet/*` - Terminal UI and styling
  - `github.com/spf13/cobra` - CLI framework
  - `modernc.org/sqlite` - SQLite for Screen Time database access
- **Build Tool**: Make

## Repository Structure

```
.
├── cmd/rekap/          # Main CLI entry point
├── internal/
│   ├── collectors/     # Data collection modules (uptime, battery, apps, etc.)
│   ├── permissions/    # macOS permission handling
│   └── ui/             # Terminal UI rendering
├── docs/               # Documentation
├── Makefile            # Build automation
├── README.md           # User documentation
└── CONTRIBUTING.md     # Contributor guidelines
```

## Build and Test Commands

### Building
```bash
make build              # Build the binary
make run                # Build and run
./rekap                 # Run with real data
./rekap demo            # Run with sample data
./rekap doctor          # Check permissions status
```

### Testing
```bash
make test               # Run all tests
go test -v ./...        # Verbose test output
```

**Important**: Some tests may fail in non-macOS environments or without proper permissions (Full Disk Access, Accessibility). This is expected behavior.

### Cleaning
```bash
make clean              # Remove build artifacts
```

## Code Style Guidelines

### Go Style
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting (auto-formatted by most Go editors)
- Run `go vet` to check for common errors
- Keep functions small and focused
- Use clear, descriptive names for variables and functions
- Minimize comments; prefer self-documenting code

### UI/UX Standards
- Use the existing color palette defined in `internal/ui/renderer.go`
- Maintain consistent spacing (2 spaces for indentation in output)
- Test output in both TTY and non-TTY environments
- Use `--quiet` mode for machine-parsable output (key=value pairs)
- Icons should be meaningful and work in most terminals

### Commit Message Format
Follow conventional commits:
```
<type>: <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example: `feat: add battery drain rate calculation`

## Development Guidelines

### macOS-Specific Considerations
- All system interactions are macOS-only (no cross-platform code)
- Screen Time data requires Full Disk Access permission
- App tracking uses SQLite database at `~/Library/Application Support/Knowledge/knowledgeC.db`
- Battery info comes from IOKit/power management APIs
- Uptime comes from kernel boot time (sysctl)

### Privacy Requirements
- Never store historical data
- Never send data to external services
- Only analyze today's activity
- All data must stay on the user's Mac
- No telemetry or analytics

### Permission Handling
The app works in degraded mode without permissions but provides more features when granted:
- **Full Disk Access**: Screen Time data, top apps, focus streaks
- **Accessibility**: Fallback for frontmost app detection
- **Media/Now Playing**: Track currently playing media

Use `rekap doctor` to check permission status during development.

### Adding New Features
1. Consider if the feature respects privacy principles
2. Ensure it works in "best-effort" mode (degrades gracefully without permissions)
3. Add tests in the appropriate `*_test.go` file
4. Update `README.md` if user-facing
5. Test with `./rekap demo` for quick validation

### Error Handling
- Collectors should return partial data if possible, not fail completely
- Use context with timeout for external system calls
- Provide helpful error messages that guide users to solutions
- Reference `rekap doctor` or `rekap init` in error messages when relevant

## Testing Strategy

### Unit Tests
- Test core logic in each package
- Mock system calls where possible
- Tests should pass on non-macOS systems for CI (skip macOS-specific features)
- Use `t.Skip()` for tests requiring specific permissions or hardware

### Manual Testing Checklist
Before submitting changes, test:
- `./rekap` - Main output
- `./rekap demo` - Demo with fake data
- `./rekap doctor` - Permission diagnostics
- `./rekap --quiet` - Machine-parsable output
- `./rekap --help` - Help text

Test with varying system states:
- With/without Full Disk Access
- Different battery levels
- Multiple apps running
- Media playing/not playing

## Common Tasks

### Adding a New Data Collector
1. Create a function in `internal/collectors/` following existing patterns
2. Add context timeout handling
3. Return best-effort data if system info unavailable
4. Add tests in `collectors_test.go`
5. Integrate into main command in `cmd/rekap/`

### Modifying UI Output
1. Update rendering functions in `internal/ui/renderer.go`
2. Maintain color consistency with existing palette
3. Test in both TTY and pipe modes (`./rekap | cat`)
4. Ensure `--quiet` mode output remains stable for scripts

### Updating Dependencies
```bash
go get -u <package>     # Update specific package
go mod tidy             # Clean up dependencies
make test               # Verify nothing broke
```

## Documentation
- Update `README.md` for user-facing changes
- Update `CONTRIBUTING.md` for process changes
- Keep code comments minimal; prefer clear code
- Update help text in CLI commands when adding features

## Release Process
- Uses GoReleaser (`.goreleaser.yml`)
- Builds for macOS arm64 and amd64
- Releases via GitHub Actions
- Planned: Homebrew tap for easy installation

## Questions or Issues?
- Check existing issues on GitHub
- Review `CONTRIBUTING.md` for guidelines
- Use descriptive issue titles
- Include steps to reproduce for bugs
