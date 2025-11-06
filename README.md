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

Coming soon via Homebrew:

```bash
brew tap alexinslc/rekap
brew install rekap
```

### Manual Installation

Download the latest release for your architecture:

```bash
# For Apple Silicon (M1/M2/M3)
curl -L https://github.com/alexinslc/rekap/releases/latest/download/rekap-darwin-arm64 -o rekap
chmod +x rekap
sudo mv rekap /usr/local/bin/

# For Intel Macs
curl -L https://github.com/alexinslc/rekap/releases/latest/download/rekap-darwin-amd64 -o rekap
chmod +x rekap
sudo mv rekap /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/alexinslc/rekap.git
cd rekap
make build
sudo make install
```

## Usage

```bash
rekap              # Today's activity summary
rekap init         # Permission setup wizard
rekap doctor       # Check capabilities and permissions
rekap demo         # See sample output with fake data
rekap --quiet      # Machine-parsable key=value output
```

### Quiet Mode Output

The `--quiet` flag outputs stable key=value pairs for scripting:

```
awake_minutes=287
boot_time=1730864122
battery_start_pct=92
battery_now_pct=68
screen_on_minutes=215
top_app_1=VS Code
top_app_1_minutes=142
focus_streak_minutes=87
focus_streak_app=VS Code
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

## Troubleshooting

**"Screen Time unavailable" message:**
- Run `rekap init` to set up Full Disk Access
- Grant permission to your terminal app in System Settings → Privacy & Security → Full Disk Access

**No app data showing:**
- Ensure Full Disk Access is granted (run `rekap doctor` to check)
- Restart your terminal after granting permissions
- macOS Screen Time must be enabled (System Settings → Screen Time)

**Binary won't run:**
- On first run, right-click the binary and select "Open" to bypass Gatekeeper
- Or run: `xattr -d com.apple.quarantine /usr/local/bin/rekap`

**Check permissions status:**
```bash
rekap doctor
```

## License

MIT
