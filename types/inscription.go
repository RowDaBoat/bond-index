package types

type Inscription struct {
	Number      int64  `json:"number"`
	Id          string `json:"id"`
	Address     string `json:"address"`
	ContentType string `json:"content_type"`
	Height      int    `json:"height"`
}
