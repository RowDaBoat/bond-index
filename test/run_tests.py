#!/usr/bin/env python3

from test_setup import test_environment


def main() -> None:
    with test_environment() as env:
        return


if __name__ == "__main__":
    main()
