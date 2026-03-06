#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

# Create and fund wallet
ord_wallet create &>$LOG_ORD
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"
ord_sync

NAME_FILE="$TEST_DIR/name.txt"

# First btcname inscription for the same domain
DOMAIN="dup.btc"
echo -n "$DOMAIN" > "$NAME_FILE"
FIRST_INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" 2>$LOG_ORD)
FIRST_ID=$(echo "$FIRST_INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')

mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

# Verify name metadata reflects the first inscription
name_response=$(bond_client "/name/$DOMAIN")
assert_equals "name endpoint returns correct domain for first inscription" "$DOMAIN" "$(echo "$name_response" | jq -r '.domain')"
assert_equals "name endpoint returns first ordinal id" "$FIRST_ID" "$(echo "$name_response" | jq -r '.ordinal_id')"

# Second btcname inscription with the same domain
echo -n "$DOMAIN" > "$NAME_FILE"
SECOND_INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" 2>$LOG_ORD)
SECOND_ID=$(echo "$SECOND_INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')

mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

# Verify duplicated name is ignored and first ordinal id is still used
name_response=$(bond_client "/name/$DOMAIN")
assert_equals "duplicated name keeps first ordinal id" "$FIRST_ID" "$(echo "$name_response" | jq -r '.ordinal_id')"

echo "All assertions passed."
