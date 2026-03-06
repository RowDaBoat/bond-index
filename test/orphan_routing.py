#!/usr/bin/env python3

import json

from test_setup import test_environment


def orphan_routing(env) -> None:
    print("Orphan routing inscriptions should be ignored and not break the service.")

    # Create a routing inscription JSON with no parent inscription
    routing_file = env.test_dir / "orphan_routing.json"
    routing_file.write_text(
        json.dumps(
            {
                "p": "btcname",
                "op": "routing",
                "nostr_npub": "npub1orphan",
                "nostr_relays": ["wss://relay.orphan.example.com"],
            }
        )
    )

    # Inscribe routing without a parent; this should be ignored by the indexer
    env.ord_inscribe(routing_file, capture_output=False)

    # Mine and sync; services should continue running normally
    env.mine_and_sync(7)

    # Sanity check: bond service is still healthy
    def check_bond_health() -> None:
        env.bond_client("/health", capture_output=True)

    env.assert_success("bond stays healthy after orphan routing", check_bond_health)


if __name__ == "__main__":
    with test_environment() as env:
        orphan_routing(env)
