# Quick Start: Setting Up Homebrew Tap

This guide provides step-by-step instructions for the repository owner (@alexinslc) to set up the Homebrew tap.

## Step 1: Create the homebrew-rekap Repository

1. Go to https://github.com/new
2. Fill in the details:
   - **Owner:** alexinslc
   - **Repository name:** `homebrew-rekap` (must start with "homebrew-")
   - **Description:** "Homebrew tap for rekap - Daily Mac activity summary CLI"
   - **Visibility:** Public ✓
   - **Initialize with README:** ✓
3. Click "Create repository"
4. Replace the auto-generated README with the content from `docs/HOMEBREW_REPO_README.md`

## Step 2: Create Personal Access Token

1. Go to https://github.com/settings/tokens
2. Click "Generate new token" → "Generate new token (classic)"
3. Configure the token:
   - **Note:** "GoReleaser Homebrew Tap for rekap"
   - **Expiration:** Set to "No expiration" or a long duration (1 year)
   - **Scopes:** Select these checkboxes:
     - ✓ `repo` (Full control of private repositories)
       - This includes: repo:status, repo_deployment, public_repo, repo:invite, security_events
     - ✓ `workflow` (Update GitHub Action workflows)
4. Click "Generate token"
5. **IMPORTANT:** Copy the token immediately (starts with `ghp_`)
   - You won't be able to see it again!
   - Store it securely (password manager, secure note)

## Step 3: Add Token to rekap Repository Secrets

1. Go to https://github.com/alexinslc/rekap/settings/secrets/actions
2. Click "New repository secret"
3. Fill in:
   - **Name:** `HOMEBREW_TAP_TOKEN`
   - **Secret:** Paste the token from Step 2 (the one starting with `ghp_`)
4. Click "Add secret"

## Step 4: Verify the Setup

The setup is complete! The GoReleaser configuration and GitHub Actions workflow are already in place.

To verify everything is ready:

1. Check `.goreleaser.yml` has the `brews` section ✓ (already present)
2. Check `.github/workflows/release.yml` uses `HOMEBREW_TAP_TOKEN` ✓ (already updated)
3. The `homebrew-rekap` repository exists ✓ (you just created it)
4. The `HOMEBREW_TAP_TOKEN` secret is set ✓ (you just added it)

## Step 5: Test with a Release

When you're ready to test:

1. Make sure all your changes are committed and pushed
2. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. Watch the GitHub Actions workflow:
   - Go to https://github.com/alexinslc/rekap/actions
   - The "Release" workflow should start automatically
   - Wait for it to complete (usually 2-5 minutes)
4. Verify the release:
   - Check https://github.com/alexinslc/rekap/releases
   - Should see a new release with binaries
5. Verify the formula:
   - Check https://github.com/alexinslc/homebrew-rekap/blob/main/Formula/rekap.rb
   - Should see the formula file created by GoReleaser
6. Test the installation:
   ```bash
   brew tap alexinslc/rekap
   brew install rekap
   rekap --version
   rekap demo
   ```

## Step 6: Update Documentation (Optional)

The README already shows Homebrew as the recommended installation method. After your first successful release, you may want to:

1. Announce the Homebrew availability in your project's README/website
2. Update any documentation that mentioned "coming soon"
3. Share the news with users (release notes, social media, etc.)

## Troubleshooting

### Release Workflow Fails

**Error:** "Could not push to homebrew-rekap repository"

**Solution:**
- Verify the `HOMEBREW_TAP_TOKEN` secret is set correctly
- Ensure the token has `repo` and `workflow` scopes
- Check if the token has expired
- Regenerate the token if needed and update the secret

### Formula Not Created

**Error:** No formula appears in homebrew-rekap after release

**Solution:**
- Check the workflow logs: https://github.com/alexinslc/rekap/actions
- Look for errors in the GoReleaser step
- Verify the `.goreleaser.yml` configuration is correct
- Ensure the `homebrew-rekap` repository is public

### Can't Install via Homebrew

**Error:** `Error: No available formula with the name "alexinslc/rekap/rekap"`

**Solution:**
- Ensure you've completed at least one release (pushed a tag)
- Wait a few minutes after the release for GitHub to sync
- Try: `brew update` then `brew tap alexinslc/rekap` then `brew install rekap`
- Verify the formula exists: https://github.com/alexinslc/homebrew-rekap/blob/main/Formula/rekap.rb

## What Happens on Each Release

When you push a new tag (e.g., `v0.2.0`):

1. GitHub Actions detects the tag push
2. Workflow starts on macOS runner
3. GoReleaser builds binaries (arm64 and amd64)
4. GoReleaser creates GitHub release with assets
5. GoReleaser generates/updates the Homebrew formula
6. GoReleaser pushes the formula to homebrew-rekap repository
7. Users can immediately install the new version via Homebrew

## Need Help?

- See `docs/HOMEBREW_TAP_SETUP.md` for detailed information
- See `docs/HOMEBREW_TESTING.md` for testing procedures
- Open an issue if you encounter problems

## Resources

- [GoReleaser Documentation](https://goreleaser.com/)
- [Homebrew Tap Documentation](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
