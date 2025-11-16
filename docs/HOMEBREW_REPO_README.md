# Homebrew Tap for rekap

This is the official Homebrew tap for [rekap](https://github.com/alexinslc/rekap), a daily Mac activity summary CLI tool.

## Installation

```bash
brew tap alexinslc/rekap
brew install rekap
```

Or install directly:

```bash
brew install alexinslc/rekap/rekap
```

## About

rekap is a single-binary macOS CLI application that provides a daily activity summary with a friendly, animated terminal UI. It's privacy-first, with no historical database, cloud sync, or telemetry—all data stays local.

### Features

- Daily activity summary (uptime, battery, screen time)
- Top 3 apps by usage time
- Focus streak detection
- Browser tab tracking (Chrome, Safari, Edge)
- Network activity summary
- Now Playing tracking (optional)

## Usage

After installation:

```bash
rekap                     # Today's activity summary
rekap demo                # See sample output
rekap doctor              # Check permissions
rekap --help              # Full help
```

## Requirements

- macOS 11.0 or later
- Optional: Full Disk Access for detailed app usage data

## Documentation

- [Main Repository](https://github.com/alexinslc/rekap)
- [README](https://github.com/alexinslc/rekap#readme)
- [Contributing Guide](https://github.com/alexinslc/rekap/blob/main/CONTRIBUTING.md)

## Formula Maintenance

The formula in this repository is automatically updated by [GoReleaser](https://goreleaser.com/) when a new version of rekap is released. Manual changes to the formula will be overwritten on the next release.

If you notice an issue with the formula, please:
1. Open an issue in the [main rekap repository](https://github.com/alexinslc/rekap/issues)
2. Mention that it's related to the Homebrew formula
3. The maintainer will update the GoReleaser configuration and republish

## License

MIT - see [LICENSE](https://github.com/alexinslc/rekap/blob/main/LICENSE) in the main repository.
