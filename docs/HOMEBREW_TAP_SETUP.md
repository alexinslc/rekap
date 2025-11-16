# Homebrew Tap Setup Guide

This document explains how to set up and maintain the Homebrew tap for rekap.

## Overview

rekap uses [GoReleaser](https://goreleaser.com/) to automatically publish to a Homebrew tap when a new version is released. The configuration is already in place in `.goreleaser.yml`, but there are some one-time setup steps required.

## Prerequisites

1. **Create the homebrew-rekap repository**
   - Go to https://github.com/alexinslc
   - Create a new public repository named `homebrew-rekap`
   - Initialize it with a README
   - No need to create any formula files manually - GoReleaser will do this automatically

2. **Create a GitHub Personal Access Token (PAT)**
   
   The default `GITHUB_TOKEN` provided by GitHub Actions only has permissions for the current repository. To push to the `homebrew-rekap` repository, we need a Personal Access Token.
   
   Steps:
   - Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
   - Click "Generate new token (classic)"
   - Name it something like "GoReleaser Homebrew Tap"
   - Select the following scopes:
     - `repo` (Full control of private repositories)
     - `workflow` (Update GitHub Action workflows)
   - Generate the token and copy it (you won't be able to see it again)

3. **Add the token as a repository secret**
   - Go to the rekap repository: https://github.com/alexinslc/rekap
   - Navigate to Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_TOKEN`
   - Value: Paste the PAT you created
   - Click "Add secret"

## Configuration

The `.goreleaser.yml` file already contains the necessary Homebrew tap configuration:

```yaml
brews:
  - name: rekap
    ids:
      - rekap-archive
    repository:
      owner: alexinslc
      name: homebrew-rekap
    homepage: https://github.com/alexinslc/rekap
    description: "Daily Mac activity summary in your terminal"
    license: MIT
    directory: Formula
    install: |
      bin.install "rekap"
    test: |
      system "#{bin}/rekap --version"
```

The release workflow (`.github/workflows/release.yml`) has been updated to use the `HOMEBREW_TAP_TOKEN` secret instead of the default `GITHUB_TOKEN`.

## How It Works

When you push a new tag (e.g., `v1.0.0`):

1. GitHub Actions triggers the release workflow
2. GoReleaser builds the binaries for macOS (arm64 and amd64)
3. GoReleaser creates a GitHub release with the binaries
4. GoReleaser creates/updates the Homebrew formula in the `homebrew-rekap` repository
5. Users can install with: `brew install alexinslc/rekap/rekap`

## Formula Location

The Homebrew formula will be automatically created at:
```
https://github.com/alexinslc/homebrew-rekap/blob/main/Formula/rekap.rb
```

## User Installation

After the first release, users can install rekap with:

```bash
# Using the tap
brew tap alexinslc/rekap
brew install rekap

# Or directly
brew install alexinslc/rekap/rekap
```

## Testing the Setup

### Before the First Release

You can test the GoReleaser configuration without creating a release:

```bash
# Install GoReleaser (if not already installed)
brew install goreleaser

# Test the configuration (dry run)
goreleaser release --snapshot --clean --skip=publish

# This will:
# - Build the binaries
# - Create archives
# - Generate the Homebrew formula locally
# - NOT publish anything to GitHub
```

The generated formula will be in `dist/homebrew/Formula/rekap.rb`.

### After the First Release

Once you've created your first release (by pushing a tag like `v0.1.0`):

1. Test on a clean macOS system or in a fresh terminal:
   ```bash
   brew tap alexinslc/rekap
   brew install rekap
   rekap --version
   rekap demo
   ```

2. Test updating:
   ```bash
   # After releasing a new version
   brew update
   brew upgrade rekap
   rekap --version
   ```

3. Test uninstall:
   ```bash
   brew uninstall rekap
   brew untap alexinslc/rekap
   ```

## Troubleshooting

### Formula Not Found After Release

- Wait a few minutes - GitHub Actions takes time to complete
- Check the release workflow: https://github.com/alexinslc/rekap/actions
- Verify the formula was created: https://github.com/alexinslc/homebrew-rekap/blob/main/Formula/rekap.rb

### Permission Errors During Release

- Verify the `HOMEBREW_TAP_TOKEN` secret is set correctly
- Ensure the PAT has the `repo` and `workflow` scopes
- Check that the token hasn't expired

### Formula Updates Not Working

- GoReleaser updates the formula on each release
- If you need to manually fix the formula, you can edit it directly in the homebrew-rekap repository
- Be aware that GoReleaser will overwrite manual changes on the next release

## Updating the Formula

The formula is automatically updated by GoReleaser on each release. However, if you need to customize it:

1. Edit `.goreleaser.yml` in the rekap repository
2. Update the `brews` section with your changes
3. Test with `goreleaser release --snapshot --clean --skip=publish`
4. Commit and push the changes
5. Create a new release to apply the changes

## Resources

- [GoReleaser Homebrew Documentation](https://goreleaser.com/customization/homebrew/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
