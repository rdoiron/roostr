# Roostr Umbrel Dockerfile
# Multi-stage build for Go API, Svelte UI, and nostr-rs-relay

# Stage 1: Build Svelte UI
FROM node:20-alpine AS ui-builder
WORKDIR /app
COPY app/ui/package*.json ./
RUN npm ci
COPY app/ui/ ./
RUN npm run build

# Stage 2: Build Go API
FROM golang:1.22-bookworm AS api-builder
WORKDIR /app

# Install SQLite dev headers for CGO
RUN apt-get update && apt-get install -y libsqlite3-dev

COPY app/api/go.mod app/api/go.sum ./
RUN go mod download

COPY app/api/ ./
RUN CGO_ENABLED=1 go build -o /roostr-api ./cmd/server

# Stage 3: Build nostr-rs-relay
FROM rust:1-bookworm AS relay-builder

RUN apt-get update && apt-get install -y cmake protobuf-compiler libsqlite3-dev

WORKDIR /app
RUN git clone --depth 1 --branch 0.8.13 https://github.com/scsibug/nostr-rs-relay.git .
RUN cargo build --release

# Stage 4: Runtime
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    sqlite3 \
    wget \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd -m -u 1000 appuser

# Copy binaries
COPY --from=api-builder /roostr-api /usr/local/bin/roostr-api
COPY --from=relay-builder /app/target/release/nostr-rs-relay /usr/local/bin/nostr-rs-relay

# Copy UI build
COPY --from=ui-builder /app/build /app/ui

# Set permissions
RUN chmod +x /usr/local/bin/roostr-api /usr/local/bin/nostr-rs-relay
RUN chown -R appuser:appuser /app

# Create data directory
RUN mkdir -p /data && chown appuser:appuser /data

WORKDIR /app

# Environment defaults
ENV PORT=8080 \
    RELAY_PORT=7000 \
    RELAY_DB_PATH=/data/nostr.db \
    APP_DB_PATH=/data/roostr.db \
    CONFIG_PATH=/data/config.toml \
    RELAY_BINARY=/usr/local/bin/nostr-rs-relay \
    STATIC_DIR=/app/ui \
    RUST_LOG=info,nostr_rs_relay=info

EXPOSE 8080 7000

# Copy entrypoint script
COPY --chmod=755 platforms/umbrel/entrypoint.sh /entrypoint.sh

USER appuser

ENTRYPOINT ["/entrypoint.sh"]
