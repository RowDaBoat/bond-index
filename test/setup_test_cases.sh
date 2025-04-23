#!/bin/bash

[ "$1" = "--dry-run" ] && DRY_RUN="--dry-run" || DRY_RUN=""

get_wallet_address() {
    ord wallet --name "$1" receive | jq -r '.addresses[0]'
}

create_name_inscription() {
    local wallet_name=$1
    local btc_name=$2

    echo $btc_name > name_file.txt
    echo "Creating name inscription for $btc_name"

    address=$(get_wallet_address "$wallet_name")
    ord wallet inscribe $DRY_RUN --postage 546sat \
        --fee-rate 1 --destination "$address" --file name_file.txt
    rm name_file.txt
}

create_routing_inscription() {
    local wallet_name=$1
    local btc_name=$2
    local nostr_npub=$3

    jq --arg name "$btc_name" --arg npub "$nostr_npub" \
        '.name = $name | .nostr_npub = $npub' \
        routing_template.json > routing_file.json
    echo "Creating routing inscription for $btc_name"
    echo "Routing file: $(cat routing_file.json)"

    address=$(get_wallet_address "$wallet_name")
    ord wallet inscribe $DRY_RUN --postage 546sat \
        --fee-rate 1 --destination "$address" --file routing_file.json
    rm routing_file.json
}

router_npub="npub1xwjap7x26602fze5epv5lpf3v5hgefjzpy49p9ulyx2sdam8jl3qdrfljs"

echo "Test - Name with routing on Alice's wallet."
create_name_inscription "alice" "yarrharr.btc"
create_routing_inscription "alice" "yarrharr.btc" "$router_npub"

echo "Test - Name without routing on Alice's wallet."
create_name_inscription "alice" "yohoho.btc"

echo "Test - Routing without name on Alice's wallet."
create_routing_inscription "alice" "ayeaye.btc" "$router_npub"

echo "Test - Name and routing on different wallets."
create_name_inscription "alice" "ahoy.btc"
create_routing_inscription "bob" "ahoy.btc" "$router_npub"

echo "Test cases setup complete!"
