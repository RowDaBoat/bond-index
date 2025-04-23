package service

import "yarr/types"

type Ord interface {
	FetchInscription(id string) (*types.Inscription, error)
	FetchContent(inscriptionId string) (string, error)
	HasContentType(inscription *types.Inscription, expectedContentType string) bool
	FetchBlock(blockId uint64) (*types.Block, error)
}
