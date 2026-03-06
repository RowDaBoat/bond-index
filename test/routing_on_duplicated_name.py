#!/usr/bin/env python3

import json

from test_setup import test_environment


def routing_on_duplicated_name(env) -> None:
    print("Routings attached to duplicated names should be ignored.")

    name_file = env.test_dir / "name.txt"

    # First btcname inscription for dup.btc
    domain = "duprouting.btc"
    name_file.write_text(domain)
    first_inscribe_out = env.ord_inscribe(name_file, capture_output=True)
    first_id = json.loads(first_inscribe_out)["inscriptions"][0]["id"]

    # Confirm first inscription so wallet has UTXOs for the second
    env.mine(7)
    env.ord_sync()

    # Second btcname inscription for dup.btc (duplicate name)
    name_file.write_text(domain)
    second_inscribe_out = env.ord_inscribe(name_file, capture_output=True)
    second_id = json.loads(second_inscribe_out)["inscriptions"][0]["id"]

    # Mine inscriptions and sync
    env.mine_and_sync(7)

    # Create routing inscription as a child of the original name
    routing_file = env.test_dir / "routing.json"
    routing_file.write_text(
        json.dumps(
            {
                "p": "btcname",
                "op": "routing",
                "nostr_npub": "npub1original",
                "nostr_relays": ["wss://relay.original.example.com"],
            }
        )
    )
    env.ord_inscribe(
        routing_file,
        parent=first_id,
        capture_output=False,
    )

    # Create routing inscription as a child of the duplicated name
    routing_file.write_text(
        json.dumps(
            {
                "p": "btcname",
                "op": "routing",
                "nostr_npub": "npub1duplicated",
                "nostr_relays": ["wss://relay.duplicated.example.com"],
            }
        )
    )
    env.ord_inscribe(
        routing_file,
        parent=second_id,
        capture_output=False,
    )

    # Mine inscriptions and sync
    env.mine_and_sync(7)

    # Resolve should respond using the routing on the original name
    response = env.bond_client(f"/resolve/{domain}", capture_output=True)
    data = json.loads(response)

    env.assert_equals(
        "resolve responds using the routing on the original name",
        "npub1original",
        data["nostr_npub"],
    )
    env.assert_equals(
        "resolve responds using the relays on the original name",
        "wss://relay.original.example.com",
        data["nostr_relays"][0],
    )


if __name__ == "__main__":
    with test_environment() as env:
        routing_on_duplicated_name(env)
