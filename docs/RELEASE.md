# Release Process

This document describes the release process for s3-backup, including versioning, tagging, and Docker image publishing.

## Overview

This project uses semantic versioning and automated releases via GitHub Actions. Docker images are **only published on official releases**, not on every commit to main. This keeps the container registry clean and makes it clear which versions are production-ready.

## Release Workflow

### 1. Prepare for Release

Before creating a release:

1. **Ensure all tests pass**

   ```bash
   go test -v ./...
   ```

2. **Update documentation** if needed

   - Update README.md with new features
   - Update examples if API changed

3. **Commit all changes** to your feature branch

   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin your-branch
   ```

4. **Create and merge PR** to `main`
   - All tests must pass in CI
   - Get required reviews
   - Merge to main

### 2. Create a Release

Once changes are merged to `main`:

#### Option A: GitHub UI (Recommended)

1. Go to: https://github.com/RyanDerr/s3-backup/releases/new

2. Click **"Choose a tag"** and create a new tag:

   - Format: `v1.2.3` (must start with `v`)
   - Follow [semantic versioning](https://semver.org/):
     - `MAJOR`: Breaking changes (v1.x.x → v2.0.0)
     - `MINOR`: New features, backwards compatible (v1.1.x → v1.2.0)
     - `PATCH`: Bug fixes, backwards compatible (v1.1.1 → v1.1.2)

3. Set the release title: Same as tag (e.g., `v1.2.3`)

4. Click **"Generate release notes"** for automatic changelog

5. Optionally add custom release notes or highlights

6. Click **"Publish release"**

#### Option B: Command Line

```bash
# Make sure you're on main and up to date
git checkout main
git pull origin main

# Create and push the tag
git tag -a v1.2.3 -m "Release v1.2.3: Brief description"
git push origin v1.2.3

# Then create the release on GitHub UI or use gh CLI:
gh release create v1.2.3 --generate-notes
```

### 3. Automated Release Process

Once you push a tag starting with `v*.*.*`, GitHub Actions automatically:

1. **Runs all tests** to ensure code quality
2. **Builds binaries** for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
3. **Generates SHA256 checksums** for each binary
4. **Creates a GitHub Release** with all binaries attached
5. **Builds Docker images** for multiple architectures (amd64, arm64)
6. **Pushes images** to GitHub Container Registry with tags:
   - `ghcr.io/ryanderr/s3-backup:1.2.3` (exact version)
   - `ghcr.io/ryanderr/s3-backup:1.2` (minor version)
   - `ghcr.io/ryanderr/s3-backup:1` (major version)
   - `ghcr.io/ryanderr/s3-backup:latest` (latest release)

### 4. Verify the Release

After the GitHub Action completes:

1. **Check the release page**: https://github.com/RyanDerr/s3-backup/releases

   - Verify binaries are attached
   - Verify release notes are correct

2. **Verify Docker image**:

   ```bash
   docker pull ghcr.io/ryanderr/s3-backup:v1.2.3
   docker run --rm ghcr.io/ryanderr/s3-backup:v1.2.3 --help
   ```

3. **Test a binary**:
   ```bash
   # Download and test (example for Linux amd64)
   curl -LO https://github.com/RyanDerr/s3-backup/releases/download/v1.2.3/s3-backup-v1.2.3-linux-amd64
   chmod +x s3-backup-v1.2.3-linux-amd64
   ./s3-backup-v1.2.3-linux-amd64 --help
   ```

## Docker Image Usage

After a release, users can pull and run the Docker image:

```bash
# Pull specific version
docker pull ghcr.io/ryanderr/s3-backup:v1.2.3

# Pull latest
docker pull ghcr.io/ryanderr/s3-backup:latest

# Run with environment variables
docker run --rm \
  -e BACKUP_DIRS=/data \
  -e AWS_REGION=us-west-2 \
  -e S3_BUCKET=my-bucket \
  -e AWS_ACCESS_KEY_ID=xxx \
  -e AWS_SECRET_ACCESS_KEY=xxx \
  -v /path/to/backup:/data \
  ghcr.io/ryanderr/s3-backup:latest

# Run with config file
docker run --rm \
  -v /path/to/config.yaml:/config.yaml \
  -v /path/to/backup:/data \
  -e S3_BACKUP_CONFIG_FILE=/config.yaml \
  ghcr.io/ryanderr/s3-backup:latest
```

## PR Labeling for Releases

To help identify which PRs should trigger releases:

### Manual Labels

Add labels to PRs to indicate the type of change:

- `release:major` - Breaking changes (increment major version)
- `release:minor` - New features (increment minor version)
- `release:patch` - Bug fixes (increment patch version)
- `no-release` - Changes that don't require a release (docs, tests, CI)

### Conventional Commits (Recommended)

Use [conventional commits](https://www.conventionalcommits.org/) in your PR titles:

- `feat: add new feature` → Minor release
- `fix: resolve bug` → Patch release
- `feat!: breaking change` or `BREAKING CHANGE:` → Major release
- `docs: update readme` → No release
- `chore: update dependencies` → No release

### Example PR Titles

```
feat: add support for encryption at rest
fix: correct timestamp format in S3 keys
docs: improve examples documentation
chore(deps): update aws-sdk-go-v2 to v1.30.0
feat!: change configuration file format
```

## Version Numbering Guidelines

### Major Version (v1.0.0 → v2.0.0)

Increment when you make incompatible API changes:

- Changing configuration file format
- Removing command-line flags
- Changing S3 key structure
- Changing environment variable names

### Minor Version (v1.1.0 → v1.2.0)

Increment when you add functionality in a backwards-compatible manner:

- Adding new configuration options
- Adding new command-line flags
- Adding new features
- Improving existing features without breaking changes

### Patch Version (v1.1.1 → v1.1.2)

Increment when you make backwards-compatible bug fixes:

- Fixing bugs
- Security patches
- Performance improvements
- Documentation updates (though these might not need a release)

## Rollback a Release

If a release has issues:

1. **Delete the tag locally and remotely**:

   ```bash
   git tag -d v1.2.3
   git push origin :refs/tags/v1.2.3
   ```

2. **Delete the GitHub release** from the releases page

3. **Delete Docker images** (if needed):

   - Go to: https://github.com/RyanDerr/s3-backup/pkgs/container/s3-backup
   - Delete the problematic version

4. **Fix the issue** and create a new release with a patch version

## Pre-releases

For testing before official release:

```bash
# Create a pre-release tag
git tag -a v1.2.3-rc.1 -m "Release candidate 1 for v1.2.3"
git push origin v1.2.3-rc.1

# Create release on GitHub and mark as "pre-release"
gh release create v1.2.3-rc.1 --prerelease --generate-notes
```

Pre-release versions won't be tagged as `latest` in Docker.

## Troubleshooting

### Release workflow fails

1. Check the Actions tab: https://github.com/RyanDerr/s3-backup/actions
2. Look for failed jobs and error messages
3. Common issues:
   - Tests failing: Fix tests and create a new tag
   - Docker build failing: Check Dockerfile
   - Permission issues: Ensure repository settings allow workflow to write

### Docker image not appearing

1. Check workflow completed successfully
2. Verify package settings: https://github.com/RyanDerr/s3-backup/pkgs/container/s3-backup
3. Ensure package visibility is set correctly (public/private)

### Binary not attached to release

1. Check if build step completed in workflow
2. Verify the upload step didn't fail
3. Re-run the workflow if needed

## Additional Notes

- **Automated Tests**: All tests must pass before release workflow runs
- **Multi-platform Builds**: Docker images are built for both amd64 and arm64
- **Checksum Verification**: SHA256 checksums are provided for all binaries
- **Minimal Images**: Docker images use Alpine Linux (< 20MB final size)
- **Security**: Images run as non-root user
- **Caching**: Docker build uses GitHub Actions cache for faster builds
