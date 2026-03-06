#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

ord_wallet create &>$LOG_ORD
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"

ord_sync

NAME_FILE="$TEST_DIR/name.txt"
NAME_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
echo -n "test.btc" > "$NAME_FILE"
INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --destination "$NAME_ADDRESS" --file "$NAME_FILE" 2>$LOG_ORD)
NAME_INSCRIPTION_ID=$(echo "$INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')

mine 7 "$WALLET_ADDRESS"
ord_sync

ROUTING_FILE="$TEST_DIR/routing.json"
cat > "$ROUTING_FILE" <<EOF
{"p":"btcname","op":"routing","nostr_npub":"npub1testpubkey","nostr_relays":["wss://relay.example.com"]}
EOF
ord_wallet inscribe --fee-rate 1 --no-backup --parent "$NAME_INSCRIPTION_ID"  --file "$ROUTING_FILE" &>$LOG_ORD
#ord_wallet inscribe --fee-rate 1 --no-backup --destination "$NAME_ADDRESS" --file "$ROUTING_FILE" &>$LOG_ORD

mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync

response=$(bond_client "/resolve/test.btc")
assert_equals "routing returns expected nostr npub" "npub1testpubkey" "$(echo "$response" | jq -r '.nostr_npub')"
assert_equals "routing returns expected relay" "wss://relay.example.com" "$(echo "$response" | jq -r '.nostr_relays[0]')"

echo "All assertions passed."
