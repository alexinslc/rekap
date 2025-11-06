# rekap

Daily Mac Activity Summary - A single-binary macOS CLI that summarizes today's computer activity in a friendly, animated terminal UI.

## Features

- **Today only, local only, best-effort only** - No historical database, no cloud sync, no telemetry
- Uptime & awake time tracking
- Battery usage monitoring
- Top 3 apps by usage time
- Screen-on time calculation
- Focus streak detection
- Now Playing tracking (optional)

## Installation

Coming soon - will be available via Homebrew:

```bash
brew tap alexinslc/rekap
brew install rekap
```

## Usage

```bash
rekap              # Default: today's summary with animations
rekap init         # Permission setup wizard
rekap --quiet      # Machine-parsable key=value output
rekap doctor       # Show capabilities matrix
rekap demo         # Sample output with randomized data
```

## Privacy

All data stays on your Mac. No telemetry, no cloud sync, no historical tracking. Only today's activity is analyzed.

## Requirements

- macOS 11.0 or later
- Optional permissions for full functionality:
  - **Accessibility** - Sample frontmost app/window (fallback method)
  - **Full Disk Access** - Read Screen Time database (primary method)
  - **Media/Now Playing** - Track currently/last played media

## Development

```bash
go build -o rekap ./cmd/rekap
./rekap
```

## License

MIT
