#!/bin/bash

# Roostr Backup Script for StartOS
# Creates a compressed tarball of all data and outputs to stdout

set -e

DATA_DIR="/data"

# Create tarball of data directory and write to stdout
# This includes:
# - nostr.db (relay database)
# - roostr.db (app database)
# - config.toml (relay configuration)
tar -czf - -C "$DATA_DIR" .
