package core

import (
	"bond/types"
)

type InscriptionProcessor interface {
	Process(inscription *types.Inscription) (string, error)
}
