# StartOS Package Makefile for Roostr
# Following Start9Labs conventions

PKG_ID := $(shell yq e ".id" manifest.yaml)
PKG_VERSION := $(shell yq e ".version" manifest.yaml)

.DELETE_ON_ERROR:

.PHONY: all arm x86 clean verify install

# Default target: build and verify the package
all: verify

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

# Build the .s9pk package (requires both arch images)
$(PKG_ID).s9pk: docker-images/aarch64.tar docker-images/x86_64.tar manifest.yaml instructions.md icon.png LICENSE
	@echo "Packing $(PKG_ID).s9pk..."
	start-sdk pack

# Verify the package
verify: $(PKG_ID).s9pk
	@echo "Verifying $(PKG_ID).s9pk..."
	start-sdk verify $(PKG_ID).s9pk
	@echo "Package verified successfully!"

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
	@echo "  make        - Build and verify the package (both architectures)"
	@echo "  make arm    - Build ARM64 image only"
	@echo "  make x86    - Build x86_64 image only"
	@echo "  make verify - Verify the .s9pk package"
	@echo "  make install - Install to StartOS (requires start-cli auth)"
	@echo "  make clean  - Remove build artifacts"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Docker with buildx"
	@echo "  - start-sdk (StartOS SDK)"
	@echo "  - yq (YAML processor)"
	@echo ""
	@echo "For ARM cross-compilation, enable QEMU:"
	@echo "  docker run --privileged --rm linuxkit/binfmt:v0.8"
