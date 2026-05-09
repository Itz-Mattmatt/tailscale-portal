# Releasing

This document covers the CI/CD pipeline, packaging configuration, and how to create releases.

## CI/CD Pipeline

The project uses **GitHub Actions** and **GoReleaser** to automatically build and publish packages when a version tag is pushed.

### Workflow File

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### What the Pipeline Does

When you push a tag matching `v*`:

1. **Checkout** — Full git history (needed for changelog)
2. **Set up Go** — Uses the latest stable Go version
3. **Run tests** — All tests must pass before building
4. **GoReleaser** — Builds binaries, creates packages, publishes release

## GoReleaser Configuration

The `.goreleaser.yaml` file defines the entire release process.

### Builds

Targets multiple platforms:

| OS | Architectures |
|----|--------------|
| Linux | amd64, arm64, 386 |
| macOS | amd64, arm64 |
| Windows | amd64, 386 |

Build flags:
- `CGO_ENABLED=0` — Static binary, no C dependencies
- `-s -w` — Strip debug info for smaller binaries
- `-X main.version={{.Version}}` — Embed version in binary

### Archives

Each platform gets an archive:
- Linux/macOS: `.tar.gz`
- Windows: `.zip`

Archive naming: `tailscale-portal_{version}_{os}_{arch}.{ext}`

Includes `README.md` and `LICENSE`.

### Packages

System packages are generated using `nfpms`:

| Format | Distros |
|--------|---------|
| `.deb` | Debian, Ubuntu, Mint |
| `.rpm` | Fedora, RHEL, CentOS, openSUSE |
| `.apk` | Alpine Linux |

Packages install the binary to `/usr/local/bin/tailscale-portal` and include documentation.

### Homebrew Tap

A Homebrew formula is automatically generated and pushed to a tap repository:

```bash
brew tap YOUR_GITHUB_USERNAME/tap
brew install tailscale-portal
```

**Requirements:**
- Create a public GitHub repo named `homebrew-tap`
- GoReleaser will push formula updates to this repo on each release

### Checksums

A `checksums.txt` file is generated with SHA-256 hashes of all artifacts for verification.

### Changelog

GoReleaser generates a changelog from commit messages, excluding:
- `docs:` commits
- `test:` commits
- `chore:` commits
- `ci:` commits
- Merge commits

## Creating a Release

### Before Your First Release

1. Update placeholders in `.goreleaser.yaml`:
   - Replace `YOUR_GITHUB_USERNAME` with your GitHub username (3 occurrences)
   - Update `maintainer` name and email

2. Update placeholders in `README.md`:
   - Replace `YOUR_GITHUB_USERNAME` with your GitHub username

3. Create the Homebrew tap repository (optional, if you want Homebrew support):
   ```bash
   # Create on GitHub: https://github.com/new
   # Name: homebrew-tap
   # Visibility: Public
   ```

### Release Steps

1. Make sure all changes are committed and pushed
2. Create an annotated tag:
   ```bash
   git tag -a v0.1.0 -m "First release"
   ```
3. Push the tag:
   ```bash
   git push origin v0.1.0
   ```
4. GitHub Actions will automatically run and create the release

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- `v1.0.0` — Major release (breaking changes)
- `v1.1.0` — Minor release (new features)
- `v1.1.1` — Patch release (bug fixes)

Pre-releases (alpha, beta, rc):

```bash
git tag -a v1.0.0-beta.1 -m "Beta 1"
```

GoReleaser will mark these as pre-releases on GitHub.

## Manual Release

If you need to release locally (not recommended for production):

### Prerequisites

```bash
# Install GoReleaser
brew install goreleaser

# Or on Linux
curl -sfL https://goreleaser.com/static/run | bash
```

### Release Command

```bash
# Create a tag
git tag -a v0.1.0 -m "First release"

# Set GitHub token (needed to create the release)
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# Run GoReleaser
goreleaser release --clean
```

### Snapshot Release (No GitHub)

Build without publishing:

```bash
goreleaser release --snapshot --clean
```

Artifacts will be in the `dist/` directory.

## Troubleshooting Releases

### "GitHub token not found"

The `GITHUB_TOKEN` secret is automatically provided by GitHub Actions. For manual releases, create a personal access token with `repo` scope.

### "Homebrew tap not found"

Ensure the `homebrew-tap` repository exists and is public. GoReleaser needs push access.

### "Tests fail during release"

The pipeline runs tests before building. Fix failing tests before tagging:

```bash
go test ./...
```

### "GoReleaser version mismatch"

The workflow pins to `~> v2`. If GoReleaser v3 is released, update the workflow:

```yaml
version: '~> v3'
```

## Release Checklist

- [ ] All tests passing locally
- [ ] Version placeholders updated in `.goreleaser.yaml`
- [ ] Version placeholders updated in `README.md`
- [ ] `CHANGELOG.md` or release notes prepared (optional)
- [ ] Homebrew tap repo created (if using Homebrew)
- [ ] Git tag created and pushed
- [ ] GitHub Actions workflow completed successfully
- [ ] Release page looks correct on GitHub
- [ ] Download links work
- [ ] Homebrew formula updated (if applicable)
