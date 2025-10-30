package actions

import (
	"bond/service"
	"bond/types"
)

type NameResolver struct {
	btcNameStore service.BtcNameStore
	routingStore service.RoutingStore
}

type NameResolution struct {
	NostrNpub   string   `json:"npub"`
	NostrRelays []string `json:"relays"`
}

func NewNameResolver(btcNameStore service.BtcNameStore, routingStore service.RoutingStore) *NameResolver {
	return &NameResolver{
		btcNameStore: btcNameStore,
		routingStore: routingStore,
	}
}

func (nr *NameResolver) Resolve(name string) (*NameResolution, error) {
	if name == "" {
		return nil, types.ErrNameRequired
	}

	btcName, err := nr.btcNameStore.Retrieve(name)
	if err != nil {
		return nil, types.ErrNameNotFound
	}

	if btcName == nil {
		return nil, types.ErrNameNotFound
	}

	routing, err := nr.routingStore.Retrieve(btcName.OwnerAddress, name)
	if err != nil {
		return nil, types.ErrRoutingNotFound
	}

	return &NameResolution{
		NostrNpub:   routing.NostrNpub,
		NostrRelays: routing.NostrRelays,
	}, nil
}
