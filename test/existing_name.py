#!/usr/bin/env python3

import json

from test_setup import test_environment


def existing_name(env) -> None:
    print("Existing name should be indexed and not resolving without a routing.")

    # Create btcname inscription
    domain = "existing.btc"
    name_file = env.test_dir / "name.txt"
    name_file.write_text(domain)

    inscribe_out = env.ord_inscribe(name_file, capture_output=True)
    inscribe_data = json.loads(inscribe_out)
    name_id = inscribe_data["inscriptions"][0]["id"]

    # Mine and sync inscription
    env.mine_and_sync(7)

    # Verify name metadata endpoint
    name_response = env.bond_client(f"/name/{domain}", capture_output=True)
    name_data = json.loads(name_response)

    env.assert_equals(
        "name endpoint returns correct domain",
        domain,
        name_data["domain"],
    )
    env.assert_equals(
        "name endpoint returns correct ordinal id",
        name_id,
        name_data["ordinal_id"],
    )

    # Verify resolution response (no routing yet)
    resolve_response = env.bond_client(f"/resolve/{domain}", capture_output=True)
    env.assert_equals(
        "name exists but has no routing",
        '{"error":"routing not found"}',
        resolve_response,
    )


if __name__ == "__main__":
    with test_environment() as env:
        existing_name(env)
