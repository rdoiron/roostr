# Contributing to Roostr

Thank you for your interest in contributing to Roostr! This guide will help you get started with development and explain our workflow.

## Ways to Contribute

- **Report bugs** - Open an issue describing the problem
- **Suggest features** - Open an issue with your idea
- **Fix bugs** - Submit a pull request with a fix
- **Add features** - Discuss first in an issue, then submit a PR
- **Improve docs** - Documentation improvements are always welcome

## Development Setup

### Prerequisites

Install the following before starting:

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.22+ | [golang.org/dl](https://golang.org/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) |
| npm | 10+ | Comes with Node.js |
| SQLite3 headers | - | Required for CGO |
| Git | - | [git-scm.com](https://git-scm.com/) |

**Install SQLite3 development headers:**

```bash
# Ubuntu/Debian
sudo apt install libsqlite3-dev

# macOS (via Homebrew)
brew install sqlite3

# Fedora
sudo dnf install sqlite-devel
```

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/rdoiron/roostr.git
cd roostr

# Install all dependencies
make deps
```

### Running Development Servers

```bash
# Run both API and UI dev servers
make dev

# Or run them separately:
make api   # Go API on port 3001
make ui    # Vite dev server on port 5173
```

Access the app at `http://localhost:5173`. The Vite dev server proxies `/api` requests to the Go backend.

### Running Tests

```bash
# Run all tests
make test

# Run only Go tests
make test-api

# Run only UI tests
make test-ui

# Run E2E tests (requires app running)
cd app/ui && npm run test:e2e
```

## Project Structure

```
roostr/
├── app/
│   ├── api/                 # Go backend
│   │   ├── cmd/server/      # Entry point (main.go)
│   │   └── internal/
│   │       ├── handlers/    # HTTP request handlers
│   │       ├── services/    # Business logic
│   │       ├── db/          # Database operations
│   │       ├── relay/       # Relay process management
│   │       └── config/      # Configuration
│   └── ui/                  # Svelte frontend
│       ├── src/
│       │   ├── routes/      # SvelteKit pages
│       │   └── lib/         # Components, stores, utilities
│       └── e2e/             # Playwright E2E tests
├── docs/
│   ├── SPECIFICATION.md     # Full product specification
│   ├── USER-GUIDE.md        # End-user documentation
│   ├── API.md               # API reference
│   └── TASKS.md             # Development task checklist
├── platforms/
│   ├── umbrel/              # Umbrel packaging (Docker)
│   └── startos/             # Start9 packaging (s9pk)
├── Makefile                 # Build commands
├── CLAUDE.md                # Development conventions
└── README.md                # Project overview
```

## Coding Standards

### Go Backend

- Use standard library where possible (`net/http`, `database/sql`, `encoding/json`)
- Keep handler files focused (<200 lines)
- Return structured JSON errors: `{"error": "message", "code": "ERROR_CODE"}`
- Use context for cancellation and timeouts
- Prefer table-driven tests
- File naming: `snake_case.go`

### Svelte Frontend

- Use Svelte 5 runes syntax (`$state`, `$derived`, `$effect`)
- Keep components small and focused (<150 lines)
- Use Tailwind for styling; avoid custom CSS unless necessary
- API calls go through `/lib/api/` client module
- Shared state in `/lib/stores/`
- Component naming: `PascalCase.svelte`

### API Design

- RESTful endpoints under `/api/v1/`
- Consistent JSON response format
- Pagination: `?limit=50&offset=0`
- Use appropriate HTTP methods and status codes

## Making Changes

### Branch Workflow

1. Fork the repository (external contributors)
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes
4. Ensure tests pass: `make test`
5. Ensure linting passes: `make lint`
6. Commit with a descriptive message
7. Push and open a pull request

### Commit Messages

Write clear, descriptive commit messages:

```
Add whitelist bulk import feature

- Add CSV parsing for bulk pubkey import
- Display progress during import
- Handle duplicate detection
```

- Use present tense ("Add feature" not "Added feature")
- First line is a brief summary (50 chars or less)
- Add details in the body if needed

### Pull Request Process

1. Ensure your PR has a clear title and description
2. Link any related issues
3. Ensure CI checks pass
4. Request review if needed
5. Address feedback and update as needed

## Testing

### Unit Tests

**Go:**
```bash
# Run all Go tests
make test-api

# Run specific package tests
cd app/api && go test -v ./internal/handlers/...

# Run with coverage
cd app/api && go test -cover ./...
```

**Svelte:**
```bash
# Run all UI tests
make test-ui

# Run in watch mode
cd app/ui && npm run test:watch
```

### E2E Tests

```bash
cd app/ui

# Run E2E tests headless
npm run test:e2e

# Run with browser visible
npm run test:e2e:headed

# Debug mode
npm run test:e2e:debug
```

## Building & Packaging

### Production Build

```bash
# Build everything
make build

# Build only API (outputs to bin/roostr-api)
make build-api

# Build only UI (outputs to app/ui/build)
make build-ui
```

### Docker Image

```bash
# Build Docker image
docker build -f platforms/umbrel/Dockerfile -t rdoiron/roostr:0.1.0 .
```

### Platform Packages

**Umbrel:**
```bash
make package-umbrel
```

**StartOS:**
```bash
cd platforms/startos
make x86    # Build x86_64 image
make pack   # Create .s9pk package
```

## Common Makefile Commands

| Command | Description |
|---------|-------------|
| `make dev` | Run API and UI dev servers |
| `make build` | Build everything for production |
| `make test` | Run all tests |
| `make lint` | Lint all code |
| `make fmt` | Format all code |
| `make deps` | Install dependencies |
| `make clean` | Remove build artifacts |
| `make help` | Show all available commands |

## Getting Help

- **Questions:** Open a GitHub issue with the "question" label
- **Bugs:** Open an issue with steps to reproduce
- **Discussion:** Feel free to discuss in issues before starting work on larger changes

## Code of Conduct

Be respectful and constructive. We're all here to build something useful for the Nostr community.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
