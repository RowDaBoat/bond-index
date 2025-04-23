package actions

import (
	"fmt"
	"yarr/service"
	"yarr/types"
)

type NameResolver struct {
	btcNameStore service.BtcNameStore
	routingStore service.RoutingStore
}

func NewNameResolver(btcNameStore service.BtcNameStore, routingStore service.RoutingStore) *NameResolver {
	return &NameResolver{
		btcNameStore: btcNameStore,
		routingStore: routingStore,
	}
}

func (nr *NameResolver) Resolve(name string) (string, error) {
	if name == "" {
		return "", types.ErrNameRequired
	}

	btcName, err := nr.btcNameStore.Retrieve(name)
	if err != nil {
		return "", types.NewStoreError(err)
	}

	if btcName == nil {
		return "", types.ErrNameNotFound
	}

	routing, err := nr.routingStore.Retrieve(btcName.OwnerAddress, name)
	if err != nil {
		return "", types.NewStoreError(err)
	}

	if routing == nil {
		return "", types.ErrRoutingNotFound
	}

	return fmt.Sprintf("%d: %s (Nostr: %s)", btcName.Number, btcName.OwnerAddress, routing.NostrNpub), nil
}
