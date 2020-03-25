package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type AddressTransaction struct {
	Hash    string       `json:"hash"`
	Txid    string       `json:"txid"`
	Height  uint64       `json:"height"`
	Index   uint         `json:"index"`
	Time    time.Time    `json:"time, omitempty"`
	Type    TransferType `json:"type"`
	Input   uint64       `json:"input"`
	Output  uint64       `json:"output"`
	Total   int64        `json:"total"`
	Balance uint64       `json:"balance"`
	Cold    bool         `json:"cold"`
}

func (a *AddressTransaction) Slug() string {
	return slug.Make(fmt.Sprintf("addresstx-%s-%s-%t", a.Hash, a.Txid, a.Cold))
}
