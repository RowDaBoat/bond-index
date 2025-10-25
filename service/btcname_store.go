package service

import "bond/types"

type BtcNameStore interface {
	Store(btcName *types.BtcName) error
	Retrieve(btcName string) (*types.BtcName, error)
	Delete(btcName string) error
}
