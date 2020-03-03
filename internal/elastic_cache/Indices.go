package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
)

type Indices string

var (
	AddressIndex            Indices = "address"
	AddressTransactionIndex Indices = "addresstransaction"
	BlockIndex              Indices = "block"
	BlockTransactionIndex   Indices = "blocktransaction"
	ConsensusIndex          Indices = "consensus"
	ProposalIndex           Indices = "proposal"
	DaoVoteIndex            Indices = "daovote"
	DaoConsultationIndex    Indices = "consultation"
	PaymentRequestIndex     Indices = "paymentrequest"
	SignalIndex             Indices = "signal"
	SoftForkIndex           Indices = "softfork"
)

// Sets the network and returns the full string
func (i *Indices) Get() string {
	return fmt.Sprintf("%s.%s", config.Get().Network, string(*i))
}

func All() []Indices {
	return []Indices{
		AddressIndex,
		AddressTransactionIndex,
		BlockIndex,
		BlockTransactionIndex,
		ConsensusIndex,
		ProposalIndex,
		PaymentRequestIndex,
		DaoVoteIndex,
		SignalIndex,
		SoftForkIndex,
	}
}
