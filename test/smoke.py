#!/usr/bin/env python3

from test_setup import test_environment, ORD_PORT
import urllib.request


with test_environment() as env:
    print("Smoke test: services started and responding")
    print()

    # bitcoind
    env.assert_success(
        "bitcoind responds to getblockchaininfo",
        env.bitcoin_cli,
        "getblockchaininfo",
        capture_output=False,
    )

    # ord
    def check_ord_status() -> None:
        with urllib.request.urlopen(f"http://localhost:{ORD_PORT}/status", timeout=2):
            pass

    env.assert_success("ord status endpoint is reachable", check_ord_status)

    # bond
    def check_bond_health() -> None:
        env.bond_client("/health", capture_output=True)

    env.assert_success("bond health endpoint is reachable", check_bond_health)

    print()
    print("All smoke test assertions passed.")
