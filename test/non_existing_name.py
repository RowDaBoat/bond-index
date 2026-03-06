#!/usr/bin/env python3

from test_setup import test_environment


def non_existing_name(env) -> None:
    print("A non-existing name should return a not-found error.")

    domain = "nonexisting.btc"
    response = env.bond_client(f"/resolve/{domain}", capture_output=True)
    env.assert_equals("name does not exist", '{"error":"name not found"}', response)


if __name__ == "__main__":
    with test_environment() as env:
        non_existing_name(env)
