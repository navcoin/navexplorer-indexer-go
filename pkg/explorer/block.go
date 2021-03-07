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
	id string

	RawBlock

	TxCount     uint   `json:"tx_count"`
	Stake       uint64 `json:"stake"`
	StakedBy    string `json:"stakedBy"`
	Spend       uint64 `json:"spend"`
	Fees        uint64 `json:"fees"`
	CFundPayout uint64 `json:"cfundPayout"`

	BlockCycle *BlockCycle `json:"block_cycle"`
	Cfund      *Cfund      `json:"cfund"`

	Supply Supply `json:"supply"`

	// Transient
	Best bool `json:"best,omitempty"`
}

type Supply struct {
	Public     float64 `json:"public"`
	PublicSat  uint64  `json:"publicsat"`
	Private    float64 `json:"private"`
	PrivateSat uint64  `json:"privatesat"`
	Wrapped    float64 `json:"wrapped"`
	WrappedSat uint64  `json:"wrappedsat"`
}

func (b *Block) Id() string {
	return b.id
}

func (b *Block) SetId(id string) {
	b.id = id
}

func (b *Block) Slug() string {
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
