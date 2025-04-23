#!/bin/bash

set -e

router_npub="npub1xwjap7x26602fze5epv5lpf3v5hgefjzpy49p9ulyx2sdam8jl3qdrfljs"

expect_success() {
    local test_case=$1
    local name=$2
    local expected_npub=$3

    echo "Testing $test_case: $name"
    response=$(curl -s "localhost:8080/name/$name")
    
    if [ "$response" = "$expected_npub" ]; then
        echo "✅ PASS: Got expected npub for $name"
    else
        echo "❌ FAIL: Expected '$expected_npub' but got '$response' for $name"
        exit 1
    fi
}

expect_failure() {
    local test_case=$1
    local name=$2

    echo "Testing $test_case: $name"
    response=$(curl -s "localhost:8080/name/$name")
    
    if [ "$response" = "" ]; then
        echo "✅ PASS: Got empty response as expected for $name"
    else
        echo "❌ FAIL: Expected empty response but got '$response' for $name"
        exit 1
    fi
}

expect_success "Name with routing on same wallet"       "yarrharr.btc" "$router_npub"
expect_failure "Name without routing"                   "yohoho.btc"
expect_failure "Routing without name"                   "ayeaye.btc"
expect_failure "Name and routing on different wallets"  "ahoy.btc"

echo "All tests completed successfully!" 
