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
	BtcNameStore service.BtcNameStore
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
		// Ignore malformed routing inscriptions.
		return "", nil
	}

	if routingInscription.Protocol != "btcname" {
		return "", nil
	}

	if routingInscription.Operation != "routing" {
		return "", nil
	}

	if len(inscription.Parents) == 0 {
		return "", nil
	}

	parentId := inscription.Parents[0]
	nameInscription, err := p.Ord.FetchInscription(parentId)
	if err != nil {
		// Network / API errors should be surfaced and retried.
		return "", err
	}

	if nameInscription == nil {
		// If the parent cannot be found, ignore this routing inscription.
		return "", nil
	}

	// Parent must be a valid btcname inscription (text/plain ending with .btc).
	if !p.Ord.HasContentType(nameInscription, "text/plain") {
		return "", nil
	}

	domainName, err := p.Ord.FetchContent(nameInscription.Id)
	if err != nil {
		// Network / API errors should be surfaced and retried.
		return "", err
	}

	if !strings.HasSuffix(domainName, ".btc") {
		// Parent is not a valid btcname inscription.
		return "", nil
	}

	// Only accept routings whose parent inscription is the original
	// btcname inscription for this domain. Duplicate btcname inscriptions
	// should not have routings indexed.
	btcName, err := p.BtcNameStore.Retrieve(domainName)
	if err != nil {
		// Storage errors should be surfaced and retried.
		return "", err
	}

	if btcName == nil {
		// No canonical btcname recorded for this domain; ignore routing.
		return "", nil
	}

	if btcName.Id != nameInscription.Id {
		// Parent is not the original name inscription; ignore routing.
		return "", nil
	}

	routing := &types.Routing{
		Id:          inscription.Id,
		Domain:      domainName,
		NostrNpub:   routingInscription.NostrNpub,
		NostrRelays: routingInscription.NostrRelays,
	}

	if err := p.RoutingStore.Store(domainName, routing); err != nil {
		return "", err
	}

	return fmt.Sprintf(
			"  Id:\t%s\n  Number:\t%d\n  Domain:\t%s\n  Nostr npub:\t%s",
			inscription.Id, inscription.Number, routing.Domain, routing.NostrNpub,
		),
		nil
}
