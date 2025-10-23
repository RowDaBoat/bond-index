# yarr

[![Go Version](https://img.shields.io/badge/Go-1.23.5+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com)


## Overview
`yarr` decentralizes domain name resolution using **Bitcoin** as the domain registration protocol, and [**Nostr**](https://github.com/nostr-protocol/nostr) for dynamic IP address lookup.

In practice, that means that `yarr` resolves [`.btc` domain names](https://docs.btcname.id/docs) such as `godofthunder.btc` or `rowboto.btc` to IP addresses posted in **Nostr** notes.


## How It Works
### Ordinals
While Ordinals are a source of controversy, they are here to stay. Their censorship resistance, ability to be uniquely minted, ability to store arbitrary data, ability to be identified, and ability to be traded make them a great way to solve a real-world problem: decentralizing the ownership of domain names on a truly immutable registry.


### Nostr
**Nostr** is a decentralized social network protocol designed for censorship resistance and identification of users through their public keys while maintaining anonymity. **Nostr** relays keep the costs of uploading notes cheap, which is ideal for maintaining records of ever-changing addresses to resolve domain names to.


### Inscriptions and Notes
**The domain name inscription**: The domain name is just a string ending in `.btc` inscribed in an ordinal (e.g., `rowboto.btc`). Its owner is the current owner of the ordinal. `yarr` indexes these inscriptions, paying attention only to the current owner of the first occurrence of each domain name.

**The routing inscription**: The routing is an inscribed JSON string owned by the same domain name owner, which contains their **Nostr** **npub** and the relays to use to find the IP to resolve the domain name to.
```json
{
    "p": "btcname",
    "op": "routing",
    "name": "rowboto.btc",
    "nostr_npub": "npub1xwjap7x26602fze5epv5lpf3v5hgefjzpy49p9ulyx2sdam8jl3qdrfljs",
    "nostr_relays": ["wss://relay.nostr.band", "wss://nos.lol", "wss://relay.damus.io"]
}
```
`yarr` always indexes the last routing inscription of the same owner, thus allowing for updates in case the owner wants to change their **npub** (e.g., in case of loss) or update the relays.

**The resolution note**: The resolution is a note with the IP to resolve the domain name to, posted to **Nostr** by the **npub** on the **routing inscription** (either manually or via an automated script). This closes the loop of decentralized domain name resolution.


## State of the `yarr`t
Currently, `yarr` is just a proof of concept. It can effectively search **domain name** and **routing** inscriptions through the blockchain and store them in a database.

Resolving to **Nostr** has not been implemented yet. To achieve this, `harr`, a plugin for [CoreDNS](https://coredns.io/), will be developed. When asked for a `.btc` domain, `harr` will query `yarr` to find the proper owner's **npub**, and then query the **Nostr** relays to find the IP to resolve the domain name to.

Note that the current solution allows the owner of a domain to store their private keys in a cold wallet. Signing is only needed when transferring the **domain name inscription** or when writing a new **routing inscription**. On the other hand, the **Nostr** **nsec** will be needed each time a **resolution note** is posted.


## Diagram
```
 ____                       .---------.                              .------.
(. _.) --- DNS Request ---> |         | ---- Who's rowboto.btc? ---> |      |
 /|\       rowboto.btc      |         |                              | yarr |
  /\                        |         | <---- It's npub1xwja... ---- |      |
 User <-.                   | CoreDNS |       with relay list        '------'
        |                   |  harr   |                              .-------------.
        |                   |         | -- What's npub1xwja...'s --> |             |
        |                   |         |    latest IP?                | Nostr Relay |
        |                   |         |                              |             |
        '-- DNS Response -- |         | <---- It's 121.99.9.12 ----- |             |
            121.99.9.12     '---------'                              '-------------'
```
