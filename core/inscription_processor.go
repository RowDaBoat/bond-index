package core

import (
	"yarr/types"
)

type InscriptionProcessor interface {
	Process(inscription *types.Inscription) (string, error)
}
