package service

import "yarr/types"

type RoutingStore interface {
	Store(address string, name string, value *types.Routing) error
	Retrieve(address string, name string) (*types.Routing, error)
	Delete(address string, name string) error
}
