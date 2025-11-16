# Configuration Guide

rekap supports a configuration file for customizing colors, display preferences, and tracking behavior. Additionally, you can use **themes** to quickly apply pre-made color schemes.

## Themes

rekap supports custom color themes that can be applied via command-line flags or theme files.

### Using Built-in Themes

rekap includes several built-in themes:

- **default** - Current magenta/cyan theme (fang aesthetic)
- **minimal** - Grayscale with minimal color
- **hacker** - Matrix-style green terminal
- **pastel** - Soft, muted pastel colors
- **nord** - Nord color scheme
- **dracula** - Dracula theme
- **solarized** - Solarized dark theme

Apply a theme using the `--theme` flag:

```bash
rekap --theme minimal
rekap demo --theme hacker
rekap --theme nord
```

### Creating Custom Themes

You can create your own theme files in YAML format:

```yaml
# ~/.config/rekap/themes/ocean.yaml
name: "Ocean"
author: "username"
colors:
  primary: "#0077be"
  secondary: "#00a8e8"
  accent: "#00c9ff"
  success: "#00ffa3"
  error: "#ff6b6b"
  muted: "#6c757d"
  text: "#ffffff"
```

### Loading Custom Themes

There are three ways to load custom themes:

1. **By name** from `~/.config/rekap/themes/`:
   ```bash
   rekap --theme ocean
   ```

2. **By relative path**:
   ```bash
   rekap --theme ./my-theme.yaml
   ```

3. **By absolute path**:
   ```bash
   rekap --theme /path/to/custom-theme.yaml
   ```

### Theme File Format

Theme files must include all required color fields:

```yaml
name: "My Theme"        # Required: theme name
author: "Your Name"     # Optional: theme author
colors:
  primary: "#ff00ff"    # Required: main title/header color
  secondary: "#00ffff"  # Required: secondary text/labels
  accent: "#ffff00"     # Required: highlights and focus items
  success: "#00ff00"    # Required: success messages
  error: "#ff0000"      # Required: errors and warnings (can also use 'warning')
  muted: "#808080"      # Required: subdued text
  text: "#ffffff"       # Required: main text color
```

Colors can be specified as:
- **Hex colors**: `"#ff00ff"`, `"#00ffff"`
- **ANSI color codes**: `"13"`, `"14"`, `"240"`

### Previewing Themes

Use demo mode to preview any theme:

```bash
rekap demo --theme dracula
rekap demo --theme ~/.config/rekap/themes/ocean.yaml
```

## Config File Location

The config file should be placed at: `~/.config/rekap/config.yaml`

## Creating Your Config

Create the directory and file:

```bash
mkdir -p ~/.config/rekap
touch ~/.config/rekap/config.yaml
```

Then edit the file with your preferred settings.

## Configuration Options

### Complete Example

```yaml
colors:
  primary: "#ff00ff"      # Main title and header color
  secondary: "#00ffff"    # Secondary text and labels
  accent: "#ffff00"       # Highlights and focus items
  success: "#00ff00"      # Success messages
  warning: "#ff0000"      # Errors and warnings
  muted: "240"            # Subdued text (ANSI color code)
  text: "255"             # Main text color (ANSI color code)

display:
  show_media: true        # Show "Now Playing" section
  show_battery: true      # Show battery information
  time_format: "12h"      # "12h" or "24h"

tracking:
  exclude_apps:
    - "Activity Monitor"
    - "System Preferences"
    - "Calendar"

accessibility:
  enabled: false          # Enable accessibility mode
  high_contrast: false    # Use high contrast colors
  no_emoji: false         # Replace emojis with text labels

domains:
  work:
    - "mycompany.atlassian.net"
    - "internal.company.com"
  distraction:
    - "news.ycombinator.com"
  neutral:
    - "gmail.com"
```

### Color Options

You can specify colors using:
- **Hex colors**: `"#ff00ff"` (magenta), `"#00ffff"` (cyan)
- **ANSI color codes**: `"13"` (bright magenta), `"14"` (cyan), `"240"` (dark gray)

Default color palette (matches fang's aesthetic):
- `primary: "13"` - Bright magenta/pink
- `secondary: "14"` - Cyan
- `accent: "11"` - Bright yellow
- `success: "10"` - Bright green
- `warning: "9"` - Bright red
- `muted: "240"` - Darker gray
- `text: "255"` - White

### Display Options

- **show_media**: Show or hide the "Now Playing" section (default: `true`)
- **show_battery**: Show or hide battery information (default: `true`)
- **time_format**: Time display format
  - `"12h"` - 12-hour format with AM/PM (e.g., "3:04 PM")
  - `"24h"` - 24-hour format (e.g., "15:04")

### Tracking Options

- **exclude_apps**: List of app names to exclude from tracking
  - Apps in this list won't appear in your top apps or focus streaks
  - Useful for filtering out system utilities or apps you don't want tracked
  - App names must match exactly as they appear in the output

### Accessibility Options

- **enabled**: Enable accessibility mode (default: `false`)
  - Adds visual markers and patterns to distinguish sections
  - Works great for color-blind users
  - Can be enabled via `--accessible` flag or config file
- **high_contrast**: Use high contrast colors (default: `false`)
  - Switches to black and white color scheme
  - Requires `enabled: true` to take effect
- **no_emoji**: Replace emojis with text labels (default: `false`)
  - Converts 🔋 to [BAT], ⏰ to [TIME], etc.
  - Useful for terminals with poor emoji support
  - Requires `enabled: true` to take effect

### Domain Categorization

rekap automatically categorizes browser tab domains into three categories:

- **work**: Development tools, documentation, project management, cloud platforms
- **distraction**: Social media, entertainment, news sites
- **neutral**: Email, uncategorized sites

Default work domains include:
- `github.com`, `gitlab.com`, `bitbucket.org`
- `stackoverflow.com`, `stackexchange.com`
- `docs.*`, `developer.*`, `api.*` (matches docs.python.org, developer.mozilla.org, etc.)
- `atlassian.net` (Jira, Confluence)
- `linear.app`, `asana.com`, `notion.so`
- `aws.amazon.com`, `console.cloud.google.com`, `portal.azure.com`

Default distraction domains include:
- `twitter.com`, `x.com`, `reddit.com`
- `facebook.com`, `instagram.com`
- `youtube.com`, `tiktok.com`, `twitch.tv`

You can override these defaults in your config:

```yaml
domains:
  work:
    - "mycompany.atlassian.net"
    - "internal.company.com"
    - "work-app.*"
  distraction:
    - "news.ycombinator.com"  # Personal preference
  neutral:
    - "gmail.com"
```

**Pattern matching:**
- Exact matches: `github.com` matches only `github.com`
- Prefix wildcards: `docs.*` matches `docs.python.org`, `docs.microsoft.com`, etc.
- Suffix wildcards: `*.google.com` matches `mail.google.com`, `drive.google.com`, etc.
- Suffix matching: `atlassian.net` matches `mycompany.atlassian.net`, `yourcompany.atlassian.net`, etc.

## Partial Configs

You don't need to specify all options. Any missing options will use defaults:

```yaml
# Minimal config - just hide media and exclude one app
display:
  show_media: false

tracking:
  exclude_apps:
    - "Slack"
```

## Examples

### Minimal & Focused

Hide media, use 24-hour time, exclude system apps:

```yaml
display:
  show_media: false
  time_format: "24h"

tracking:
  exclude_apps:
    - "Activity Monitor"
    - "System Preferences"
```

### Custom Color Scheme

Dark theme with blue accents:

```yaml
colors:
  primary: "#4a9eff"      # Light blue
  secondary: "#8be9fd"    # Cyan
  accent: "#f1fa8c"       # Yellow
  success: "#50fa7b"      # Green
  warning: "#ff5555"      # Red
  muted: "240"            # Dark gray
  text: "255"             # White
```

### Privacy-Focused

Exclude work-related apps from tracking:

```yaml
tracking:
  exclude_apps:
    - "Slack"
    - "Microsoft Teams"
    - "Zoom"
    - "Mail"
```

### Accessibility Mode

For color-blind users or high contrast needs:

```yaml
accessibility:
  enabled: true
  high_contrast: true
  no_emoji: false
```

Or use the `--accessible` flag:

```bash
rekap --accessible
rekap demo --accessible
```

Features in accessibility mode:
- Visual markers (===, >>, **, •) to distinguish sections
- High contrast black and white colors (when `high_contrast: true`)
- Text labels instead of emojis (when `no_emoji: true`)
- [OK], [ERROR], [INFO] prefixes instead of symbols
- No reliance on color alone to convey information

## Testing Your Config

Use the demo command to preview your color choices:

```bash
rekap demo
```

Or test with a specific theme:

```bash
rekap demo --theme hacker
```

## Theme vs Config

Themes and config colors work together:

- **Config file colors** - Set your default colors in `~/.config/rekap/config.yaml`
- **Theme flag** - Override colors temporarily with `--theme` flag
- **Theme flag takes precedence** - When using `--theme`, those colors override config file colors

You can use the config file for your daily theme and experiment with other themes using the `--theme` flag without modifying your config.

## Troubleshooting

### Config Not Loading

1. Check the file path: `~/.config/rekap/config.yaml`
2. Verify YAML syntax (indentation matters!)
3. Check for warnings when running rekap

### Invalid Color Values

If a color value is invalid, rekap will fall back to the default color. Colors should be:
- Hex format: `"#RRGGBB"`
- ANSI codes: `"0"` to `"255"`

### Apps Still Showing After Exclusion

Make sure the app name exactly matches what appears in rekap output:
- ❌ `"vscode"` 
- ✓ `"VS Code"`

Run rekap normally to see the exact app names, then add them to your exclude list.
