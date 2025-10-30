#!/bin/bash

set -e

[ "$1" = "--dry-run" ] && DRY_RUN="--dry-run" || DRY_RUN=""

get_wallet_address() {
    ord --testnet4 wallet --name "$1" receive | jq -r '.addresses[0]'
}

create_name_inscription() {
    local destination=$1
    local btc_name=$2

    echo -n $btc_name > name_file.txt
    echo "Creating name inscription for $btc_name"

    ord --testnet4 wallet inscribe $DRY_RUN --postage 546sat \
        --fee-rate 1 --destination "$destination" --file name_file.txt
    rm name_file.txt
}

create_routing_inscription() {
    local destination=$1
    local btc_name=$2
    local nostr_npub=$3

    jq --arg name "$btc_name" --arg npub "$nostr_npub" \
        '.name = $name | .nostr_npub = $npub' \
        routing_template.json > routing_file.json
    echo "Creating routing inscription for $btc_name"
    echo "Routing file: $(cat routing_file.json)"

    ord --testnet4 wallet inscribe $DRY_RUN --postage 546sat \
        --fee-rate 1 --destination "$destination" --file routing_file.json
    rm routing_file.json
}

router_npub="npub1xwjap7x26602fze5epv5lpf3v5hgefjzpy49p9ulyx2sdam8jl3qdrfljs"

echo "Test - Name with routing on Alice's wallet."
destination=$(get_wallet_address "alice")
create_name_inscription "$destination" "bond.btc"
create_routing_inscription "$destination" "bond.btc" "$router_npub"

echo "Test - Name without routing on Alice's wallet."
destination=$(get_wallet_address "alice")
create_name_inscription "$destination" "noroute.btc"

echo "Test - Routing without name on Alice's wallet."
destination=$(get_wallet_address "alice")
create_routing_inscription "$destination" "noname.btc" "$router_npub"

echo "Test - Name and routing on different wallets."
destination=$(get_wallet_address "alice")
create_name_inscription "$destination" "nowallet.btc"
destination=$(get_wallet_address "bob")
create_routing_inscription "$destination" "nowallet.btc" "$router_npub"

echo "Test cases setup complete!"
