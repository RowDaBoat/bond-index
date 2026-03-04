#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

echo -n "Creating and funding wallet... "
ord_wallet create &>$LOG_ORD
WALLET_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
mine 101 "$WALLET_ADDRESS"
echo "ok."

ord_sync

echo -n "Creating btcname inscription... "
NAME_FILE="$TEST_DIR/name.txt"
NAME_ADDRESS=$(ord_wallet receive 2>$LOG_ORD | jq -r '.addresses[0]')
echo -n "test.btc" > "$NAME_FILE"
INSCRIBE_OUTPUT=$(ord_wallet inscribe --fee-rate 1 --no-backup --destination "$NAME_ADDRESS" --file "$NAME_FILE" 2>$LOG_ORD)
NAME_INSCRIPTION_ID=$(echo "$INSCRIBE_OUTPUT" | jq -r '.inscriptions[0].id')
echo "ok. ($NAME_INSCRIPTION_ID)"

echo -n "Mining name inscription... "
mine 7 "$WALLET_ADDRESS"
ord_sync
echo "ok."

echo -n "Creating routing inscription (child of name)... "
ROUTING_FILE="$TEST_DIR/routing.json"
cat > "$ROUTING_FILE" <<EOF
{"p":"btcname","op":"routing","name":"test.btc","nostr_npub":"npub1testpubkey","nostr_relays":["wss://relay.example.com"]}
EOF
#--parent "$NAME_INSCRIPTION_ID" 
ord_wallet inscribe --fee-rate 1 --no-backup --destination "$NAME_ADDRESS" --file "$ROUTING_FILE" &>$LOG_ORD
echo "ok."

echo -n "Mining routing inscription... "
mine 7 "$WALLET_ADDRESS"
ord_sync
bond_sync
echo "ok."

echo "Verifying name resolves with routing... "
response=$(bond_client "/name/test.btc")
assert_equals "name resolves with nostr npub" "npub1testpubkey" "$(echo "$response" | jq -r '.nostr_npub')"
assert_equals "name resolves with relay" "wss://relay.example.com" "$(echo "$response" | jq -r '.nostr_relays[0]')"

echo "All assertions passed."
