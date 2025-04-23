package types

type Block struct {
	Height       uint64   `json:"height"`
	BestHeight   uint64   `json:"best_height"`
	Inscriptions []string `json:"inscriptions"`
}
