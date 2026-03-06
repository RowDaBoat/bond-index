#!/usr/bin/env python3

from test_setup import test_environment
from smoke_test import smoke_test
from existing_name import existing_name
from non_existing_name import non_existing_name
from valid_routing import valid_routing
from duplicated_name import duplicated_name
from routing_on_duplicated_name import routing_on_duplicated_name
from orphan_routing import orphan_routing


def main() -> None:
    tests = [
        ("smoke_test", smoke_test),
        ("existing_name", existing_name),
        ("non_existing_name", non_existing_name),
        ("valid_routing", valid_routing),
        ("duplicated_name", duplicated_name),
        ("routing_on_duplicated_name", routing_on_duplicated_name),
        ("orphan_routing", orphan_routing),
    ]

    with test_environment() as env:
        for name, fn in tests:
            print(f"[Running {name}]")
            fn(env)
            print("")


if __name__ == "__main__":
    main()
