# AGENTS.md - rekap

## Project Summary
rekap is a macOS daily activity summary CLI with an animated terminal UI. Built with Go and Charmbracelet libraries.

## Tech Stack
| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| TUI | Charmbracelet (fang, lipgloss) |
| CLI | Cobra |
| Database | SQLite (modernc.org/sqlite) |
| Platform | macOS only |

## Directory Structure
```
rekap/
├── cmd/           # CLI commands (Cobra)
├── internal/      # Internal packages
├── docs/          # Documentation
├── go.mod/sum     # Go dependencies
├── Makefile       # Build commands
└── .goreleaser.yml # Release config
```

## Commands
| Command | Purpose |
|---------|---------|
| `make build` | Build binary |
| `make install` | Install to /usr/local/bin |
| `go run main.go` | Run directly |
| `goreleaser release --snapshot` | Build release |

## Code Conventions
1. Follow standard Go idioms
2. Use Charmbracelet for all terminal styling
3. Keep macOS-specific code in isolated packages
4. Return errors, don't panic
5. Single binary with no runtime dependencies

## Design Philosophy
- **Today only**: Current day stats, no history
- **Local only**: No cloud, no telemetry
- **Best-effort**: Graceful degradation for missing data

## Features
- Uptime/awake tracking
- Battery monitoring
- Top 3 apps by usage
- Screen-on time
- Focus streaks
- Now Playing (optional)
- Network activity
