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

# First btcname inscription for dup.btc
DOMAIN="dup.btc"
echo -n "$DOMAIN" > "$NAME_FILE"
FIRST_INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" 2>$LOG_ORD)
FIRST_NAME_ID=$(echo "$FIRST_INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')

# Confirm first inscription so wallet has UTXOs for the second
mine 7 "$WALLET_ADDRESS"
ord_sync

# Second btcname inscription for dup.btc (duplicate name)
echo -n "$DOMAIN" > "$NAME_FILE"
SECOND_INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --file "$NAME_FILE" 2>$LOG_ORD)
SECOND_NAME_ID=$(echo "$SECOND_INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')

# Mine inscriptions and sync
mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

# Create routing inscription as a child of the original name
ROUTING_FILE="$TEST_DIR/routing.json"
cat > "$ROUTING_FILE" <<EOF
{"p":"btcname","op":"routing","nostr_npub":"npub1original","nostr_relays":["wss://relay.original.example.com"]}
EOF
ord_wallet inscribe --fee-rate 1 --no-backup --parent "$FIRST_NAME_ID" --file "$ROUTING_FILE" &>$LOG_ORD

# Create routing inscription as a child of the duplicated name
ROUTING_FILE="$TEST_DIR/routing.json"
cat > "$ROUTING_FILE" <<EOF
{"p":"btcname","op":"routing","nostr_npub":"npub1duplicated","nostr_relays":["wss://relay.duplicated.example.com"]}
EOF
ord_wallet inscribe --fee-rate 1 --no-backup --parent "$SECOND_NAME_ID" --file "$ROUTING_FILE" &>$LOG_ORD

# Mine inscriptions and sync
mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

# Resolve should respond using the routing on the original name
response=$(bond_client "/resolve/$DOMAIN")
assert_equals "resolve responds using the routing on the original name" "npub1original" "$(echo "$response" | jq -r '.nostr_npub')"
assert_equals "resolve responds using the relays on the original name" "wss://relay.original.example.com" "$(echo "$response" | jq -r '.nostr_relays[0]')"

echo "All assertions passed."
