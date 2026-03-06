#!/bin/bash

set -e

# Setup
source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

# Verify response
response=$(bond_client "/resolve/test.btc")
assert_equals "name does not exist" '{"error":"name not found"}' "$response"
echo "All assertions passed."
