package elastic_cache

import (
	"fmt"
	"github.com/NavExplorer/navexplorer-indexer-go/internal/config"
)

type Indices string

var (
	AddressIndex          Indices = "address"
	AddressHistoryIndex   Indices = "addresshistory"
	BlockIndex            Indices = "block"
	BlockTransactionIndex Indices = "blocktransaction"
	ConsensusIndex        Indices = "consensus"
	ProposalIndex         Indices = "proposal"
	DaoVoteIndex          Indices = "daovote"
	DaoConsultationIndex  Indices = "consultation"
	PaymentRequestIndex   Indices = "paymentrequest"
	SignalIndex           Indices = "signal"
	SoftForkIndex         Indices = "softfork"
)

// Sets the network and returns the full string
func (i *Indices) Get() string {
	return fmt.Sprintf("%s.%s.%s", config.Get().Network, config.Get().Index, string(*i))
}

func All() []Indices {
	return []Indices{
		AddressIndex,
		AddressHistoryIndex,
		BlockIndex,
		BlockTransactionIndex,
		ConsensusIndex,
		ProposalIndex,
		DaoVoteIndex,
		DaoConsultationIndex,
		PaymentRequestIndex,
		SignalIndex,
		SoftForkIndex,
	}
}
