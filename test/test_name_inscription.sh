#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

echo -n "Creating and funding wallet..."
ord_wallet create &>$LOG
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"
echo " done."

sync_ord

echo -n "Creating btcname inscription... "
NAME_FILE="$TEST_DIR/name.txt"
echo -n "test.btc" > "$NAME_FILE"
ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" &>$LOG
echo "done."

echo -n "Mining inscription... "
mine 7 "$WALLET_ADDRESS"
echo "done."

sync_ord

echo -n "Waiting for bond to index name... "
wait_for "bond to index test.btc" \
    'bond_client /name/test.btc | grep -q "routing not found"' \
    15 || exit 1
echo "done."

echo -n "Verifying response... "
# Verify the response
response=$(bond_client "/name/test.btc")
assert_equals "name exists but has no routing" '{"error":"routing not found"}' "$response"
echo "ok."

echo "All assertions passed."
echo ""
