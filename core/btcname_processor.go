package core

import (
	"fmt"
	"strings"

	"bond/service"
	"bond/types"
)

type BtcNameProcessor struct {
	Ord          service.Ord
	BtcNameStore service.BtcNameStore
}

func (p *BtcNameProcessor) Process(inscription *types.Inscription) (string, error) {
	//TODO: parsing errors should not be reported, since the blockchain can contain anything really.
	//TODO: network errors on the other hand should be reported and retried.
	if !p.Ord.HasContentType(inscription, "text/plain") {
		return "", nil
	}

	content, err := p.Ord.FetchContent(inscription.Id)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(content, ".btc") {
		return "", nil
	}

	newBtcName := types.BtcName{
		Domain:       content,
		Number:       inscription.Number,
		Id:           inscription.Id,
		OwnerAddress: inscription.Address,
	}

	oldBtcName, err := p.BtcNameStore.Retrieve(newBtcName.Domain)
	if err == nil && oldBtcName == nil {
		p.BtcNameStore.Store(&newBtcName)
	} else if err == nil && newBtcName.Number < oldBtcName.Number {
		p.BtcNameStore.Store(&newBtcName)
	} else if err != nil {
		return "", err
	}

	return fmt.Sprintf(
			"  Id:\t\t%s\n  Number:\t%d\n  OwnerAddress:\t%s\n  Domain:\t%s",
			newBtcName.Id, newBtcName.Number, newBtcName.OwnerAddress, newBtcName.Domain,
		),
		nil
}
