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

## Permissions

rekap works without any permissions but provides more data when granted:

| Permission | Enables | Required For |
|------------|---------|--------------|
| **Full Disk Access** | App usage, screen time, focus streaks | Top apps, screen-on time |
| **Accessibility** | Frontmost app detection | App tracking fallback |
| **Media/Now Playing** | Track currently playing media | Now Playing section |

Run `rekap init` for guided permission setup. Run `rekap doctor` to check current status.

## Privacy

All data stays on your Mac. No telemetry, no cloud sync, no historical tracking. Only today's activity is analyzed.

## Requirements

- macOS 11.0 or later

Optional permissions for full functionality (use `rekap init` to set up):
- **Full Disk Access** - Screen Time database access
- **Accessibility** - Frontmost app detection (fallback)
- **Media/Now Playing** - Track playing media

## Development

```bash
go build -o rekap ./cmd/rekap
./rekap
```

## License

MIT
