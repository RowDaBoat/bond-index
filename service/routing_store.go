package service

import "bond/types"

type RoutingStore interface {
	Store(name string, value *types.Routing) error
	Retrieve(name string) (*types.Routing, error)
	Delete(name string) error
}
