package core

import (
	"fmt"
	"strings"

	"yarr/service"
	"yarr/types"
)

type BtcNameProcessorMonad struct {
	Ord          service.Ord
	BtcNameStore service.BtcNameStore
}

func (p *BtcNameProcessorMonad) Process(inscription *types.Inscription) (string, error) {
	//TODO write as a monad using flat map
	// Early return if not text/plain
	if !p.Ord.HasContentType(inscription, "text/plain") {
		return "", nil
	}

	// Create the processing pipeline
	content, err := p.Ord.FetchContent(inscription.Id)
	if err != nil {
		return "", err
	}

	// Check if it's a .btc domain
	if !strings.HasSuffix(content, ".btc") {
		return "", nil
	}

	// Create BtcName and handle storage
	newBtcName := types.BtcName{
		Domain:       content,
		Number:       inscription.Number,
		Id:           inscription.Id,
		OwnerAddress: inscription.Address,
	}

	oldBtcName, err := p.BtcNameStore.Retrieve(newBtcName.Domain)
	if err != nil {
		return "", err
	}

	if oldBtcName == nil || newBtcName.Number < oldBtcName.Number {
		if err := p.BtcNameStore.Store(&newBtcName); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf(
		"  Id:\t\t%s\n  Number:\t%d\n  OwnerAddress:\t%s\n  Domain:\t%s",
		newBtcName.Id, newBtcName.Number, newBtcName.OwnerAddress, newBtcName.Domain,
	), nil
}
