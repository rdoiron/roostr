# Building the StartOS Package

This document explains how to build the Roostr `.s9pk` package for StartOS.

## StartOS Version Compatibility

| StartOS Version | Package Format | Manifest | Notes |
|-----------------|----------------|----------|-------|
| v0.3.5.x (stable) | s9pk v1 | YAML | Current target |
| v0.4.x (alpha) | s9pk v2 | JavaScript | Requires `start-cli s9pk convert` |

This guide targets **StartOS v0.3.5.x** (the current stable release).

## Prerequisites

Install the following tools:

1. **Docker with Buildx**
   ```bash
   # Verify buildx is available
   docker buildx version
   ```

2. **yq** (YAML processor)
   ```bash
   # Install via snap, brew, or your package manager
   sudo snap install yq
   ```

3. **QEMU for cross-compilation** (if building ARM on x86 or vice versa)
   ```bash
   docker run --privileged --rm linuxkit/binfmt:v0.8
   ```

> **Note**: The StartOS SDK is built automatically in Docker during the pack step. You don't need to install it locally.

## Building the Package

From the `platforms/startos` directory:

```bash
cd platforms/startos

# Build for both architectures and create .s9pk
make

# Or build a single architecture
make x86    # For Intel/AMD servers
make arm    # For Raspberry Pi / ARM servers
```

The build compiles:
- Svelte UI (Node.js)
- Go API
- nostr-rs-relay (Rust) - this takes the longest

**Expected build times:**
- x86_64 on x86_64: ~10-15 minutes
- ARM64 on x86_64 (QEMU): ~60-90 minutes
- ARM64 on ARM64 (native): ~20-30 minutes

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make` | Build both architectures and verify package |
| `make arm` | Build ARM64 image only |
| `make x86` | Build x86_64 image only |
| `make verify` | Verify the .s9pk package |
| `make install` | Install to StartOS (requires start-cli auth) |
| `make clean` | Remove build artifacts |
| `make help` | Show available targets |

## Installing on StartOS

### Option 1: Sideload via Web UI

1. Open your StartOS dashboard
2. Go to **System > Sideload Service**
3. Upload the `roostr.s9pk` file
4. Follow the installation prompts

### Option 2: Install via CLI

```bash
# Configure your StartOS instance
echo "host: https://your-server.local" > ~/.embassy/config.yaml

# Authenticate
start-cli auth login

# Install
make install
```

## Uploading to GitHub Release

After building, upload to an existing GitHub release:

```bash
# Upload to release (requires gh CLI)
gh release upload v0.1.0 roostr.s9pk

# Or with checksum
sha256sum roostr.s9pk > roostr.s9pk.sha256
gh release upload v0.1.0 roostr.s9pk roostr.s9pk.sha256
```

## Why Local Builds?

StartOS packages require Docker images for both x86_64 and ARM64 architectures. Cross-compiling Rust (nostr-rs-relay) via QEMU emulation takes 40-60x longer than native compilation. GitHub Actions has a 6-hour job limit, which is insufficient for ARM64 cross-compilation.

This matches the approach used by Start9Labs and the official [nostr-rs-relay-startos](https://github.com/Start9Labs/nostr-rs-relay-startos) wrapper.

## Package Contents

The `.s9pk` includes:
- `manifest.yaml` - Package metadata and configuration
- `docker-images/` - Container images for both architectures
- `instructions.md` - User documentation
- `icon.png` - App icon (512x512)
- `scripts/` - Health check, backup, and restore scripts
- `LICENSE` - MIT license

## Troubleshooting

### QEMU errors during ARM build
```bash
# Re-register QEMU handlers
docker run --privileged --rm linuxkit/binfmt:v0.8
```

### start-sdk not found or won't compile

The SDK is difficult to compile locally due to Rust version requirements and git submodules. Use the Docker-based build instead:

```bash
# Build SDK in Docker and pack s9pk
docker run --rm \
  -v "$(pwd)/../..:/workspace" \
  -w /workspace/platforms/startos \
  rust:1.82-bookworm bash -c "
    apt-get update -qq && apt-get install -y -qq git yq >/dev/null 2>&1
    git config --global --add safe.directory /workspace
    git clone --depth 1 --branch sdk --recurse-submodules \
      https://github.com/Start9Labs/start-os.git /tmp/sdk
    cd /tmp/sdk && echo 'sdk' > GIT_HASH.txt && mkdir -p web/dist/static
    cd core && cargo install --path=./startos --no-default-features \
      --features=sdk --locked --root /usr/local 2>&1 | tail -5
    ln -sf /usr/local/bin/startbox /usr/local/bin/start-sdk
    cd /workspace/platforms/startos
    start-sdk init && start-sdk pack
  "
```

### Build fails with "no space left on device"
Docker images are large. Clean up old images:
```bash
docker system prune -a
```

### Service crashes with "Permission denied" on /data

StartOS mounts the `/data` volume with root ownership. The container must run as root, not a non-root user. Ensure the Dockerfile does NOT have a `USER` directive.

```dockerfile
# WRONG - will cause permission errors
USER appuser
ENTRYPOINT ["/entrypoint.sh"]

# CORRECT - run as root for StartOS
ENTRYPOINT ["/entrypoint.sh"]
```

### "Invalid Package File" when sideloading

The s9pk format requires proper SDK signing/headers. You cannot manually create a tar archive - you must use `start-sdk pack`. Use the Docker-based build command above.

## Converting to v2 for StartOS 0.4.x (Alpha)

If testing on StartOS v0.4.x alpha, convert the v1 package to v2:

```bash
# Requires start-cli v0.4 and squashfs-tools-ng (for tar2sqfs)
start-cli s9pk convert roostr.s9pk
```

This creates a v2 package with SquashFS format.

## References

- [Start9 Developer Docs](https://docs.start9.com/0.3.5.x/developer-docs/packaging)
- [StartOS Package Specification](https://github.com/Start9Labs/start-os)
- [nostr-rs-relay-startos](https://github.com/Start9Labs/nostr-rs-relay-startos) - Reference implementation
