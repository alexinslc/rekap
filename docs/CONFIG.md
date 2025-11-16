# Configuration Guide

rekap supports a configuration file for customizing colors, display preferences, and tracking behavior.

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

## Testing Your Config

Use the demo command to preview your color choices:

```bash
rekap demo
```

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
