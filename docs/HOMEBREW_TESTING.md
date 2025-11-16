# Testing Homebrew Installation

This document provides a testing checklist for validating the Homebrew tap installation on clean macOS systems.

## Prerequisites

- macOS 11.0 or later
- Homebrew installed (`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
- Internet connection

## Test Environments

Test on at least one of each:
- [ ] Apple Silicon Mac (M1/M2/M3/M4) running macOS 11+
- [ ] Intel Mac running macOS 11+
- [ ] macOS 15+ (latest)
- [ ] macOS 14 (Sonoma)
- [ ] macOS 13 (Ventura)

## Clean System Testing

### Initial State Verification

Before testing, verify the system is in a clean state:

```bash
# Verify rekap is not installed
which rekap
# Should output nothing or "rekap not found"

# Verify no previous tap exists
brew tap | grep alexinslc/rekap
# Should output nothing

# Update Homebrew
brew update
```

### Installation Tests

#### Test 1: Install via Tap (Two-Step)

```bash
# Step 1: Add the tap
brew tap alexinslc/rekap

# Verify tap was added
brew tap | grep alexinslc/rekap
# Should output: alexinslc/rekap

# Step 2: Install rekap
brew install rekap

# Verify installation
which rekap
# Should output: /usr/local/bin/rekap (Intel) or /opt/homebrew/bin/rekap (Apple Silicon)

rekap --version
# Should output: rekap version X.Y.Z

# Test basic functionality
rekap demo
# Should display sample data with colored output
```

#### Test 2: Direct Installation (One-Step)

First, clean up from Test 1:

```bash
brew uninstall rekap
brew untap alexinslc/rekap
```

Then test direct installation:

```bash
# Install directly without tapping first
brew install alexinslc/rekap/rekap

# Verify installation
which rekap
rekap --version
rekap demo
```

#### Test 3: Upgrade

After a new version is released:

```bash
# Check current version
rekap --version

# Update Homebrew
brew update

# Check for updates
brew outdated | grep rekap

# Upgrade
brew upgrade rekap

# Verify new version
rekap --version
# Should show the new version number
```

#### Test 4: Reinstall

```bash
# Reinstall to test formula robustness
brew reinstall rekap

# Verify it still works
rekap --version
rekap demo
```

#### Test 5: Uninstall

```bash
# Uninstall
brew uninstall rekap

# Verify removal
which rekap
# Should output nothing or "rekap not found"

# Remove tap
brew untap alexinslc/rekap

# Verify tap removal
brew tap | grep alexinslc/rekap
# Should output nothing
```

### Functional Tests

After installing rekap via Homebrew:

#### Basic Commands

```bash
# Version check
rekap --version
# Should output version information

# Help
rekap --help
# Should display help text

# Demo mode
rekap demo
# Should display sample data

# Doctor command
rekap doctor
# Should show permission status

# Quiet mode
rekap --quiet
# Should output key=value pairs
```

#### Shell Completion

```bash
# Test completion generation
rekap completion zsh > /tmp/test_completion.zsh
cat /tmp/test_completion.zsh
# Should contain zsh completion code

rekap completion bash > /tmp/test_completion.bash
cat /tmp/test_completion.bash
# Should contain bash completion code

rekap completion fish > /tmp/test_completion.fish
cat /tmp/test_completion.fish
# Should contain fish completion code

# Clean up
rm /tmp/test_completion.*
```

#### Real Usage (Optional)

If you want to test with actual system data:

```bash
# Run without permissions (should work in degraded mode)
rekap
# Should display basic info (battery, uptime, etc.)

# Run permission setup
rekap init
# Should guide through permission setup

# Grant permissions as prompted, then run again
rekap
# Should display more detailed information
```

## Formula Validation

### Check Formula Content

```bash
# View the formula
brew cat rekap

# Verify it contains:
# - Correct homepage URL
# - Proper description
# - MIT license
# - Test block with version check
# - Install block with bin.install
```

### Formula Audit

```bash
# Run Homebrew's audit tool
brew audit --new rekap

# Should pass with no errors
# Minor warnings about formula style are acceptable
```

### Formula Test

```bash
# Run the formula's test block
brew test rekap

# Should pass - this runs: system "#{bin}/rekap --version"
```

## Edge Cases

### Test with Homebrew in Different Locations

- Intel Mac: `/usr/local/bin/rekap`
- Apple Silicon: `/opt/homebrew/bin/rekap`
- Custom location: Verify PATH is set correctly

### Test Conflicts

```bash
# If rekap was manually installed to /usr/local/bin
# Check for conflicts
ls -la /usr/local/bin/rekap

# Homebrew should warn about files in the way
# Follow Homebrew's instructions to resolve
```

### Test Permissions

```bash
# Verify binary is executable
ls -la $(which rekap)
# Should show executable permissions (e.g., -rwxr-xr-x)

# Try running without full path
rekap --version
# Should work if PATH is set correctly
```

## Checklist Summary

After completing all tests, verify:

- [ ] Tap can be added successfully
- [ ] rekap installs via `brew tap` + `brew install`
- [ ] rekap installs via direct formula `brew install alexinslc/rekap/rekap`
- [ ] `rekap --version` works
- [ ] `rekap demo` displays output correctly
- [ ] `rekap --help` shows help text
- [ ] `brew upgrade rekap` works (after new release)
- [ ] `brew reinstall rekap` works
- [ ] `brew uninstall rekap` removes the binary
- [ ] `brew tap` and `brew untap` work correctly
- [ ] Formula passes `brew audit`
- [ ] Formula test passes `brew test rekap`
- [ ] Binary is placed in correct Homebrew bin directory
- [ ] Works on both Intel and Apple Silicon Macs
- [ ] No errors or warnings during installation

## Reporting Issues

If any test fails, document:

1. **System Information:**
   - macOS version: `sw_vers`
   - Architecture: `uname -m`
   - Homebrew version: `brew --version`
   - Homebrew location: `brew --prefix`

2. **Test that Failed:**
   - Which test failed
   - Exact command run
   - Full error output
   - Expected vs actual behavior

3. **Formula Information:**
   - Formula URL: `brew info rekap`
   - Formula content: `brew cat rekap`

4. **Logs:**
   - Installation log: `brew install rekap -v`
   - Debug log: `brew install rekap --debug`

## Continuous Testing

For ongoing validation:

1. **Before Each Release:**
   - Test locally with `goreleaser release --snapshot --clean --skip=publish`
   - Review generated formula in `dist/homebrew/Formula/rekap.rb`

2. **After Each Release:**
   - Wait for GitHub Actions to complete
   - Verify formula was updated in homebrew-rekap repository
   - Test installation on at least one clean system

3. **Periodic Testing:**
   - Test on newly released macOS versions
   - Test after major Homebrew updates
   - Test when dependencies (Go, libraries) are updated

## Resources

- [Homebrew Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Troubleshooting](https://docs.brew.sh/Troubleshooting)
- [Homebrew Common Issues](https://docs.brew.sh/Common-Issues)
