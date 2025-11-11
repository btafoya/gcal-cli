# Release Process

This document describes how to create and publish releases for gcal-cli.

## Automated Releases (Recommended)

Releases are automatically built and published via GitHub Actions when you push a version tag.

### Steps

1. **Ensure all changes are committed and pushed**
   ```bash
   git status  # Should be clean
   git push origin master
   ```

2. **Create a version tag**
   ```bash
   # For a new release version (e.g., 1.0.0)
   ./scripts/version.sh tag 1.0.0

   # This creates a git tag: v1.0.0
   ```

3. **Push the tag to trigger the release**
   ```bash
   git push origin v1.0.0
   ```

4. **Monitor the build**
   - Go to: https://github.com/btafoya/gcal-cli/actions
   - Watch the "Build and Release" workflow
   - Build creates binaries for:
     - Linux (amd64, arm64)
     - macOS (amd64, arm64)
     - Windows (amd64)

5. **Release is published**
   - GitHub automatically creates a release
   - Binaries are attached as downloadable assets
   - Release notes are auto-generated from commits
   - Docker image is published to GitHub Container Registry

## Manual Local Builds

For testing or local distribution, you can build all platform binaries locally:

```bash
# Build all platform binaries
./scripts/build-release.sh

# Output in dist/ directory:
# - gcal-cli-{version}-linux-amd64.tar.gz
# - gcal-cli-{version}-linux-arm64.tar.gz
# - gcal-cli-{version}-darwin-amd64.tar.gz
# - gcal-cli-{version}-darwin-arm64.tar.gz
# - gcal-cli-{version}-windows-amd64.zip
# - checksums.txt
```

## Versioning Scheme

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backwards-compatible)
- **PATCH** version: Bug fixes (backwards-compatible)

### Version Commands

```bash
# Show current version
./scripts/version.sh current
# Output: 1.0.0-dev.5+a1b2c3d (development)
# Output: 1.0.0 (released)

# Show next patch version
./scripts/version.sh next
# Output: 1.0.1

# Increment version types
./scripts/version.sh patch   # 1.0.0 → 1.0.1
./scripts/version.sh minor   # 1.0.0 → 1.1.0
./scripts/version.sh major   # 1.0.0 → 2.0.0
```

## Pre-Release Checklist

Before creating a release tag:

- [ ] All tests passing: `go test ./...`
- [ ] Code builds successfully: `go build ./cmd/gcal-cli`
- [ ] Documentation updated (README.md, USER-INSTRUCTIONS.md)
- [ ] CHANGELOG or release notes prepared
- [ ] No uncommitted changes: `git status`
- [ ] On master branch: `git branch`

## Release Types

### Patch Release (1.0.X)

For bug fixes and minor improvements:

```bash
./scripts/version.sh tag 1.0.1
git push origin v1.0.1
```

### Minor Release (1.X.0)

For new features (backwards-compatible):

```bash
./scripts/version.sh tag 1.1.0
git push origin v1.1.0
```

### Major Release (X.0.0)

For breaking changes:

```bash
./scripts/version.sh tag 2.0.0
git push origin v2.0.0
```

## Docker Images

Docker images are automatically built and published when you push a version tag.

**Available at**: `ghcr.io/btafoya/gcal-cli`

**Tags**:
- `latest` - Latest stable release
- `{version}` - Specific version (e.g., `1.0.0`)

**Pull and run**:
```bash
docker pull ghcr.io/btafoya/gcal-cli:latest
docker run ghcr.io/btafoya/gcal-cli:latest version
```

## GitHub Actions Workflow

The build workflow (`.github/workflows/build.yml`) performs:

1. **Build Job** (runs on push and tag):
   - Builds on Linux, macOS, and Windows
   - Runs all tests
   - Creates versioned binaries
   - Uploads artifacts

2. **Release Job** (runs on tag only):
   - Downloads all build artifacts
   - Creates GitHub release
   - Attaches binary archives
   - Generates release notes

3. **Docker Job** (runs on tag only):
   - Builds multi-arch Docker images
   - Publishes to GitHub Container Registry
   - Tags with version and `latest`

## Troubleshooting

### Tag Already Exists

If you need to move a tag:
```bash
git tag -d v1.0.0              # Delete local tag
git push origin :refs/tags/v1.0.0  # Delete remote tag
./scripts/version.sh tag 1.0.0     # Recreate tag
git push origin v1.0.0             # Push new tag
```

### Build Fails

Check the GitHub Actions logs:
1. Go to: https://github.com/btafoya/gcal-cli/actions
2. Click on the failed workflow
3. Review the logs for each job

Common issues:
- Test failures → Fix tests before releasing
- Build errors → Verify code compiles locally
- Permission errors → Check GitHub token permissions

### Release Not Created

Verify:
- Tag starts with `v` (e.g., `v1.0.0`, not `1.0.0`)
- GitHub Actions has permission to create releases
- Build job completed successfully

## Best Practices

1. **Always test before tagging**
   ```bash
   go test ./...
   go build ./cmd/gcal-cli
   ./gcal-cli version
   ```

2. **Review changes since last release**
   ```bash
   git log v1.0.0..HEAD --oneline
   ```

3. **Keep releases focused**
   - One logical change set per release
   - Group related bug fixes
   - Document breaking changes clearly

4. **Use meaningful version numbers**
   - Don't skip versions
   - Follow semantic versioning strictly
   - Document version changes

5. **Monitor the build**
   - Watch GitHub Actions until release is published
   - Test download links work
   - Verify release notes are accurate

## Support

For issues with the release process:
- GitHub Actions: https://github.com/btafoya/gcal-cli/actions
- Issues: https://github.com/btafoya/gcal-cli/issues
- Documentation: https://github.com/btafoya/gcal-cli/tree/master/docs
