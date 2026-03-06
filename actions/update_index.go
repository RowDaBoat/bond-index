package actions

import (
	"fmt"
	"os"
	"time"

	"bond/core"
	"bond/service"
	"bond/types"

	"golang.org/x/term"
)

type SyncRequest chan uint64

type IndexUpdater struct {
	ord            service.Ord
	btcNameStore   service.BtcNameStore
	routingStore   service.RoutingStore
	blockStore     service.BlockStore
	processors     []core.InscriptionProcessor
	startBlock     uint64
	nonInteractive bool
	noAutoIndex    bool
	syncRequests   chan SyncRequest
}

func NewIndexUpdater(ord service.Ord, btcNameStore service.BtcNameStore, routingStore service.RoutingStore, blockStore service.BlockStore, startBlock uint64, nonInteractive bool, noAutoIndex bool) *IndexUpdater {
	processors := []core.InscriptionProcessor{
		&core.BtcNameProcessor{Ord: ord, BtcNameStore: btcNameStore},
		&core.RoutingProcessor{Ord: ord, RoutingStore: routingStore, BtcNameStore: btcNameStore},
	}

	return &IndexUpdater{
		ord:            ord,
		btcNameStore:   btcNameStore,
		routingStore:   routingStore,
		blockStore:     blockStore,
		processors:     processors,
		startBlock:     startBlock,
		nonInteractive: nonInteractive,
		noAutoIndex:    noAutoIndex,
		syncRequests:   make(chan SyncRequest, 10),
	}
}

func (iu *IndexUpdater) SyncRequests() chan SyncRequest {
	return iu.syncRequests
}

func (iu *IndexUpdater) Start(startBlock uint64) {
	lastIndexedBlock, err := iu.blockStore.GetLastIndexedBlock()
	if err != nil {
		fmt.Printf("Error getting last indexed block: %v\n", err)
		lastIndexedBlock = 0
	}

	blockId := max(lastIndexedBlock, startBlock)
	backoff := time.Second
	var result ProcessBlockResult

	for {
		blockId, result = iu.processBlock(blockId)
		if result == Success {
			if err := iu.blockStore.SetLastIndexedBlock(blockId); err != nil {
				fmt.Printf("Error setting last indexed block: %v\n", err)
			}
		}

		if result != Success {
			iu.respondToSyncRequests(blockId)
			backoff = iu.wait(backoff, result)
		}
	}
}

func (iu *IndexUpdater) respondToSyncRequests(blockId uint64) {
	for {
		select {
		case req := <-iu.syncRequests:
			req <- blockId
		default:
			return
		}
	}
}

type ProcessBlockResult int

const (
	Success ProcessBlockResult = iota
	Backoff
	MaxBackoff
)

func (iu *IndexUpdater) processBlock(blockId uint64) (uint64, ProcessBlockResult) {
	block, result, err := iu.getNextBlock(blockId)
	if err != nil || block == nil {
		return blockId, result
	}

	iu.showBlock(block)
	iu.processInscriptions(block)

	return blockId + 1, result
}

func (iu *IndexUpdater) getNextBlock(blockId uint64) (*types.Block, ProcessBlockResult, error) {
	block, err := iu.ord.FetchBlock(blockId)
	if err != nil {
		fmt.Printf("%v", err)
		return nil, Backoff, err
	}

	if block == nil || block.BestHeight-6 < blockId {
		fmt.Printf("Reached top of chain: %d/%d\n", blockId, block.BestHeight)
		return nil, MaxBackoff, nil
	}

	return block, Success, nil
}

func (iu *IndexUpdater) showBlock(block *types.Block) {
	bestHeight := block.BestHeight
	start := float32(block.Height - iu.startBlock)
	total := float32(bestHeight - iu.startBlock)
	percentage := start / total * 100
	percentageString := fmt.Sprintf("%.2f%%", percentage)
	inscriptionCount := len(block.Inscriptions)

	if !iu.nonInteractive && term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf(
			"\033[K\rBlock %d/%d %s %d inscriptions:",
			block.Height, bestHeight, percentageString, inscriptionCount,
		)
	} else {
		fmt.Printf(
			"Block %d/%d %s %d inscriptions\n",
			block.Height, bestHeight, percentageString, inscriptionCount,
		)
	}
}

func (iu *IndexUpdater) processInscriptions(block *types.Block) {
	for _, inscription := range block.Inscriptions {
		result, err := iu.processInscription(inscription)
		if err != nil {
			fmt.Printf("\n  Error processing inscription %s: %v\n", inscription, err)
		} else if result != "" {
			fmt.Printf("\n%s\n", result)
		}
	}
}

func (iu *IndexUpdater) wait(backoff time.Duration, result ProcessBlockResult) time.Duration {
	if iu.noAutoIndex {
		req := <-iu.syncRequests
		iu.syncRequests <- req
		return backoff
	}

	if result == MaxBackoff {
		backoff = time.Minute * 5
	}

	fmt.Printf("Waiting %v...\n", backoff)

	select {
	case req := <-iu.syncRequests:
		fmt.Printf("Sync requested, waking up.\n")
		iu.syncRequests <- req
	case <-time.After(backoff):
	}

	if result == MaxBackoff {
		return backoff
	}

	return min(backoff*2, time.Minute*5)
}

func (iu *IndexUpdater) processInscription(id string) (string, error) {
	inscription, err := iu.ord.FetchInscription(id)

	if err != nil {
		return "", err
	}

	if inscription == nil {
		return "", nil
	}

	for _, processor := range iu.processors {
		result, err := processor.Process(inscription)

		if err != nil {
			return "", err
		}

		if result != "" {
			return result, nil
		}
	}

	return "", nil
}
