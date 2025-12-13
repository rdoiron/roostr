# Roostr Makefile
# Private Nostr relay management app

.PHONY: all dev api ui build build-api build-ui test test-api test-ui \
        lint lint-api lint-ui db-reset db-migrate clean deps fmt \
        package-umbrel package-startos help

# Default target
all: build

# ============================================================================
# Development
# ============================================================================

# Run both API and UI dev servers
dev:
	@echo "Starting development servers..."
	@make -j2 api ui

# Run API dev server
api:
	cd app/api && go run ./cmd/server

# Run UI dev server
ui:
	cd app/ui && npm run dev

# ============================================================================
# Building
# ============================================================================

# Build everything
build: build-api build-ui
	@echo "Build complete!"

# Build Go binary
build-api:
	@mkdir -p bin
	cd app/api && go build -o ../../bin/roostr-api ./cmd/server

# Build Svelte app
build-ui:
	cd app/ui && npm run build

# ============================================================================
# Testing
# ============================================================================

# Run all tests
test: test-api test-ui

# Test Go code
test-api:
	cd app/api && go test -v ./...

# Test Svelte code
test-ui:
	cd app/ui && npm run test

# ============================================================================
# Linting
# ============================================================================

# Lint everything
lint: lint-api lint-ui

lint-api:
	cd app/api && go vet ./...
	@command -v golangci-lint >/dev/null 2>&1 && (cd app/api && golangci-lint run) || echo "golangci-lint not installed, skipping"

lint-ui:
	cd app/ui && npm run lint

# ============================================================================
# Database
# ============================================================================

# Reset app database
db-reset:
	rm -f data/roostr.db
	@echo "Database reset. Run 'make db-migrate' to recreate."

# Run migrations
db-migrate:
	@mkdir -p data
	cd app/api && go run ./cmd/migrate

# ============================================================================
# Packaging
# ============================================================================

# Build Umbrel package
package-umbrel:
	@echo "Building Umbrel package..."
	@mkdir -p dist/umbrel
	cp -r platforms/umbrel/* dist/umbrel/
	@echo "Umbrel package prepared in dist/umbrel/"

# Build Start9 package
package-startos:
	@echo "Building Start9 package..."
	@mkdir -p dist/startos
	cp -r platforms/startos/* dist/startos/
	@echo "Start9 package prepared in dist/startos/"

# ============================================================================
# Utilities
# ============================================================================

# Install dependencies
deps:
	cd app/api && go mod download
	cd app/ui && npm install

# Format code
fmt:
	cd app/api && go fmt ./...
	@cd app/ui && npm run format 2>/dev/null || true

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/
	rm -rf app/ui/.svelte-kit
	rm -rf app/ui/build
	@echo "Cleaned build artifacts."

# Show help
help:
	@echo "Roostr Development Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Run API and UI dev servers"
	@echo "  make api          - Run API server only"
	@echo "  make ui           - Run UI dev server only"
	@echo ""
	@echo "Building:"
	@echo "  make build        - Build everything"
	@echo "  make build-api    - Build Go binary"
	@echo "  make build-ui     - Build Svelte app"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run all tests"
	@echo "  make test-api     - Run Go tests"
	@echo "  make test-ui      - Run Svelte tests"
	@echo ""
	@echo "Linting:"
	@echo "  make lint         - Lint all code"
	@echo ""
	@echo "Database:"
	@echo "  make db-reset     - Reset app database"
	@echo "  make db-migrate   - Run migrations"
	@echo ""
	@echo "Packaging:"
	@echo "  make package-umbrel   - Build Umbrel package"
	@echo "  make package-startos  - Build Start9 package"
	@echo ""
	@echo "Utilities:"
	@echo "  make deps         - Install dependencies"
	@echo "  make fmt          - Format code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make help         - Show this help"
