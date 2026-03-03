#!/bin/bash

set -e

source "$(dirname "$0")/test_setup.sh" "$@"
trap cleanup EXIT
start_services

echo ""
echo "Smoke test: services started and responding"
echo ""
