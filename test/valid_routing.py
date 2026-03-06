#!/usr/bin/env python3

import json

from test_setup import test_environment


def valid_routing(env) -> None:
    print("A valid routing as child of a name should resolve to nostr npub and relays.")

    # Create btcname inscription at a dedicated address
    domain = "namewithrouting.btc"
    name_file = env.test_dir / "name.txt"
    name_file.write_text(domain)

    name_addr_out = env.ord_wallet("receive", capture_output=True)
    name_addr_data = json.loads(name_addr_out)
    name_address = name_addr_data["addresses"][0]

    inscribe_out = env.ord_inscribe(
        name_file,
        destination=name_address,
        capture_output=True,
    )
    inscribe_data = json.loads(inscribe_out)
    name_inscription_id = inscribe_data["inscriptions"][0]["id"]

    env.mine(7)
    env.ord_sync()

    # Create routing inscription as a child of the name inscription
    routing_file = env.test_dir / "routing.json"
    routing_file.write_text(
        json.dumps(
            {
                "p": "btcname",
                "op": "routing",
                "nostr_npub": "npub1testpubkey",
                "nostr_relays": ["wss://relay.example.com"],
            }
        )
    )

    env.ord_inscribe(
        routing_file,
        parent=name_inscription_id,
        capture_output=False,
    )

    env.mine_and_sync(7)

    # Verify resolution uses routing
    response = env.bond_client(f"/resolve/{domain}", capture_output=True)
    data = json.loads(response)

    env.assert_equals(
        "routing returns expected nostr npub",
        "npub1testpubkey",
        data["nostr_npub"],
    )
    env.assert_equals(
        "routing returns expected relay",
        "wss://relay.example.com",
        data["nostr_relays"][0],
    )


if __name__ == "__main__":
    with test_environment() as env:
        valid_routing(env)

