package service

type BlockStore interface {
	GetLastIndexedBlock() (uint64, error)
	SetLastIndexedBlock(blockId uint64) error
}
