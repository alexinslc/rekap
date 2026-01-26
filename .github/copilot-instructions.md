# GitHub Copilot Instructions for rekap

## Project Overview
rekap is a macOS CLI that summarizes your daily computer activity in a beautiful terminal UI. Built with Go using Charmbracelet's fang library.

## Tech Stack
- **Language**: Go 1.25+
- **TUI**: Charmbracelet (fang, lipgloss)
- **CLI Framework**: Cobra
- **Database**: SQLite (modernc.org/sqlite)
- **Platform**: macOS only

## Project Structure
```
rekap/
├── cmd/           # Cobra CLI commands
├── internal/      # Internal packages
├── docs/          # Documentation
├── go.mod         # Go module definition
└── Makefile       # Build automation
```

## Commands
```bash
make build         # Build binary
make install       # Install to /usr/local/bin
go run main.go     # Run directly
```

## Features
- Uptime & awake time tracking
- Battery usage monitoring
- Top apps by usage time
- Screen-on time
- Focus streak detection
- Now Playing tracking (optional)
- Network activity summary

## Code Conventions
- Follow Go idioms and effective Go guidelines
- Use Charmbracelet styling for terminal output
- Keep it macOS-specific (uses system APIs)
- Single binary, no external dependencies at runtime
- Today only, local only, best-effort data collection
