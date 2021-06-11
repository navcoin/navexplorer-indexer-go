package vote

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
	"go.uber.org/zap"
)

func CreateVotes(block *explorer.Block, tx explorer.BlockTransaction, header *navcoind.BlockHeader, votingAddress string) *explorer.DaoVotes {
	if !tx.IsCoinbase() {
		return nil
	}

	daoVote := &explorer.DaoVotes{Cycle: block.BlockCycle.Cycle, Height: tx.Height, Address: votingAddress}

	if block.Nonce&1 == 1 {
		vote := explorer.Vote{Type: explorer.ExcludeVote, Hash: block.StakedBy, Vote: 0}
		zap.S().Debugf("%d Excluding vote for %s - %d", block.Height, block.StakedBy, 0)
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	for _, v := range header.CfundVotes {
		vote := explorer.Vote{Type: explorer.ProposalVote, Hash: v.Hash, Vote: v.Vote}
		zap.S().Debugf("%d Adding cfund proposal vote for %s - %d", block.Height, v.Hash, v.Vote)
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	for _, v := range header.CfundRequestVotes {
		vote := explorer.Vote{Type: explorer.PaymentRequestVote, Hash: v.Hash, Vote: v.Vote}
		zap.S().Debugf("%d Adding cfund payment request vote for %s - %d", block.Height, v.Hash, v.Vote)
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	for _, v := range header.DaoSupport {
		vote := explorer.Vote{Type: explorer.DaoSupport, Hash: v}
		zap.S().Debugf("%d Adding dao support for %s", block.Height, v)
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	for _, v := range header.DaoVotes {
		vote := explorer.Vote{Type: explorer.DaoVote, Hash: v.Hash, Vote: v.Vote}
		zap.S().Debugf("%d Adding dao vote for %s - %d", block.Height, v.Hash, v.Vote)
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	if len(daoVote.Votes) != 0 {
		return daoVote
	}

	// Legacy support
	daoVote = &explorer.DaoVotes{Height: tx.Height, Address: block.StakedBy}
	for _, vout := range tx.Vout {
		if !vout.IsProposalVote() && !vout.IsPaymentRequestVote() {
			continue
		}
		vote := explorer.Vote{Hash: vout.ScriptPubKey.Hash, Vote: persuasion(vout.ScriptPubKey.Type)}
		if vout.IsProposalVote() {
			vote.Type = explorer.ProposalVote
		}
		if vout.IsPaymentRequestVote() {
			vote.Type = explorer.PaymentRequestVote
		}
		daoVote.Votes = append(daoVote.Votes, vote)
	}

	if len(daoVote.Votes) != 0 {
		return daoVote
	}

	return nil
}

func persuasion(voutType explorer.VoutType) int {
	switch voutType {

	case explorer.VoutProposalYesVote:
		return 1
	case explorer.VoutPaymentRequestYesVote:
		return 1
	case explorer.VoutProposalNoVote:
		return -1
	case explorer.VoutPaymentRequestNoVote:
		return -1
	}

	return 0
}
