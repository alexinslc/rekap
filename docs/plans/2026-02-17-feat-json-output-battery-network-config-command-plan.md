---
title: "feat: Add JSON output, fix battery/network collectors, add config command"
type: feat
status: completed
date: 2026-02-17
---

# Add JSON Output, Fix Battery/Network Collectors, Add Config Command

## Overview

Three focused improvements to rekap: (1) structured `--json` output for scripting and automation, (2) fix battery and network collectors to show accurate today-only data, (3) add `rekap config` subcommands for config management.

## Feature 1: `--json` Output Flag

### Problem

rekap has no structured output format suitable for piping to other tools, scripts, or dashboards. The existing `--quiet` flag outputs `key=value` pairs which are hard to parse for nested data (browsers, apps, issues).

### Proposed Solution

Add a `--json` boolean flag that skips the TUI and outputs structured JSON to stdout.

**Design decisions:**

| Decision | Choice | Rationale |
|---|---|---|
| Flag name | `--json` | Matches `gh`, `docker`, `kubectl` conventions |
| Schema | Separate output structs | Internal structs expose `Error`, `Available` which are implementation details |
| Missing sections | Omitted entirely | Pointer fields + `omitempty` -- presence means available |
| Errors in JSON mode | Section omitted (non-fatal) or JSON error envelope (fatal) | Clean schema, parseable errors |
| stdout/stderr | JSON only on stdout, everything else stderr | Standard CLI convention |
| Mutual exclusion | `--json` and `--quiet` are exclusive | Cobra `MarkFlagsMutuallyExclusive` |
| `--theme`/`--accessible` with `--json` | Silently ignored | Only affect rendering |
| Schema version | `"version": "1"` field in output | Future-proofs without complexity |

### JSON Schema

```json
{
  "version": "1",
  "date": "2026-02-17",
  "collected_at": "2026-02-17T15:04:05-07:00",
  "uptime": {
    "awake_minutes": 287,
    "boot_time_unix": 1739800000
  },
  "battery": {
    "start_pct": 92,
    "current_pct": 68,
    "plug_events": 1,
    "is_plugged": false
  },
  "screen": {
    "screen_on_minutes": 240,
    "lock_count": 5,
    "avg_mins_between_locks": 48
  },
  "apps": {
    "top_apps": [
      {"name": "VS Code", "minutes": 120, "bundle_id": "com.microsoft.VSCode"}
    ],
    "total_switches": 45,
    "switches_per_hour": 8.2,
    "avg_mins_between_switches": 7.3
  },
  "focus": {
    "streak_minutes": 47,
    "app_name": "VS Code"
  },
  "media": {
    "track": "Song Name",
    "app": "Spotify"
  },
  "network": {
    "interface": "en0",
    "network_name": "HomeWiFi",
    "bytes_received": 2469606195,
    "bytes_sent": 471859200
  },
  "browsers": {
    "total_tabs": 23,
    "chrome": {"tabs": 15},
    "safari": {"tabs": 8},
    "urls_visited": 142,
    "top_domain": "github.com",
    "top_domain_visits": 34,
    "work_visits": 89,
    "distraction_visits": 23,
    "neutral_visits": 30,
    "issues_viewed": ["PROJ-123", "PROJ-456"]
  },
  "notifications": {
    "total": 47,
    "top_apps": [
      {"name": "Slack", "count": 22}
    ]
  },
  "fragmentation": {
    "score": 42,
    "level": "moderate"
  },
  "burnout": {
    "warnings": [
      {"type": "high_switching", "severity": "medium", "message": "High context switching detected"}
    ]
  },
  "context_overload": {
    "is_overloaded": true,
    "message": "Running 8 apps with 45 tabs across 12 domains"
  }
}
```

### Implementation

**Files to create:**
- `cmd/rekap/json_output.go` -- JSON output structs and `printJSON()` function

**Files to modify:**
- `cmd/rekap/main.go:19` -- Add `jsonFlag` variable
- `cmd/rekap/main.go:47` -- Pass `jsonFlag` to `runSummary`
- `cmd/rekap/main.go:52` -- Add `--json` flag definition + mutual exclusion
- `cmd/rekap/summary.go:28` -- Add `json bool` parameter to `runSummary`
- `cmd/rekap/summary.go:84-88` -- Add JSON output path alongside quiet/human

### Acceptance Criteria

- [x] `rekap --json` outputs valid JSON to stdout, exits 0
- [x] `rekap --json --quiet` produces an error before collection starts
- [x] Sections with `Available=false` are omitted from JSON output
- [x] No TUI, ANSI codes, or spinner output when `--json` is active
- [x] `rekap --json | jq .` works cleanly
- [x] `rekap --json --theme nord` silently ignores theme
- [x] Output includes `version`, `date`, `collected_at` fields
- [x] Context overload is computed and included in JSON

---

## Feature 2: Fix Battery & Network Collectors

### Problem: Battery

`battery.go:55-58` has a TODO: StartPct always equals CurrentPct, PlugCount is always 0. The battery section is functionally incomplete -- it never shows the day's battery drain or plug events.

### Problem: Network

`network.go:49` uses `netstat -ib` which reports cumulative stats since system boot, not since midnight. On machines that run for days, the "today" label is misleading.

### Proposed Solution: Battery

Parse `pmset -g log` for battery percentage and power source changes since midnight.

**Approach:**
1. Run `pmset -g log` and read output line by line
2. Filter to lines after midnight today
3. Extract first battery percentage after midnight as `StartPct`
4. Count transitions from "Battery Power" to "AC Power" as plug events (`PlugCount`)
5. If machine booted after midnight, use boot-time percentage as `StartPct`
6. On desktop Macs (no battery), skip log parsing entirely -- current `Available=false` is correct

**Key log patterns to match** (from `pmset -g log`):
- Power source change: `Using AC Power` / `Using Batt Power`
- Timestamp format: `2026-02-17 14:30:22 -0700`

**Risk: Log size.** `pmset -g log` can be large on long-running machines. Mitigations:
- Use `pmset -g log | grep -E 'Using (AC|Batt)'` to filter server-side (let the OS paginate)
- Or: pipe through context-aware line scanner that stops after processing today's entries
- Stays within the 5-second collector timeout

### Proposed Solution: Network

Store a baseline snapshot on first run of each day, calculate delta on subsequent runs.

**Approach:**
1. On each run, check for `~/.local/share/rekap/network-YYYY-MM-DD.json`
2. If file exists: read baseline, compute delta (current - baseline)
3. If file does not exist: save current stats as baseline, show since-boot values with a note
4. Add `SinceBoot bool` field to `NetworkResult` so the UI can qualify the data

**Baseline file format:**
```json
{"interface":"en0","bytes_received":1234567,"bytes_sent":456789,"timestamp":"2026-02-17T08:30:00-07:00"}
```

**Edge cases:**
- Machine rebooted (counters reset): if current < baseline, treat current as the delta (counters wrapped/reset)
- Interface changed: if active interface differs from baseline interface, show since-boot for new interface
- First run of the day: always shows since-boot (honest about data)

### Files to modify

**Battery:**
- `internal/collectors/battery.go` -- Add `parsePmsetLog()` function, update `CollectBattery` to call it

**Network:**
- `internal/collectors/network.go` -- Add baseline read/write logic, add `SinceBoot` field to `NetworkResult`
- `cmd/rekap/output.go` -- Update `printHuman` to show "(since boot)" qualifier when applicable
- `cmd/rekap/output.go` -- Update `printQuiet` to include `network_since_boot=1/0`

### Acceptance Criteria

- [x] Battery StartPct reflects first observed percentage after midnight (not current)
- [x] Battery PlugCount reflects actual AC power transitions since midnight
- [x] Battery works correctly on desktop Macs (no battery = `Available=false`, no crash)
- [x] Battery collector stays within 5-second timeout even with large pmset logs
- [x] Network shows delta from baseline when available
- [x] Network creates baseline file on first run of the day
- [x] Network shows "(since boot)" qualifier when no baseline exists
- [x] Network handles interface changes and counter resets gracefully
- [x] Baseline files are stored in `~/.local/share/rekap/`

---

## Feature 3: `rekap config` Command

### Problem

Users must manually create `~/.config/rekap/config.yaml` and know the exact YAML structure. There's no way to generate a starter config or validate an existing one.

### Proposed Solution

Add a `config` parent command with two subcommands:

**`rekap config init`**
- Creates `~/.config/rekap/config.yaml` with a commented template
- If file exists: abort with error, suggest `--force` to overwrite
- Uses a raw string template (not yaml.Marshal) to preserve comments
- Creates directory if needed

**`rekap config validate`**
- Reads and parses the config file
- Reports YAML syntax errors with line numbers (yaml.v3 provides these)
- Reports semantic errors (invalid values) as warnings instead of silently correcting
- Exits 0 if valid, 1 if errors found
- If no config file exists: exit 0 with "No config file found, defaults will be used"

**`rekap config show`** (bonus, trivial to add)
- Prints the resolved effective config (defaults + overrides)
- Useful for debugging "what config is rekap actually using?"

### Implementation

**Files to create:**
- `cmd/rekap/config_cmd.go` -- `config` parent command + `init`, `validate`, `show` subcommands

**Files to modify:**
- `cmd/rekap/main.go:106` -- Register `configCmd` alongside other commands
- `internal/config/config.go` -- Add `ValidateStrict()` method that reports errors instead of silently fixing them

### Config template (for `init`)

```yaml
# rekap configuration
# Documentation: https://github.com/alexinslc/rekap/blob/main/docs/CONFIG.md

# Colors (hex "#RRGGBB" or ANSI codes "0"-"255")
# colors:
#   primary: "13"       # Main titles
#   secondary: "14"     # Labels
#   accent: "11"        # Highlights
#   success: "10"       # Success messages
#   warning: "9"        # Warnings
#   muted: "240"        # Subdued text
#   text: "255"         # Main text

# Display options
# display:
#   show_media: true    # Show "Now Playing" section
#   show_battery: true  # Show battery information
#   time_format: "12h"  # "12h" or "24h"

# App tracking
# tracking:
#   exclude_apps:
#     - "Activity Monitor"

# Accessibility
# accessibility:
#   enabled: false
#   high_contrast: false
#   no_emoji: false

# Domain categorization
# domains:
#   work:
#     - "mycompany.atlassian.net"
#   distraction:
#     - "news.ycombinator.com"
```

### Acceptance Criteria

- [x] `rekap config init` creates config file with commented template
- [x] `rekap config init` aborts if file exists (unless `--force`)
- [x] `rekap config init` creates `~/.config/rekap/` directory if needed
- [x] `rekap config validate` reports YAML syntax errors with line info
- [x] `rekap config validate` reports invalid values (e.g., `time_format: "13h"`) as errors
- [x] `rekap config validate` exits 0 for valid config, 1 for errors
- [x] `rekap config validate` handles missing config file gracefully
- [x] `rekap config show` prints the resolved effective config
- [x] `rekap config` (no subcommand) shows help text

---

## Implementation Order

1. **Feature 3: Config command** -- Smallest scope, no collector changes, independent. Good warmup.
2. **Feature 1: JSON output** -- Moderate scope, touches output layer only. Benefits from config command existing for testing.
3. **Feature 2: Battery & network fixes** -- Largest scope, requires careful testing on real hardware. JSON output makes it easier to verify collector data.

## References

- `cmd/rekap/main.go` -- CLI structure, flag definitions
- `cmd/rekap/summary.go` -- SummaryData struct, collector orchestration
- `cmd/rekap/output.go` -- printQuiet (line 13), printHuman (line 131)
- `internal/collectors/battery.go:55-58` -- TODO for battery improvements
- `internal/collectors/network.go:49` -- netstat since-boot stats
- `internal/config/config.go` -- Config loading, validation, defaults
- `docs/CONFIG.md` -- User-facing config documentation
