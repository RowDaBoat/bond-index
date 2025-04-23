package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"yarr/types"
)

type Ord struct {
	blocks       map[uint64]types.Block
	inscriptions map[string]types.Inscription
	contents     map[string]string
	height       uint64
}

type OrdsJson struct {
	Blocks       []types.Block       `json:"blocks"`
	Inscriptions []types.Inscription `json:"inscriptions"`
	Contents     map[string]string   `json:"contents"`
	Height       uint64              `json:"height"`
}

func NewOrd(dataDir string) *Ord {
	ord := &Ord{
		blocks:       make(map[uint64]types.Block),
		inscriptions: make(map[string]types.Inscription),
		contents:     make(map[string]string),
		height:       0,
	}

	if err := ord.loadData(dataDir); err != nil {
		fmt.Println("Error: failed to load ord data: %w", err)
		os.Exit(1)
	}

	return ord
}

func (o *Ord) FetchBlock(blockId uint64) (*types.Block, error) {
	if blockId > o.height {
		return nil, nil
	}

	if block, exists := o.blocks[blockId]; exists {
		return &block, nil
	}

	return &types.Block{
		Height:       blockId,
		BestHeight:   o.height,
		Inscriptions: []string{},
	}, nil
}

func (o *Ord) FetchInscription(id string) (*types.Inscription, error) {
	inscription, exists := o.inscriptions[id]
	if !exists {
		return nil, nil
	}
	return &inscription, nil
}

func (o *Ord) FetchContent(inscriptionId string) (string, error) {
	content, exists := o.contents[inscriptionId]
	if !exists {
		return "", nil
	}
	return content, nil
}

func (o *Ord) HasContentType(inscription *types.Inscription, expectedContentType string) bool {
	return inscription.ContentType == expectedContentType
}

func (o *Ord) loadData(dataDir string) error {
	data, err := os.ReadFile(filepath.Join(dataDir, "ords.json"))
	if err != nil {
		return fmt.Errorf("failed to read ords.json: %w", err)
	}

	var jsonData OrdsJson
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse oesa.json: %w", err)
	}

	for _, block := range jsonData.Blocks {
		o.blocks[block.Height] = block
	}

	for _, inscription := range jsonData.Inscriptions {
		o.inscriptions[inscription.Id] = inscription
	}

	o.contents = jsonData.Contents
	o.height = jsonData.Height

	return nil
}
