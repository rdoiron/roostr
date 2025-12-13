#!/bin/bash

# Health check script for StartOS

set -e

# Check if API is responding
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

if [ "$response" = "200" ]; then
    echo "API is healthy"
    exit 0
else
    echo "API health check failed with status: $response"
    exit 1
fi
