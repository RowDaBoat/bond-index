# Bond Index

[![Go Version](https://img.shields.io/badge/Go-1.23.5+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com)


## Overview
`bond` is **Bitcoin**, **Ordinals**, and **Nostr**, into **DNS**. It decentralizes domain name resolution using **Bitcoin** and **Ordinals** as the domain registration protocol, and [**Nostr**](https://github.com/nostr-protocol/nostr) for dynamic IP address lookup.

In practice, that means that `bond` resolves [`.btc` domain names](https://docs.btcname.id/docs) such as `godofthunder.btc` or `rowboto.btc` to IP addresses posted in **Nostr** notes.
This project is paired with a [CoreDNS plugin](http://github.com/RowDaBoat/bond-coredns) that implements the actual DNS resolution.

> **Warning**: The current implementation is a **Proof of Concept** and is not production ready. Do not use it on mainnet **by any means**.


## Indexer
`bond-index` has the inscription indexer that routes domain names to IPs obtained from Nostr, it offers a REST API for the resolution. The DNS protocol is handled by the [`bond-coredns`](https://github.com/RowDaBoat/bond-coredns) plugin for CoreDNS.


## How It Works
### Ordinals
While Ordinals are a source of controversy, they are here to stay. Their censorship resistance, ability to be uniquely minted, ability to store arbitrary data, ability to be identified, and ability to be traded, make them a great way to solve a real-world problem: decentralizing the ownership of domain names on a truly immutable registry.


### Nostr
**Nostr** is a decentralized social network protocol designed for censorship resistance and identification of users through their public keys while maintaining anonymity. **Nostr** relays keep the costs of uploading notes cheap, which is ideal for maintaining records of ever-changing addresses to resolve domain names to.


### Inscriptions and Notes
**The domain name inscription**: The domain name is just a string ending in `.btc` inscribed in an ordinal (e.g., `rowboto.btc`). Its owner is the current owner of the ordinal. `bond` indexes these inscriptions, paying attention only to the current owner of the first occurrence of each domain name.

**The routing inscription**: The routing is an inscribed JSON string owned [TODO: owned or inscribed?] by the same domain name owner, which contains their **Nostr** **npub** and the relays to use to find the IP to resolve the domain name to.
```json
{
    "p": "btcname",
    "op": "routing",
    "name": "rowboto.btc",
    "nostr_npub": "npub1xwjap7x26602fze5epv5lpf3v5hgefjzpy49p9ulyx2sdam8jl3qdrfljs",
    "nostr_relays": ["wss://relay.nostr.band", "wss://nos.lol", "wss://relay.damus.io"]
}
```
`bond` always indexes the last routing inscription of the same owner, thus allowing for updates in case the owner wants to change their **npub** (e.g., in case of loss) or update the relays.

**The resolution note**: The resolution is a note with the IP to resolve the domain name to, posted to **Nostr** by the **npub** on the **routing inscription** (either manually or via an automated script). This closes the loop of decentralized domain name resolution.


## State of the art
Although `bond` can resolve inscribed `.btc` domain names to IPs in Nostr notes via REST queries, it is currently just a proof of concept.

Note that the current solution allows the owner of a domain to store their private keys in a cold wallet. Signing is only needed when transferring the **domain name inscription** or when writing a new **routing inscription**. On the other hand, the **Nostr** **nsec** will be needed each time a **resolution note** is posted.

> **Warning**: While it's ok to get your .btcname domain registered, **do not rush to create a routing inscription for it**. The current implementation checks the ownership of domain name and routing inscriptions by checking they belong to the same address. This leaves wide open the possibility of a routing attack, where the attacker can just gift a routing inscription to the same address as the name inscription, and override the current routing. This can be solved in multiple ways, some of the solutions are: requiring the routing to be signed by the owner of the btcname, or requiring that the routing is inscribed in the same sat as the name. More solutions will be explored later in development.

## Diagram
```
 ____                       .-------------.                              .--------------.
(. _.) --- DNS Request ---> |             | ---- Who's rowboto.btc? ---> |              |
 /|\       rowboto.btc      |             |                              | bond service |
  /\                        |             | <---- It's npub1xwja... ---- |              |
 User <-.                   |   CoreDNS   |       with relay list        '--------------'
        |                   | bond plugin |                              .-------------.
        |                   |             | -- What's npub1xwja...'s --> |             |
        |                   |             |    latest IP?                | Nostr Relay |
        |                   |             |                              |             |
        '-- DNS Response -- |             | <---- It's 121.99.9.12 ----- |             |
            121.99.9.12     '-------------'                              '-------------'
```
