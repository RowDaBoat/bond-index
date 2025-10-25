package infrastructure

import (
	"bond/types"
	"fmt"
	"strings"

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
	situation := fmt.Sprintf("fetching inscription %s", id)
	url := fmt.Sprintf("%s/inscription/%s", s.ordUrl, id)
	var inscription types.Inscription
	resp, err := s.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&inscription).
		Get(url)

	switch {
	case err != nil:
		return nil, FormatError(situation, url, err)
	case resp.StatusCode() != 200:
		return nil, BuildError(situation, url, resp)
	default:
		return &inscription, nil
	}
}

func (s *Ord) FetchContent(inscriptionId string) (string, error) {
	situation := fmt.Sprintf("fetching content %s", inscriptionId)
	url := fmt.Sprintf("%s/content/%s", s.ordUrl, inscriptionId)
	resp, err := s.client.R().
		Get(url)

	switch {
	case err != nil:
		return "", FormatError(situation, url, err)
	case resp.StatusCode() != 200:
		return "", BuildError(situation, url, resp)
	default:
		return string(resp.Body()), nil
	}
}

func (s *Ord) HasContentType(inscription *types.Inscription, expectedContentType string) bool {
	return strings.HasPrefix(inscription.ContentType, expectedContentType)
}

func (s *Ord) FetchBlock(blockId uint64) (*types.Block, error) {
	var block types.Block
	situation := fmt.Sprintf("fetching block %d", blockId)
	url := fmt.Sprintf("%s/block/%d", s.ordUrl, blockId)
	resp, err := s.client.R().
		SetHeader("Accept", "application/json").
		SetResult(&block).
		Get(url)

	switch {
	case err != nil:
		return nil, FormatError(situation, url, err)
	case resp.StatusCode() != 200:
		return nil, BuildError(situation, url, resp)
	default:
		return &block, nil
	}
}

func FormatError(situation string, url string, err error) error {
	head := fmt.Sprintf("Error when %s:\n", situation)
	requestMessage := fmt.Sprintf("\tGET %s", url)
	errorMessage := fmt.Sprintf("\terror: %v", err)
	return fmt.Errorf("%s\n%s\n%s", head, requestMessage, errorMessage)
}

func BuildError(situation string, url string, response *resty.Response) error {
	head := fmt.Sprintf("Error when %s:", situation)
	requestMessage := fmt.Sprintf("\tGET %s", url)
	responseMessage := fmt.Sprintf("\tresponse: status %d", response.StatusCode())
	return fmt.Errorf("%s\n%s\n%s", head, requestMessage, responseMessage)
}
