#!/usr/bin/env bash
# ABOUTME: Validates hex still works after a mux library upgrade
# ABOUTME: Run this after updating github.com/2389-research/mux in go.mod

set -eo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0

log_step() { echo -e "\n${CYAN}▶ $1${NC}"; }
log_pass() { echo -e "${GREEN}✓ $1${NC}"; ((PASS++)) || true; }
log_fail() { echo -e "${RED}✗ $1${NC}"; ((FAIL++)) || true; }
log_warn() { echo -e "${YELLOW}⚠ $1${NC}"; }

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════╗"
echo "║     Mux Upgrade Validation Check           ║"
echo "╚════════════════════════════════════════════╝"
echo -e "${NC}"

# Show current mux version
log_step "Checking current mux version"
MUX_VERSION=$(grep 'github.com/2389-research/mux' go.mod | awk '{print $2}')
echo "Current mux version: ${MUX_VERSION}"

# Step 1: go mod verify
log_step "Verifying go.mod"
if go mod verify >/dev/null 2>&1; then
    log_pass "go.mod verified"
else
    log_fail "go.mod verification failed"
fi

# Step 2: go mod tidy
log_step "Tidying dependencies"
if go mod tidy 2>&1; then
    log_pass "go mod tidy succeeded"
else
    log_fail "go mod tidy failed"
fi

# Step 3: Build
log_step "Building hex binary"
if go build -o ./hex ./cmd/hex/ 2>&1; then
    log_pass "Build succeeded"
else
    log_fail "Build failed - cannot continue"
    exit 1
fi

# Step 4: Unit tests
log_step "Running unit tests"
if go test ./... -count=1 >/dev/null 2>&1; then
    log_pass "All unit tests passed"
else
    log_fail "Some unit tests failed"
fi

# Step 5: Provider tests
log_step "Running provider tests (mux adapters)"
if go test ./internal/providers/... -count=1 >/dev/null 2>&1; then
    log_pass "Provider tests passed"
else
    log_fail "Provider tests failed"
fi

# Step 6: MCP tests
log_step "Running MCP tests"
if go test ./internal/mcp/... -count=1 >/dev/null 2>&1; then
    log_pass "MCP tests passed"
else
    log_fail "MCP tests failed"
fi

# Step 7: hex --version
log_step "Checking hex --version"
if ./hex --version 2>&1 | grep -q "hex version"; then
    log_pass "hex --version works"
else
    log_fail "hex --version failed"
fi

# Step 8: hex --help
log_step "Checking hex --help"
HELP_OUT=$(./hex --help 2>&1)
if echo "$HELP_OUT" | grep -q "Hex is"; then
    log_pass "hex --help works"
else
    log_fail "hex --help failed"
fi

# Step 9: Scenario tests
log_step "Running scenario tests"
SCENARIO_DIR=".scratch"
if [ -d "$SCENARIO_DIR" ]; then
    SCENARIO_PASS=0
    SCENARIO_FAIL=0
    for scenario in "$SCENARIO_DIR"/scenario_*.sh; do
        if [ -f "$scenario" ]; then
            SCENARIO_NAME=$(basename "$scenario")
            echo -n "  $SCENARIO_NAME... "
            if timeout 60 bash "$scenario" >/dev/null 2>&1; then
                echo -e "${GREEN}PASS${NC}"
                ((SCENARIO_PASS++)) || true
            else
                echo -e "${RED}FAIL${NC}"
                ((SCENARIO_FAIL++)) || true
            fi
        fi
    done
    if [ $SCENARIO_FAIL -eq 0 ]; then
        log_pass "All $SCENARIO_PASS scenario tests passed"
    else
        log_fail "$SCENARIO_FAIL scenario tests failed"
    fi
else
    log_warn "No scenario tests found in $SCENARIO_DIR"
fi

# Step 10: Smoke test
log_step "Smoke test (print mode)"
if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    if timeout 30 ./hex -p "Say 'mux check ok'" 2>&1 | grep -qi "ok"; then
        log_pass "Print mode smoke test passed"
    else
        log_fail "Print mode smoke test failed"
    fi
else
    log_warn "Skipping smoke test (ANTHROPIC_API_KEY not set)"
fi

# Summary
echo ""
echo -e "${CYAN}════════════════════════════════════════════${NC}"
echo "Mux Version: ${MUX_VERSION}"
echo -e "${GREEN}Passed: $PASS${NC}"
if [ $FAIL -gt 0 ]; then
    echo -e "${RED}Failed: $FAIL${NC}"
    echo ""
    echo -e "${RED}⚠ Mux upgrade validation FAILED${NC}"
    exit 1
else
    echo "Failed: 0"
    echo ""
    echo -e "${GREEN}✓ Mux upgrade validation PASSED${NC}"
    exit 0
fi
