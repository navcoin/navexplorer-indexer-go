package consensus

import (
	"github.com/NavExplorer/navcoind-go"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

func CreateConsensusParameter(consensusParameter navcoind.ConsensusParameter) *explorer.ConsensusParameter {
	return &explorer.ConsensusParameter{
		Id:          consensusParameter.Id,
		Description: consensusParameter.Description,
		Type:        explorer.ConsensusParameterType(consensusParameter.Type),
		Value:       consensusParameter.Value,
	}
}

func UpdateConsensus(consensusParameter navcoind.ConsensusParameter, c *explorer.ConsensusParameter) {
	c.Id = consensusParameter.Id
	c.Description = consensusParameter.Description
	c.Type = explorer.ConsensusParameterType(consensusParameter.Type)
	c.Value = consensusParameter.Value
}
