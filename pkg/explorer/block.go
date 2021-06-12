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

	BlockCycle BlockCycle `json:"block_cycle"`
	Cfund      Cfund      `json:"cfund"`

	SupplyBalance SupplyBalance `json:"supply_balance"`
	SupplyChange  SupplyChange  `json:"supply_change"`

	// Transient
	Best bool `json:"best,omitempty"`
}

type SupplyBalance struct {
	Public  uint64 `json:"public"`
	Private uint64 `json:"private"`
	Wrapped uint64 `json:"wrapped"`
}

func (sb SupplyBalance) Total() uint64 {
	return sb.Public + sb.Private + sb.Wrapped
}

type SupplyChange struct {
	Public  int64 `json:"public"`
	Private int64 `json:"private"`
	Wrapped int64 `json:"wrapped"`
}

func (b Block) Slug() string {
	return slug.Make(fmt.Sprintf("block-%s", b.Hash))
}

type BlockCycle struct {
	Size           uint `json:"size"`
	Cycle          uint `json:"cycle"`
	Index          uint `json:"index"`
	Transitory     bool `json:"transitory"`
	TransitorySize uint `json:"transitorySize"`
}

func (b *BlockCycle) IsEnd() bool {
	return b.Index == b.Size-1
}

func GetQuorum(size uint, quorum int) int {
	return int((float64(quorum) / 100.0) * float64(size))
}

type Cfund struct {
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
}
