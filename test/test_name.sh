#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

echo -n "Creating and funding wallet... "
ord_wallet create &>$LOG_ORD
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"
ord_sync
echo "ok."

echo -n "Creating btcname inscription... "
NAME_FILE="$TEST_DIR/name.txt"
echo -n "test.btc" > "$NAME_FILE"
ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" &>$LOG_ORD
echo "ok."

echo -n "Mining inscription... "
mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync
echo "ok."

echo "Verifying response... "
response=$(bond_client "/name/test.btc")
assert_equals "name exists but has no routing" '{"error":"routing not found"}' "$response"

echo "All assertions passed."
