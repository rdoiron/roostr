# Building the Start9 Package

This document explains how to build the Roostr .s9pk package for StartOS.

## Prerequisites

### 1. Install Rust

If you don't have Rust installed:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env
```

### 2. Install Start9 SDK

```bash
cargo install --git https://github.com/Start9Labs/start-os.git start-sdk
```

This may take several minutes as it compiles the SDK from source.

### 3. Install Docker

The Start9 SDK uses Docker to build the package. Ensure Docker is installed and running:

```bash
# Ubuntu/Debian
sudo apt install docker.io
sudo usermod -aG docker $USER
# Log out and back in for group changes to take effect

# Verify Docker is running
docker info
```

## Building the Package

### Using Make

From the project root:

```bash
make package-startos
```

### Manual Build

From the `platforms/startos` directory:

```bash
cd platforms/startos
start-sdk pack
```

This will:
1. Build the Docker image using the Dockerfile
2. Package all assets (manifest, instructions, icon, scripts)
3. Create a `.s9pk` file

## Output

The build produces a file named `roostr.s9pk` which can be:
- Sideloaded onto a StartOS server
- Submitted to the Start9 Marketplace

## Sideloading for Testing

1. Open your StartOS dashboard
2. Go to System > Sideload
3. Upload the `.s9pk` file
4. Follow the installation prompts

## Package Contents

The .s9pk includes:
- `manifest.yaml` - Package metadata and configuration
- `Dockerfile` - Container build instructions
- `instructions.md` - User documentation
- `icon.png` - App icon (512x512)
- `scripts/` - Health check, backup, and restore scripts
- `LICENSE` - MIT license

## Troubleshooting

### "start-sdk not found"

Ensure Cargo bin is in your PATH:
```bash
export PATH="$HOME/.cargo/bin:$PATH"
```

### Docker permission denied

Add your user to the docker group:
```bash
sudo usermod -aG docker $USER
```
Then log out and back in.

### Build fails during Rust compilation

The nostr-rs-relay build requires significant memory. Ensure you have at least 4GB RAM available.

## References

- [Start9 Developer Docs](https://docs.start9.com/)
- [StartOS Package Specification](https://github.com/Start9Labs/start-os)
