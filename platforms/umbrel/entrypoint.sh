#!/bin/bash
set -e

# Roostr Umbrel Entrypoint
# Starts nostr-rs-relay in background, then runs the Roostr API

CONFIG_PATH="${CONFIG_PATH:-/data/config.toml}"
RELAY_PORT="${RELAY_PORT:-7000}"

# Create default config.toml if not exists
if [ ! -f "$CONFIG_PATH" ]; then
    echo "Creating default relay configuration..."
    cat > "$CONFIG_PATH" << EOF
[info]
relay_url = "wss://your-relay.example.com/"
name = "Roostr Relay"
description = "A private Nostr relay managed by Roostr"
pubkey = ""
contact = ""

[database]
data_directory = "/data"
engine = "sqlite"

[network]
address = "0.0.0.0"
port = ${RELAY_PORT}

[options]
reject_future_seconds = 1800

[authorization]
nip42_auth = false

[verified_users]
mode = "passive"

[limits]
messages_per_sec = 5
subscriptions_per_min = 10
max_event_bytes = 131072
max_ws_message_bytes = 131072
max_ws_frame_bytes = 131072
broadcast_buffer = 16384
event_persist_buffer = 4096
EOF
fi

# Function to handle shutdown
cleanup() {
    echo "Shutting down..."
    # Send SIGTERM to relay process
    if [ -n "$RELAY_PID" ] && kill -0 "$RELAY_PID" 2>/dev/null; then
        kill -TERM "$RELAY_PID"
        wait "$RELAY_PID" 2>/dev/null || true
    fi
    exit 0
}

trap cleanup SIGTERM SIGINT

# Start nostr-rs-relay in background
echo "Starting nostr-rs-relay on port ${RELAY_PORT}..."
/usr/local/bin/nostr-rs-relay --config "$CONFIG_PATH" &
RELAY_PID=$!

# Wait for relay to be ready
echo "Waiting for relay to start..."
sleep 2

# Check if relay is still running
if ! kill -0 "$RELAY_PID" 2>/dev/null; then
    echo "ERROR: Relay failed to start"
    exit 1
fi

echo "Relay started with PID $RELAY_PID"

# Start Roostr API (serves UI + API)
echo "Starting Roostr API on port ${PORT:-8080}..."
exec /usr/local/bin/roostr-api
