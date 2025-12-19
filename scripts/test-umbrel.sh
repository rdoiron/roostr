#!/bin/bash
# Umbrel Integration Test Script for Roostr
# Automated API endpoint testing with curl
#
# Usage:
#   ./scripts/test-umbrel.sh                    # Test localhost:8080
#   ROOSTR_URL=http://umbrel.local:8880 ./scripts/test-umbrel.sh
#
# Exit codes:
#   0 = All tests passed
#   1 = One or more tests failed

# Don't exit on error - we want to continue running all tests
# set -e

# Configuration
BASE_URL="${ROOSTR_URL:-http://localhost:8080}"
VERBOSE="${VERBOSE:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Test pubkey for CRUD operations (generate a random one)
TEST_PUBKEY="$(printf '%064d' $RANDOM$RANDOM)"
TEST_NPUB="npub1test$(printf '%058d' $RANDOM)"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Roostr Umbrel Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Target: ${YELLOW}${BASE_URL}${NC}"
echo ""
echo -e "${YELLOW}Note: Some tests require nostr-rs-relay to be running.${NC}"
echo -e "${YELLOW}Tests may fail with 503 in dev mode without relay DB.${NC}"
echo ""

# Test function
# Usage: test_endpoint "Name" "METHOD" "/endpoint" expected_status [json_body]
test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local expected_status="$4"
    local body="$5"

    local url="${BASE_URL}${endpoint}"
    local status
    local response

    # Build curl command
    local curl_cmd="curl -s -o /dev/null -w '%{http_code}' -X ${method}"

    if [ -n "$body" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '${body}'"
    fi

    curl_cmd="$curl_cmd '${url}'"

    # Execute
    status=$(eval "$curl_cmd" 2>/dev/null) || status="000"

    # Check result
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}[PASS]${NC} ${name} (${status})"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}[FAIL]${NC} ${name} - Expected ${expected_status}, got ${status}"
        ((FAILED++))
        return 1
    fi
}

# Test with response validation
# Usage: test_endpoint_json "Name" "METHOD" "/endpoint" expected_status "jq_filter" "expected_value"
test_endpoint_json() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local expected_status="$4"
    local jq_filter="$5"
    local expected_value="$6"
    local body="$7"

    local url="${BASE_URL}${endpoint}"
    local response
    local status
    local actual_value

    # Execute curl
    if [ -n "$body" ]; then
        response=$(curl -s -w $'\n%{http_code}' -X "${method}" -H 'Content-Type: application/json' -d "${body}" "${url}" 2>/dev/null) || response=$'\n000'
    else
        response=$(curl -s -w $'\n%{http_code}' -X "${method}" "${url}" 2>/dev/null) || response=$'\n000'
    fi

    # Extract status (last line) and body (everything else)
    status=$(echo "$response" | tail -n1)
    local body_response=$(echo "$response" | sed '$d')

    # Check status first
    if [ "$status" != "$expected_status" ]; then
        echo -e "${RED}[FAIL]${NC} ${name} - Expected status ${expected_status}, got ${status}"
        ((FAILED++))
        return 1
    fi

    # Check JSON value if filter provided
    if [ -n "$jq_filter" ]; then
        actual_value=$(echo "$body_response" | jq -r "$jq_filter" 2>/dev/null) || actual_value=""

        if [ "$actual_value" = "$expected_value" ]; then
            echo -e "${GREEN}[PASS]${NC} ${name} (${jq_filter}=${actual_value})"
            ((PASSED++))
            return 0
        else
            echo -e "${RED}[FAIL]${NC} ${name} - Expected ${jq_filter}='${expected_value}', got '${actual_value}'"
            ((FAILED++))
            return 1
        fi
    fi

    echo -e "${GREEN}[PASS]${NC} ${name} (${status})"
    ((PASSED++))
    return 0
}

# Skip test helper
skip_test() {
    local name="$1"
    local reason="$2"
    echo -e "${YELLOW}[SKIP]${NC} ${name} - ${reason}"
    ((SKIPPED++))
}

# ============================================
# Health & Status Tests
# ============================================
echo ""
echo -e "${BLUE}--- Health & Status ---${NC}"

test_endpoint "Health check (/health)" "GET" "/health" "200"
test_endpoint_json "Health status field" "GET" "/health" "200" ".status" "ok"

# ============================================
# Setup Status Tests
# ============================================
echo ""
echo -e "${BLUE}--- Setup Status ---${NC}"

test_endpoint "Setup status endpoint" "GET" "/api/v1/setup/status" "200"

# ============================================
# Dashboard & Stats Tests
# ============================================
echo ""
echo -e "${BLUE}--- Dashboard & Stats ---${NC}"

test_endpoint "Stats summary" "GET" "/api/v1/stats/summary" "200"
test_endpoint "Events by kind" "GET" "/api/v1/stats/events-by-kind" "200"
test_endpoint "Events over time" "GET" "/api/v1/stats/events-over-time" "200"
test_endpoint "Top authors" "GET" "/api/v1/stats/top-authors?limit=5" "200"

# ============================================
# Relay Status Tests
# ============================================
echo ""
echo -e "${BLUE}--- Relay Status ---${NC}"

test_endpoint "Relay status" "GET" "/api/v1/relay/status" "200"
test_endpoint "Relay URLs" "GET" "/api/v1/relay/urls" "200"
# Note: Relay logs requires nostr-rs-relay to be running (may return 503 in dev)
test_endpoint "Relay logs" "GET" "/api/v1/relay/logs?limit=10" "200"

# ============================================
# Access Control Tests
# ============================================
echo ""
echo -e "${BLUE}--- Access Control ---${NC}"

test_endpoint "Access mode GET" "GET" "/api/v1/access/mode" "200"
test_endpoint "Whitelist GET" "GET" "/api/v1/access/whitelist" "200"
test_endpoint "Blacklist GET" "GET" "/api/v1/access/blacklist" "200"
test_endpoint "Pricing GET" "GET" "/api/v1/access/pricing" "200"
test_endpoint "Paid users GET" "GET" "/api/v1/access/paid-users" "200"
test_endpoint "Revenue GET" "GET" "/api/v1/access/revenue" "200"

# ============================================
# Events Tests
# ============================================
echo ""
echo -e "${BLUE}--- Events ---${NC}"

test_endpoint "Events list" "GET" "/api/v1/events?limit=10" "200"
test_endpoint "Events list with kind filter" "GET" "/api/v1/events?limit=10&kinds=1" "200"
test_endpoint "Recent events" "GET" "/api/v1/events/recent" "200"
test_endpoint "Export estimate" "GET" "/api/v1/events/export/estimate" "200"

# ============================================
# Configuration Tests
# ============================================
echo ""
echo -e "${BLUE}--- Configuration ---${NC}"

test_endpoint "Config GET" "GET" "/api/v1/config" "200"

# ============================================
# Storage Tests
# ============================================
echo ""
echo -e "${BLUE}--- Storage ---${NC}"

test_endpoint "Storage status" "GET" "/api/v1/storage/status" "200"
test_endpoint "Retention policy GET" "GET" "/api/v1/storage/retention" "200"
test_endpoint "Deletion requests" "GET" "/api/v1/storage/deletion-requests" "200"
test_endpoint "Storage estimate" "GET" "/api/v1/storage/estimate?before_date=2020-01-01T00:00:00Z" "200"

# ============================================
# Sync Tests
# ============================================
echo ""
echo -e "${BLUE}--- Sync ---${NC}"

test_endpoint "Sync status" "GET" "/api/v1/sync/status" "200"
test_endpoint "Sync history" "GET" "/api/v1/sync/history" "200"
test_endpoint "Sync relays" "GET" "/api/v1/sync/relays" "200"

# ============================================
# Lightning Tests
# ============================================
echo ""
echo -e "${BLUE}--- Lightning ---${NC}"

test_endpoint "Lightning status" "GET" "/api/v1/lightning/status" "200"

# ============================================
# Settings Tests
# ============================================
echo ""
echo -e "${BLUE}--- Settings ---${NC}"

test_endpoint "Timezone GET" "GET" "/api/v1/settings/timezone" "200"

# ============================================
# Support/About Tests
# ============================================
echo ""
echo -e "${BLUE}--- Support ---${NC}"

test_endpoint "Support config" "GET" "/api/v1/support/config" "200"

# ============================================
# Public Endpoints Tests
# ============================================
echo ""
echo -e "${BLUE}--- Public Endpoints ---${NC}"

test_endpoint "Public relay info" "GET" "/public/relay-info" "200"

# ============================================
# NIP-05 Resolution Tests
# ============================================
echo ""
echo -e "${BLUE}--- NIP-05 ---${NC}"

# NIP-05 resolution - expects error for invalid identifier (400, 404, or 502 are all valid)
# We test that the endpoint responds, not the specific error code
skip_test "NIP-05 resolution" "Requires external network access"

# ============================================
# Summary
# ============================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Passed:${NC}  ${PASSED}"
echo -e "${RED}Failed:${NC}  ${FAILED}"
echo -e "${YELLOW}Skipped:${NC} ${SKIPPED}"
echo ""

TOTAL=$((PASSED + FAILED))
if [ $TOTAL -gt 0 ]; then
    PERCENT=$((PASSED * 100 / TOTAL))
    echo -e "Pass rate: ${PERCENT}%"
fi

echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
