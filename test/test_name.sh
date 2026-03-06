#!/bin/bash

set -e

# Setup
source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

# Create and fund wallet
ord_wallet create &>$LOG_ORD
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"
ord_sync

# Create btcname inscription
NAME_FILE="$TEST_DIR/name.txt"
echo -n "test.btc" > "$NAME_FILE"
INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" 2>$LOG_ORD)

# Mine and sync inscription
mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

# Verify name metadata endpoint
name_response=$(bond_client "/name/test.btc")
assert_equals "name endpoint returns correct domain" "test.btc" "$(echo "$name_response" | jq -r '.domain')"
assert_equals "name endpoint returns correct ordinal id" "$(echo "$INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')" "$(echo "$name_response" | jq -r '.ordinal_id')"

# Verify resolution response
response=$(bond_client "/resolve/test.btc")
assert_equals "name exists but has no routing" '{"error":"routing not found"}' "$response"

echo "All assertions passed."
