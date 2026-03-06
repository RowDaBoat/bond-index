package core

type RoutingInscription struct {
	Protocol    string   `json:"p"`
	Operation   string   `json:"op"`
	NostrNpub   string   `json:"nostr_npub"`
	NostrRelays []string `json:"nostr_relays"`
}
