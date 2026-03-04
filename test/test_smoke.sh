#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

echo ""
echo "Smoke test: services started and responding"
echo ""

assert_success "bitcoind responds to getblockchaininfo" \
    bitcoin_cli getblockchaininfo

assert_success "ord status endpoint is reachable" \
    curl -sf "http://localhost:$ORD_PORT/status"

assert_success "bond health endpoint is reachable" \
    curl -sf "http://localhost:$BOND_PORT/health"

echo ""
echo "All smoke test assertions passed."
