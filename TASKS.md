# rekap - Development Tasks

## Phase 0: Project Setup ✅
- [x] Initialize Go module (`go mod init github.com/alexinslc/rekap`)
- [x] Create directory structure (cmd/, internal/, docs/)
- [x] Add .gitignore (binaries, IDE files, macOS artifacts)
- [x] Install Charm dependencies (`github.com/charmbracelet/bubbletea`, `lipgloss`, `spinner`)
- [x] Create basic README.md with project overview
- [x] Set up Makefile with `build`, `release`, `install` targets
- [x] Configure .goreleaser.yml for universal binary builds

## Phase 1: Core Data Collectors (Best-Effort) ✅

### 1.1 Uptime Collector ✅
- [x] Create `internal/collectors/uptime.go`
- [x] Read kernel boot time via `sysctl kern.boottime`
- [x] Calculate "awake time since midnight" (basic version)
- [ ] Parse sleep/wake events from `pmset -g log` to exclude sleep (future enhancement)
- [x] Add timeout wrapper (800-1200ms max)
- [x] Return struct: `{BootTime, AwakeMinutes, FormattedTime}`

### 1.2 Battery Collector ✅
- [x] Create `internal/collectors/battery.go`
- [ ] Parse `pmset -g log` for battery events since midnight (future enhancement)
- [x] Track: start %, current %, plug/unplug count
- [x] Alternative: Use IOKit power events if pmset insufficient
- [x] Add timeout wrapper
- [x] Return struct: `{StartPct, CurrentPct, PlugCount}`

### 1.3 Screen-On Collector ✅
- [x] Create `internal/collectors/screen.go`
- [x] Parse display power state changes from `pmset -g log`
- [x] Sum "on" intervals since midnight (exclude sleep)
- [x] Alternative: Query IOKit display power state
- [x] Add timeout wrapper
- [x] Return struct: `{ScreenOnMinutes}`

### 1.4 App Tracking Collector (Primary) ✅
- [x] Create `internal/collectors/apps.go`
- [x] Implement KnowledgeC SQLite reader
- [x] Query `~/Library/Application Support/Knowledge/knowledgeC.db`
- [x] Join `ZOBJECT` + `ZSTRUCTUREDMETADATA` tables
- [x] Filter by today's timestamp range (Core Data format)
- [x] Aggregate app usage times
- [x] Resolve bundle IDs to human-readable names
- [x] Add timeout wrapper
- [x] Return struct: `{TopApps []AppUsage, Source: "ScreenTime"}`

### 1.5 App Tracking Collector (Fallback)
- [ ] Create `internal/collectors/apps_accessibility.go`
- [ ] Implement Accessibility API frontmost app sampling
- [ ] Use `lsappinfo` to enumerate running apps
- [ ] Sample every 10-15s (persistent goroutine if running as daemon)
- [ ] For CLI mode: estimate based on `lsappinfo` active times
- [ ] Aggregate samples into time totals
- [ ] Cache bundle ID → name mappings
- [ ] Add timeout wrapper
- [ ] Return struct: `{TopApps []AppUsage, Source: "Sampling"}`

### 1.6 Focus Streak Collector ✅
- [x] Create `internal/collectors/focus.go`
- [x] Use same data source as app tracking
- [x] Find longest continuous same-app interval
- [x] Ignore switches <30s
- [x] Exclude system apps (Finder, System Settings, etc.)
- [x] Add timeout wrapper
- [x] Return struct: `{StreakMinutes, AppName}`

### 1.7 Media/Now Playing Collector ✅
- [x] Create `internal/collectors/media.go`
- [x] Research MediaRemote framework access from Go (CGo required?)
- [x] Alternative: Shell out to `nowplaying-cli` if available
- [x] Parse last played track + duration from today
- [x] Handle case where player is closed
- [x] Add timeout wrapper
- [x] Return struct: `{Track, App, DurationMinutes, Available}`

## Phase 2: Permissions System ✅

### 2.1 Permission Checker ✅
- [x] Create `internal/permissions/check.go`
- [x] Implement Accessibility permission check
- [x] Implement Full Disk Access check (try reading KnowledgeC)
- [x] Implement Now Playing permission check
- [x] Return capabilities matrix: `map[string]bool`

### 2.2 Permission Request Flow (`rekap init`) ✅
- [x] Create `internal/permissions/request.go`
- [x] Test each permission in real-time (show ✓/✗)
- [x] Open System Settings to exact pane for missing permissions
- [x] Use `open "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"`
- [x] Re-test after user grants (polling loop)
- [x] Show explanations for each permission
- [x] Never block—annotate missing permissions

### 2.3 Graceful Degradation ✅
- [x] Each collector checks its required permissions first
- [x] Return "unavailable" status if permission missing
- [x] UI shows helpful hints ("run 'rekap init' to enable...")
- [x] Never crash on permission denial

## Phase 3: CLI Commands & Orchestration

### 3.1 Main Command (`rekap`)
- [ ] Create `cmd/rekap/main.go`
- [ ] Set up CLI framework (cobra or basic flag parsing)
- [ ] Orchestrate concurrent collector execution
- [ ] Implement timeout wrapper for each collector
- [ ] Aggregate results (even if some fail)
- [ ] Pass results to UI renderer
- [ ] Default to human-friendly animated output

### 3.2 Init Command (`rekap init`)
- [ ] Create `cmd/rekap/init.go`
- [ ] Run permission setup wizard
- [ ] Show live testing with spinners
- [ ] Guide user through System Settings
- [ ] Confirm successful grants

### 3.3 Quiet Mode (`rekap --quiet`)
- [ ] Create stable key=value output format
- [ ] Skip animations, emojis, formatting
- [ ] Exit 0 with parsable output
- [ ] Document stable key names

### 3.4 Doctor Command (`rekap doctor`)
- [ ] Create `cmd/rekap/doctor.go`
- [ ] Show capabilities matrix (✓/✗ for each source)
- [ ] Test each collector's dependencies
- [ ] Show permission status
- [ ] Display helpful troubleshooting info

### 3.5 Demo Command (`rekap demo`)
- [ ] Create `cmd/rekap/demo.go`
- [ ] Generate randomized plausible data
- [ ] Show full animations/UI
- [ ] Clearly label as demo data
- [ ] Use same UI renderer as main command

## Phase 4: UI/TUI (Charmbracelet Bubbletea)

### 4.1 Base UI Components
- [ ] Create `internal/ui/renderer.go`
- [ ] Set up Bubbletea model & update loop
- [ ] Implement word-by-word typing effect for opening line
- [ ] Detect TTY vs pipe (skip animations if not TTY)
- [ ] Create lipgloss styles (colors, formatting)
- [ ] Design emoji + monochrome fallback scheme

### 4.2 Loading/Collection UI
- [ ] Create parallel spinners for each data source
- [ ] Show real-time collection status
- [ ] Handle partial results gracefully
- [ ] Timeout visualization (progress indicators)

### 4.3 Summary Display
- [ ] Render human-friendly output with emojis
- [ ] Format time durations consistently
- [ ] Show top 3 apps with times
- [ ] Display battery story
- [ ] Highlight focus streak
- [ ] Include "now playing" if available
- [ ] Show hints for missing data/permissions

### 4.4 Plain Output Mode
- [ ] Create simple text renderer (no Bubbletea)
- [ ] Output stable key=value pairs
- [ ] Ensure machine-parsable format

## Phase 5: Build & Distribution

### 5.1 Build System
- [ ] Complete Makefile with all targets
- [ ] Configure goreleaser for macOS universal binary
- [ ] Test arm64 build
- [ ] Test amd64 build
- [ ] Test universal binary
- [ ] Ensure binary is <10MB

### 5.2 Homebrew Formula
- [ ] Create separate repo: `alexinslc/homebrew-rekap`
- [ ] Write Formula/rekap.rb
- [ ] Test tap installation
- [ ] Document installation process
- [ ] Set up GitHub release workflow

### 5.3 Documentation
- [ ] Complete README.md with screenshots
- [ ] Add privacy/local-only statement
- [ ] Create permissions table
- [ ] Write installation instructions
- [ ] Add troubleshooting section
- [ ] Include quick start guide
- [ ] Add sample output examples

## Phase 6: Testing & Polish

### 6.1 Acceptance Testing
- [ ] Test with no permissions granted
- [ ] Test after `rekap init` with all permissions
- [ ] Test `rekap --quiet` output stability
- [ ] Test `rekap doctor` output
- [ ] Test `rekap demo` animations
- [ ] Test on fresh macOS install

### 6.2 Error Handling
- [ ] Verify no crashes on permission denial
- [ ] Test timeout behavior for slow collectors
- [ ] Test malformed data handling
- [ ] Test missing binary dependencies
- [ ] Verify graceful degradation in all cases

### 6.3 Performance
- [ ] Ensure total runtime <2 seconds
- [ ] Optimize SQLite queries
- [ ] Profile collector timeouts
- [ ] Test on older Mac hardware

### 6.4 Edge Cases
- [ ] Test immediately after midnight
- [ ] Test with no activity today
- [ ] Test on freshly booted system
- [ ] Test with battery at 100%
- [ ] Test with no apps open
- [ ] Test with system apps only

## Phase 7: Nice-to-Haves (Post-MVP)

- [ ] Add `--no-emoji` flag
- [ ] Add `--no-animate` flag
- [ ] Implement "Good morning" Kismet moment (within 3min of wake)
- [ ] Consider `--json` output format
- [ ] Optimize bundle ID → name resolution caching

---

## Development Priority Order

**Week 1: Foundation**
1. Project setup (Phase 0)
2. Uptime collector (1.1)
3. Battery collector (1.2)
4. Basic CLI orchestration (3.1)
5. Simple text output (4.4)

**Week 2: Core Functionality**
6. Screen-on collector (1.3)
7. Permissions system (2.1, 2.2)
8. App tracking - KnowledgeC primary (1.4)
9. Focus streak (1.6)

**Week 3: UI & Polish**
10. Bubbletea UI implementation (4.1, 4.2, 4.3)
11. All commands (init, doctor, demo)
12. App tracking fallback (1.5)
13. Media collector (1.7)

**Week 4: Distribution**
14. Build system & testing (Phase 5 & 6)
15. Documentation & release (5.3)

---

## Quick Start Implementation Path

To get something working quickly:

1. ✅ Initialize Go project
2. ✅ Create basic main.go with CLI flags
3. ✅ Implement uptime collector (easiest, no permissions)
4. ✅ Implement battery collector (no permissions)
5. ✅ Create simple text output
6. ✅ Test end-to-end with 2 collectors
7. → Then expand to permissions, more collectors, and TUI

---

## Notes

- **Bubbletea vs Fang**: Spec mentions "Fang" but that doesn't exist in Charmbracelet. Assuming Bubbletea + Bubbles + Lipgloss.
- **MediaRemote**: May require private APIs or CGo. Consider `nowplaying-cli` as primary approach.
- **Time Zone**: All "today" calculations should use local time zone.
- **Sleep Tracking**: Parsing `pmset -g log` is critical for accurate "awake time" calculations.
