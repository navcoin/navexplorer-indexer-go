package explorer

import "time"

type AddressTransaction struct {
	//MetaData MetaData `json:"-"`

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
