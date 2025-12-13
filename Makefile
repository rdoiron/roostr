.PHONY: dev api ui build build-api build-ui test test-api test-ui lint clean help

# Default target
help:
	@echo "Roostr Development Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev        - Run API and UI dev servers"
	@echo "  make api        - Run API server only"
	@echo "  make ui         - Run UI dev server only"
	@echo ""
	@echo "Building:"
	@echo "  make build      - Build everything"
	@echo "  make build-api  - Build Go binary"
	@echo "  make build-ui   - Build Svelte app"
	@echo ""
	@echo "Testing:"
	@echo "  make test       - Run all tests"
	@echo "  make test-api   - Run Go tests"
	@echo "  make test-ui    - Run Svelte tests"
	@echo ""
	@echo "Other:"
	@echo "  make lint       - Lint all code"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make db-reset   - Reset app database"

# Development
dev:
	@echo "Starting development servers..."
	@make -j2 api ui

api:
	cd app/api && go run ./cmd/...

ui:
	cd app/ui && npm run dev

# Building
build: build-api build-ui
	@echo "Build complete!"

build-api:
	cd app/api && go build -o ../../dist/roostr-api ./cmd/...

build-ui:
	cd app/ui && npm run build
	cp -r app/ui/build dist/ui

# Testing
test: test-api test-ui

test-api:
	cd app/api && go test ./...

test-ui:
	cd app/ui && npm run test

# Linting
lint:
	cd app/api && golangci-lint run
	cd app/ui && npm run lint

# Database
db-reset:
	rm -f data/roostr.db
	@echo "Database reset. Will be recreated on next API start."

# Cleaning
clean:
	rm -rf dist/
	rm -rf app/ui/.svelte-kit
	rm -rf app/ui/build
	@echo "Cleaned build artifacts."

# Platform packaging (TODO)
package-umbrel:
	@echo "Umbrel packaging not yet implemented"

package-startos:
	@echo "Start9 packaging not yet implemented"
