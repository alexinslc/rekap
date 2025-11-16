# Themes Guide

rekap supports custom color themes to personalize your terminal output. This guide covers everything you need to know about using and creating themes.

## Quick Start

### Using Built-in Themes

rekap comes with 7 built-in themes ready to use:

```bash
rekap --theme default      # Magenta/cyan (default)
rekap --theme minimal      # Grayscale
rekap --theme hacker       # Matrix green
rekap --theme pastel       # Soft pastels
rekap --theme nord         # Nord color scheme
rekap --theme dracula      # Dracula theme
rekap --theme solarized    # Solarized dark
```

### Preview Themes

Use demo mode to preview any theme without collecting real data:

```bash
rekap demo --theme hacker
rekap demo --theme dracula
```

## Built-in Themes

### default
The original rekap color scheme with bright magenta and cyan accents. Inspired by the fang CLI aesthetic.

```bash
rekap --theme default
```

### minimal
A clean, grayscale theme with minimal color. Perfect for distraction-free viewing or professional screenshots.

```bash
rekap --theme minimal
```

### hacker
Matrix-style green terminal theme. For when you want to feel like you're in a 90s hacker movie.

```bash
rekap --theme hacker
```

### pastel
Soft, muted pastel colors. Easy on the eyes for long viewing sessions.

```bash
rekap --theme pastel
```

### nord
Based on the popular Nord color palette. Cool, arctic-inspired blues and greens.

```bash
rekap --theme nord
```

### dracula
The famous Dracula theme. Dark background with vibrant pink and cyan accents.

```bash
rekap --theme dracula
```

### solarized
Based on Solarized Dark. Precision colors for professionals.

```bash
rekap --theme solarized
```

## Creating Custom Themes

### Theme File Format

Themes are YAML files with the following structure:

```yaml
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

### Required Fields

- **name**: Display name of the theme
- **colors.primary**: Main headers and titles
- **colors.secondary**: Secondary text and labels
- **colors.accent**: Highlighted content (focus streaks, etc.)
- **colors.success**: Success messages and positive indicators
- **colors.error** or **colors.warning**: Error messages and warnings
- **colors.muted**: Subdued text (hints, metadata)
- **colors.text**: Main body text

### Optional Fields

- **author**: Theme creator's name

### Color Formats

Themes support two color formats:

**Hex colors** (24-bit true color):
```yaml
primary: "#ff00ff"    # Magenta
secondary: "#00ffff"  # Cyan
```

**ANSI color codes** (256-color palette):
```yaml
primary: "13"    # Bright magenta
secondary: "14"  # Cyan
muted: "240"     # Dark gray
```

## Installing Custom Themes

### Method 1: Themes Directory (Recommended)

1. Create the themes directory:
   ```bash
   mkdir -p ~/.config/rekap/themes
   ```

2. Save your theme file:
   ```bash
   # ~/.config/rekap/themes/ocean.yaml
   ```

3. Use by name:
   ```bash
   rekap --theme ocean
   ```

### Method 2: Absolute Path

Save your theme anywhere and reference it by full path:

```bash
rekap --theme /path/to/my-theme.yaml
```

### Method 3: Relative Path

Use a theme file relative to your current directory:

```bash
rekap --theme ./themes/custom.yaml
```

## Example Themes

### Ocean Theme

Cool blues inspired by the ocean:

```yaml
name: "Ocean"
author: "rekap"
colors:
  primary: "#0077be"
  secondary: "#00a8e8"
  accent: "#00c9ff"
  success: "#00ffa3"
  error: "#ff6b6b"
  muted: "#6c757d"
  text: "#ffffff"
```

### Sunset Theme

Warm oranges and reds:

```yaml
name: "Sunset"
author: "rekap"
colors:
  primary: "#ff6b35"
  secondary: "#f7931e"
  accent: "#fdc500"
  success: "#c1d82f"
  error: "#e63946"
  muted: "#6c757d"
  text: "#ffffff"
```

### Forest Theme

Natural greens and browns:

```yaml
name: "Forest"
author: "rekap"
colors:
  primary: "#2d5016"
  secondary: "#4a7c59"
  accent: "#87a96b"
  success: "#76b947"
  error: "#8b4513"
  muted: "#556b2f"
  text: "#f5f5dc"
```

### Cyberpunk Theme

Neon pinks and blues:

```yaml
name: "Cyberpunk"
author: "rekap"
colors:
  primary: "#ff00ff"
  secondary: "#00ffff"
  accent: "#ffff00"
  success: "#00ff00"
  error: "#ff0066"
  muted: "#8b00ff"
  text: "#ffffff"
```

## Sharing Themes

Want to share your theme with the community?

1. Create a GitHub Gist with your theme YAML
2. Share the raw URL
3. Others can download and use it:
   ```bash
   curl -o ~/.config/rekap/themes/mytheme.yaml https://gist.githubusercontent.com/...
   rekap --theme mytheme
   ```

## Tips for Creating Themes

### Color Selection

- **Contrast**: Ensure sufficient contrast between text and background
- **Readability**: Test your theme in both bright and dim lighting
- **Accessibility**: Consider colorblind-friendly palettes
- **Consistency**: Use colors from the same palette for harmony

### Testing Themes

Always test your theme with demo mode before using it with real data:

```bash
rekap demo --theme ./my-new-theme.yaml
```

### Color Tools

Useful tools for selecting colors:

- [coolors.co](https://coolors.co) - Color scheme generator
- [paletton.com](https://paletton.com) - Color palette designer
- [nordtheme.com](https://www.nordtheme.com) - Nord palette
- [draculatheme.com](https://draculatheme.com) - Dracula colors

### ANSI Color Reference

Common ANSI 256 color codes:

- `0-7`: Standard colors (black, red, green, yellow, blue, magenta, cyan, white)
- `8-15`: Bright variants of standard colors
- `16-231`: 216 colors (6×6×6 RGB cube)
- `232-255`: Grayscale from dark to light

## Troubleshooting

### Theme Not Found

If you get "failed to load theme" error:

1. Check the theme name spelling
2. Verify the file exists at `~/.config/rekap/themes/<name>.yaml`
3. Try using the full path instead

### Colors Look Wrong

- Verify your terminal supports 256 colors or true color
- Check that color values are properly quoted in YAML
- Test with demo mode first

### Theme Not Loading

1. Validate YAML syntax (use a YAML validator)
2. Ensure all required color fields are present
3. Check for typos in field names

### Invalid Color Values

Colors must be:
- Hex format: `"#RRGGBB"` (with quotes)
- ANSI codes: `"0"` to `"255"` (with quotes)

## Theme vs Config

Understanding the relationship between themes and config:

- **Config colors** (`~/.config/rekap/config.yaml`) - Your default colors
- **Theme flag** (`--theme`) - Temporary color override
- **Precedence**: `--theme` > config colors > built-in defaults

You can set a default color scheme in your config and experiment with themes using the `--theme` flag without modifying your config file.
