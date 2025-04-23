package infrastructure

import (
	"fmt"
	"strings"
	"yarr/types"

	"github.com/go-resty/resty/v2"
)

type Ord struct {
	client *resty.Client
	ordUrl string
}

func NewOrd(client *resty.Client, ordUrl string) *Ord {
	return &Ord{
		client: client,
		ordUrl: ordUrl,
	}
}

func (s *Ord) FetchInscription(id string) (*types.Inscription, error) {
	var inscription types.Inscription
	resp, err := s.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&inscription).
		Get(fmt.Sprintf("http://%s/inscription/%s", s.ordUrl, id))

	switch {
	case err != nil:
		return nil, fmt.Errorf("error: %v", err)
	case resp.StatusCode() != 200:
		return nil, fmt.Errorf("error: http status %d", resp.StatusCode())
	default:
		return &inscription, nil
	}
}

func (s *Ord) FetchContent(inscriptionId string) (string, error) {
	resp, err := s.client.R().
		Get(fmt.Sprintf("http://%s/content/%s", s.ordUrl, inscriptionId))

	switch {
	case err != nil:
		return "", fmt.Errorf("error: %v", err)
	case resp.StatusCode() != 200:
		return "", fmt.Errorf("error: http status %d", resp.StatusCode())
	default:
		return string(resp.Body()), nil
	}
}

func (s *Ord) HasContentType(inscription *types.Inscription, expectedContentType string) bool {
	return strings.HasPrefix(inscription.ContentType, expectedContentType)
}

func (s *Ord) FetchBlock(blockId uint64) (*types.Block, error) {
	var block types.Block
	resp, err := s.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&block).
		Get(fmt.Sprintf("http://%s/block/%d", s.ordUrl, blockId))

	switch {
	case err != nil:
		return nil, fmt.Errorf("error: %v", err)
	case resp.StatusCode() != 200:
		return nil, fmt.Errorf("error: http status %d", resp.StatusCode())
	default:
		return &block, nil
	}
}
