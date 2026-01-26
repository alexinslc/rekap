# CLAUDE.md - rekap

## What is this?
rekap is a macOS CLI that displays a daily summary of your computer activity in an animated terminal UI. Shows uptime, battery usage, top apps, screen time, focus streaks, and network activity.

## Tech Stack
- **Language**: Go 1.25+
- **TUI**: Charmbracelet (fang for terminal UI, lipgloss for styling)
- **CLI Framework**: Cobra
- **Database**: SQLite (modernc.org/sqlite - pure Go driver)
- **Platform**: macOS only (uses system APIs)

## Project Structure
```
rekap/
├── cmd/           # Cobra CLI commands entry point
├── internal/      # Internal packages (collectors, ui, etc.)
├── docs/          # Documentation
├── go.mod         # Dependencies
├── go.sum         # Dependency checksums
└── Makefile       # Build automation
```

## Development
```bash
# Build
make build

# Install locally
make install         # Installs to /usr/local/bin

# Run directly
go run main.go

# Build for release (uses goreleaser)
goreleaser release --snapshot --clean
```

## Design Principles
- **Today only**: No historical database, shows current day only
- **Local only**: No cloud sync, no telemetry
- **Best-effort**: Gracefully handles missing data
- **Single binary**: No external runtime dependencies

## Features Tracked
- Uptime & awake time
- Battery usage
- Top 3 apps by usage time
- Screen-on time
- Focus streak detection
- Now Playing (optional)
- Network activity (data transferred, active connection)

## Code Conventions
- Standard Go project layout
- Use Charmbracelet components for TUI
- Error handling: return errors, don't panic
- Keep macOS-specific code isolated
