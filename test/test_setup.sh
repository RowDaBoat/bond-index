#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

LOG_BITCOIND=/dev/null
LOG_ORD=/dev/null
LOG_BOND=/dev/stdout

for arg in "$@"; do
    case "$arg" in
        --debug)          LOG_BITCOIND=/dev/stdout; LOG_ORD=/dev/stdout; LOG_BOND=/dev/stdout ;;
        --debug-bitcoind) LOG_BITCOIND=/dev/stdout ;;
        --debug-ord)      LOG_ORD=/dev/stdout ;;
        --debug-bond)     LOG_BOND=/dev/stdout ;;
        --quiet)          LOG_BITCOIND=/dev/null; LOG_ORD=/dev/null; LOG_BOND=/dev/null ;;
    esac
done

ORD_PORT=4080
BOND_PORT=8080

TEST_DIR=$(mktemp -d)
BITCOIN_DATADIR="$TEST_DIR/bitcoin"
ORD_DATADIR="$TEST_DIR/ord"
BOND_DATADIR="$TEST_DIR/bond-data"
mkdir -p "$BITCOIN_DATADIR" "$ORD_DATADIR" "$BOND_DATADIR"

BOND_PID=""
ORD_PID=""
BITCOIND_PID=""

cleanup() {
    echo ""
    echo -n "Stopping bond... "
    [ -n "$BOND_PID" ] && kill "$BOND_PID" 2>/dev/null
    [ -n "$BOND_PID" ] && wait "$BOND_PID" 2>/dev/null || true
    echo "ok."

    echo -n "Stopping ord... "
    [ -n "$ORD_PID" ] && kill "$ORD_PID" 2>/dev/null
    [ -n "$ORD_PID" ] && wait "$ORD_PID" 2>/dev/null || true
    echo "ok."

    echo -n "Stopping bitcoind... "
    bitcoin_cli stop &>$LOG_BITCOIND || true
    [ -n "$BITCOIND_PID" ] && wait "$BITCOIND_PID" 2>/dev/null || true
    echo "ok."

    rm -rf "$TEST_DIR"
    echo "All services stopped."
}

wait_for() {
    local name=$1
    local check_cmd=$2
    local max_attempts=${3:-30}

    echo -n "Waiting for $name... "
    for i in $(seq 1 "$max_attempts"); do
        if eval "$check_cmd" &>/dev/null; then
            echo "ok."
            return 0
        fi
        sleep 1
    done

    echo " timed out after $max_attempts seconds."
    return 1
}

wait_for_start() {
    local name=$1
    local check_cmd=$2
    local max_attempts=${3:-30}

    echo -n "Starting $name... "

    for i in $(seq 1 "$max_attempts"); do
        if eval "$check_cmd" &>/dev/null; then
            echo "ok."
            return 0
        fi
        sleep 1
    done

    echo ""
    echo "$name failed to start after $max_attempts seconds."
    return 1
}

bitcoin_server() {
    exec bitcoind -regtest -txindex -datadir="$BITCOIN_DATADIR" "$@"
}

bitcoin_cli() {
    bitcoin-cli -regtest -datadir="$BITCOIN_DATADIR" "$@"
}

ord_server() {
    exec ord --regtest \
        --bitcoin-data-dir="$BITCOIN_DATADIR" \
        --data-dir="$ORD_DATADIR" \
        server "$@"
}

ord_wallet() {
    command ord --regtest \
        --bitcoin-data-dir="$BITCOIN_DATADIR" \
        --data-dir="$ORD_DATADIR" \
        wallet --server-url "http://localhost:$ORD_PORT" "$@"
}

bond_server() {
    exec "$TEST_DIR/bond" \
        --rest-listen-url "0.0.0.0:$BOND_PORT" \
        --ord-url "http://localhost:$ORD_PORT" \
        --data-dir "$BOND_DATADIR/data" \
        --start-block 0 \
        --non-interactive \
        --no-auto-index "$@"
}

bond_client() {
    curl -s "http://localhost:$BOND_PORT$1"
}

bond_sync() {
    "$TEST_DIR/bond" sync --data-dir "$BOND_DATADIR/data"
}

mine() {
    local blocks=${1:-1}
    local address=$2

    bitcoin_cli generatetoaddress "$blocks" "$address" &>$LOG_BITCOIND
}

ord_sync() {
    local height
    height=$(bitcoin_cli getblockcount)
    wait_for "ord to sync to block $height" \
        "[ \$(curl -sf http://localhost:$ORD_PORT/blockcount) -ge $height ]"
}

assert_equals() {
    local description=$1
    local expected=$2
    local actual=$3

    if [ "$actual" = "$expected" ]; then
        echo "PASS: $description"
    else
        echo "FAIL: $description"
        echo "  expected: $expected"
        echo "  actual:   $actual"
        return 1
    fi
}

start_services() {
    for cmd in bitcoind bitcoin-cli ord; do
        if ! command -v "$cmd" &>/dev/null; then
            echo "Error: $cmd is not installed."
            return 1
        fi
    done

    echo -n "Building bond... "
    (cd "$PROJECT_DIR" && go build -o "$TEST_DIR/bond" .)
    echo "ok."

    bitcoin_server &>$LOG_BITCOIND &
    BITCOIND_PID=$!
    wait_for_start "bitcoind" "bitcoin_cli getblockchaininfo" || return 1

    ord_server --http-port "$ORD_PORT" &>$LOG_ORD &
    ORD_PID=$!
    wait_for_start "ord" "curl -sf http://localhost:$ORD_PORT/status" || return 1

    bond_server &>$LOG_BOND &
    BOND_PID=$!
    wait_for_start "bond" "curl -sf http://localhost:$BOND_PORT/health" 10 || return 1

    echo "All services running."
    echo ""
}
