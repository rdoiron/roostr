#!/bin/bash

# Roostr Restore Script for StartOS
# Reads a compressed tarball from stdin and extracts to data directory

set -e

DATA_DIR="/data"

# Clear existing data
rm -rf "${DATA_DIR:?}"/*

# Extract tarball from stdin to data directory
tar -xzf - -C "$DATA_DIR"

# Ensure proper permissions
chown -R appuser:appuser "$DATA_DIR" 2>/dev/null || true
