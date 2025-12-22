# StartOS Package Makefile for Roostr
# Following Start9Labs conventions

PKG_ID := $(shell yq e ".id" manifest.yaml)
PKG_VERSION := $(shell yq e ".version" manifest.yaml)

.DELETE_ON_ERROR:

.PHONY: all arm x86 pack clean verify install help

# Default target: build both architectures and pack
all: pack

# Build ARM64 Docker image (for Raspberry Pi / ARM servers)
arm: docker-images/aarch64.tar

# Build x86_64 Docker image (for Intel/AMD servers)
x86: docker-images/x86_64.tar

docker-images/aarch64.tar:
	@echo "Building ARM64 image (this will take a while)..."
	mkdir -p docker-images
	docker buildx build \
		--platform linux/arm64 \
		--tag start9/$(PKG_ID)/main:$(PKG_VERSION) \
		-o type=docker,dest=docker-images/aarch64.tar \
		-f Dockerfile \
		../..

docker-images/x86_64.tar:
	@echo "Building x86_64 image..."
	mkdir -p docker-images
	docker buildx build \
		--platform linux/amd64 \
		--tag start9/$(PKG_ID)/main:$(PKG_VERSION) \
		-o type=docker,dest=docker-images/x86_64.tar \
		-f Dockerfile \
		../..

# Pack the .s9pk using Docker-based SDK (no local SDK install required)
pack: docker-images/aarch64.tar docker-images/x86_64.tar
	@echo "Packing $(PKG_ID).s9pk using Docker-based SDK..."
	docker run --rm \
		-v "$$(cd ../.. && pwd):/workspace" \
		-w /workspace/platforms/startos \
		rust:1.82-bookworm bash -c "\
			apt-get update -qq && apt-get install -y -qq git >/dev/null 2>&1 && \
			git config --global --add safe.directory /workspace && \
			git clone --depth 1 --branch sdk --recurse-submodules \
				https://github.com/Start9Labs/start-os.git /tmp/sdk 2>/dev/null && \
			cd /tmp/sdk && echo 'sdk' > GIT_HASH.txt && mkdir -p web/dist/static && \
			cd core && cargo install --path=./startos --no-default-features \
				--features=sdk --locked --root /usr/local 2>&1 | tail -3 && \
			ln -sf /usr/local/bin/startbox /usr/local/bin/start-sdk && \
			cd /workspace/platforms/startos && \
			start-sdk init && start-sdk pack \
		"
	@echo "Package created: $(PKG_ID).s9pk"
	@sha256sum $(PKG_ID).s9pk | tee $(PKG_ID).s9pk.sha256

# Verify the package (requires local start-sdk or use Docker)
verify: $(PKG_ID).s9pk
	@echo "Verifying $(PKG_ID).s9pk..."
	@if command -v start-sdk >/dev/null 2>&1; then \
		start-sdk verify s9pk $(PKG_ID).s9pk; \
	else \
		echo "Skipping verification (start-sdk not installed locally)"; \
		echo "Package contents:"; \
		tar -tvf $(PKG_ID).s9pk | head -15; \
	fi

# Install to a StartOS instance (requires ~/.embassy/config.yaml)
install: $(PKG_ID).s9pk
	@echo "Installing $(PKG_ID).s9pk..."
	start-cli package install $(PKG_ID).s9pk

# Clean build artifacts
clean:
	rm -rf docker-images *.s9pk

# Help
help:
	@echo "Roostr StartOS Package Build"
	@echo ""
	@echo "Targets:"
	@echo "  make        - Build both architectures and pack s9pk"
	@echo "  make arm    - Build ARM64 image only"
	@echo "  make x86    - Build x86_64 image only"
	@echo "  make pack   - Create .s9pk (uses Docker-based SDK)"
	@echo "  make verify - Verify the .s9pk package"
	@echo "  make install - Install to StartOS (requires start-cli auth)"
	@echo "  make clean  - Remove build artifacts"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Docker with buildx"
	@echo "  - yq (YAML processor)"
	@echo ""
	@echo "Note: SDK is built automatically in Docker - no local install needed."
	@echo ""
	@echo "For ARM cross-compilation, enable QEMU first:"
	@echo "  docker run --privileged --rm linuxkit/binfmt:v0.8"
