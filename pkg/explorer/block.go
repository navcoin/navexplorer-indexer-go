package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"time"
)

type RawBlock struct {
	Hash              string    `json:"hash"`
	Confirmations     uint64    `json:"confirmations"`
	StrippedSize      uint64    `json:"strippedsize"`
	Size              uint64    `json:"size"`
	Weight            uint64    `json:"weight"`
	Height            uint64    `json:"height"`
	Version           uint32    `json:"version"`
	VersionHex        string    `json:"versionHex"`
	Merkleroot        string    `json:"merkleroot"`
	Tx                []string  `json:"tx"`
	Time              time.Time `json:"time"`
	MedianTime        time.Time `json:"mediantime"`
	Nonce             uint64    `json:"nonce"`
	Bits              string    `json:"bits"`
	Difficulty        string    `json:"difficulty"`
	Chainwork         string    `json:"chainwork,omitempty"`
	Previousblockhash string    `json:"previousblockhash"`
	Nextblockhash     string    `json:"nextblockhash"`
}

type Block struct {
	RawBlock

	TxCount     uint   `json:"tx_count"`
	Stake       uint64 `json:"stake"`
	StakedBy    string `json:"stakedBy"`
	Spend       uint64 `json:"spend"`
	Fees        uint64 `json:"fees"`
	CFundPayout uint64 `json:"cfundPayout"`

	// Transient
	Best bool `json:"best,omitempty"`
}

func (b *Block) Slug() string {
	return slug.Make(fmt.Sprintf("block-%s", b.Hash))
}

type BlockCycle struct {
	Size   int
	Cycle  int
	Index  int
	Quorum int
}

func (b *Block) BlockCycle(size int) *BlockCycle {
	cycle := GetCycleForHeight(b.Height, size)

	return &BlockCycle{
		Size:  size,
		Cycle: cycle,
		Index: GetCycleIndex(b.Height, cycle, size),
	}
}

func (b *BlockCycle) IsEnd() bool {
	return b.Index == b.Size-1
}

func GetCycleForHeight(height uint64, size int) int {
	return (int(height) / size) + 1
}

func GetCycleIndex(height uint64, cycle int, size int) int {
	base := (cycle * size) - size
	return int(height) - base
}

func GetQuorum(size int, quorum int) int {
	return int((float64(quorum) / 100) * float64(size))
}
