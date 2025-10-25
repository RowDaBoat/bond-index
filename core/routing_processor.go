package core

import (
	"bond/service"
	"bond/types"
	"encoding/json"
	"fmt"
	"strings"
)

type RoutingProcessor struct {
	Ord          service.Ord
	RoutingStore service.RoutingStore
}

func (p *RoutingProcessor) Process(inscription *types.Inscription) (string, error) {
	//TODO: parsing errors should not be reported, since the blockchain can contain anything really.
	//TODO: network errors on the other hand should be reported and retried.
	if !p.Ord.HasContentType(inscription, "application/json") {
		return "", nil
	}

	content, err := p.Ord.FetchContent(inscription.Id)
	if err != nil {
		return "", err
	}

	var routingInscription RoutingInscription
	if err := json.Unmarshal([]byte(content), &routingInscription); err != nil {
		return "", err
	}

	if routingInscription.Protocol != "btcname" {
		return "", nil
	}

	if routingInscription.Operation != "routing" {
		return "", nil
	}

	if !strings.HasSuffix(routingInscription.Name, ".btc") {
		return "", nil
	}

	routing := &types.Routing{
		Address:     inscription.Address,
		Domain:      routingInscription.Name,
		NostrNpub:   routingInscription.NostrNpub,
		NostrRelays: routingInscription.NostrRelays,
	}

	if err := p.RoutingStore.Store(inscription.Address, routing.Domain, routing); err != nil {
		return "", err
	}

	return fmt.Sprintf(
			"  Id:\t%s\n  Number:\t%d\n  OwnerAddress:\t%s\n  Domain:\t%s\n  Nostr npub:\t%s",
			inscription.Id, inscription.Number, inscription.Address, routing.Domain, routing.NostrNpub,
		),
		nil
}
