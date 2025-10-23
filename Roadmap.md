First POC
- [x] Handle negative ordinals
- [x] Store first occurrence of a BTC Name in an index
- [x] Store last occurrence of a Address+BTCName Routing in an index
- [x] Dockerize
- [x] Build docker image on CI
- [x] Properly handle reaching the end of the chain
- [x] Build the index
- [ ] Implement a DNS plugin for CoreDNS
  - [ ] Build CoreDNS from source
  - [ ] Install and configure CoreDNS as a regular DNS server
  - [ ] Implement a basic plugin
  - [ ] Catch `.btc` requests
  - [ ] Query the `yarr` API to resolve `.btc` domains to Nostr relays+npubs
  - [ ] Respond with the IP address from the Nostr relays+npub
- [ ] Handle ownership of BTC Names and Routings

Golden Milestone
- [ ] Setup a public DNS server to resolve `.btc` domains
- [ ] Offer a service to register a `.btc` domain for sats

Improvements
- [ ] Review error handling
- [ ] Download a synced ordinals index
- [ ] Use a pruned Bitcion node
