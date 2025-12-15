#!/bin/bash

# Roostr exports.sh
# Export environment variables for inter-app communication on Umbrel
# Other apps can use these to connect to the Roostr relay

# Relay WebSocket URL for Nostr clients
export APP_ROOSTR_RELAY_URL="ws://roostr_app_1:7000"

# Relay port for direct connections
export APP_ROOSTR_RELAY_PORT="7000"

# API endpoint for apps that want to interact with Roostr
export APP_ROOSTR_API_URL="http://roostr_app_1:8080/api/v1"
