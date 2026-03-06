#!/usr/bin/env python3

import urllib.request
from test_setup import test_environment, ORD_PORT


def smoke_test(env) -> None:
    print("Basic health checks for bitcoind, ord, and bond.")

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


if __name__ == "__main__":
    with test_environment() as env:
        smoke_test(env)
