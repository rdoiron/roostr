#!/bin/bash

# Health check script for StartOS
# Outputs YAML format required by StartOS

# Check if API is responding
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null)

if [ "$response" = "200" ]; then
    echo "result:"
    echo "  type: success"
    exit 0
else
    echo "result:"
    echo "  type: failure"
    echo "  message: API health check failed with status $response"
    exit 0
fi
