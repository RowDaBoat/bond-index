#!/usr/bin/env python3

import json

from test_setup import test_environment


def duplicated_name(env) -> None:
    print("Duplicated names should be ignored.")

    # Create and fund wallet
    env.ord_wallet("create")
    receive_out = env.ord_wallet("receive", capture_output=True)
    receive_data = json.loads(receive_out)
    wallet_address = receive_data["addresses"][0]

    env.mine(101, wallet_address)
    env.ord_sync()

    # First btcname inscription for the same domain
    domain = "dup.btc"
    name_file = env.test_dir / "name.txt"
    name_file.write_text(domain)

    first_inscribe_out = env.ord_inscribe(name_file, capture_output=True)
    first_id = json.loads(first_inscribe_out)["inscriptions"][0]["id"]

    env.mine(7, wallet_address)
    env.ord_sync()
    env.bond_sync()

    # Verify name metadata reflects the first inscription
    name_response = env.bond_client(f"/name/{domain}", capture_output=True)
    name_data = json.loads(name_response)

    env.assert_equals(
        "name endpoint returns correct domain for first inscription",
        domain,
        name_data["domain"],
    )
    env.assert_equals(
        "name endpoint returns first ordinal id",
        first_id,
        name_data["ordinal_id"],
    )

    # Second btcname inscription with the same domain
    name_file.write_text(domain)
    second_inscribe_out = env.ord_inscribe(name_file, capture_output=True)
    second_id = json.loads(second_inscribe_out)["inscriptions"][0]["id"]

    env.mine(7, wallet_address)
    env.ord_sync()
    env.bond_sync()

    # Verify duplicated name is ignored and first ordinal id is still used
    name_response = env.bond_client(f"/name/{domain}", capture_output=True)
    name_data = json.loads(name_response)

    env.assert_equals(
        "duplicated name keeps first ordinal id",
        first_id,
        name_data["ordinal_id"],
    )

    print("All assertions passed.")


if __name__ == "__main__":
    with test_environment() as env:
        duplicated_name(env)
