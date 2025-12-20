# Building the Start9 Package

This document explains how to build the Roostr `.s9pk` package for StartOS.

## Automated Builds (Recommended)

The `.s9pk` package is built automatically via GitHub Actions when you create a release.

### Creating a Release

1. Tag the release:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. GitHub Actions will:
   - Build the StartOS package using the Start9 SDK
   - Build and push the Docker image
   - Create a GitHub Release with the `.s9pk` attached

3. Download `roostr.s9pk` from the [Releases page](https://github.com/rdoiron/roostr/releases)

### Manual Trigger

You can also trigger a build manually:

1. Go to Actions > Release workflow
2. Click "Run workflow"
3. Enter the version number (e.g., `0.1.0`)
4. Download the artifact when complete

## Sideloading for Testing

1. Open your StartOS dashboard
2. Go to **System > Sideload Service**
3. Upload the `.s9pk` file
4. Follow the installation prompts

## Package Contents

The `.s9pk` includes:
- `manifest.yaml` - Package metadata and configuration
- `Dockerfile` - Container build instructions
- `instructions.md` - User documentation
- `icon.png` - App icon (512x512)
- `scripts/` - Health check, backup, and restore scripts
- `LICENSE` - MIT license

## Local Development (Advanced)

For local SDK installation, see the [Start9 Developer Docs](https://docs.start9.com/).

The basic process:
```bash
# Clone and build the SDK
git clone --recursive https://github.com/Start9Labs/start-os.git --branch sdk
cd start-os && make sdk

# Build the package
cd /path/to/roostr/platforms/startos
start-sdk pack
```

Note: Local SDK installation requires specific Rust versions and system dependencies.
The GitHub Actions approach is recommended to avoid compatibility issues.

## CI/CD Secrets Required

For automated releases, configure these repository secrets:

| Secret | Description |
|--------|-------------|
| `DOCKERHUB_USERNAME` | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token |

## References

- [Start9 Developer Docs](https://docs.start9.com/)
- [Start9 SDK GitHub Action](https://github.com/Start9Labs/sdk)
- [StartOS Package Specification](https://github.com/Start9Labs/start-os)
