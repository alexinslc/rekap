# Example Themes

This directory contains example theme files that demonstrate how to create custom color themes for rekap.

## Using These Themes

Copy any theme file to your themes directory:

```bash
mkdir -p ~/.config/rekap/themes
cp ocean.yaml ~/.config/rekap/themes/
```

Then use it:

```bash
rekap --theme ocean
```

Or use it directly without copying:

```bash
rekap --theme ./docs/themes/ocean.yaml
```

## Available Examples

### ocean.yaml
Cool blues inspired by the ocean. Deep ocean blue headers with bright cyan accents.

### sunset.yaml
Warm oranges and reds inspired by sunset. Vibrant and energetic.

### forest.yaml
Natural greens and browns inspired by the forest. Earthy and calming.

## Creating Your Own

Use these files as templates for creating your own themes. See [../THEMES.md](../THEMES.md) for a complete guide on creating custom themes.

Required color fields:
- `primary` - Main headers and titles
- `secondary` - Secondary text and labels
- `accent` - Highlighted content
- `success` - Success messages
- `error` - Error messages (can also use `warning`)
- `muted` - Subdued text
- `text` - Main body text
