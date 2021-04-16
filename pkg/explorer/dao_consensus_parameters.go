package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type ConsensusParameterType int
type Parameter int

var (
	NumberType                              ConsensusParameterType = 0
	PercentType                             ConsensusParameterType = 1
	NavType                                 ConsensusParameterType = 2
	BoolType                                ConsensusParameterType = 3
	VOTING_CYCLE_LENGTH                     Parameter              = 0
	CONSULTATION_MIN_SUPPORT                Parameter              = 1
	CONSULTATION_ANSWER_MIN_SUPPORT         Parameter              = 2
	CONSULTATION_MIN_CYCLES                 Parameter              = 3
	CONSULTATION_MAX_VOTING_CYCLES          Parameter              = 4
	CONSULTATION_MAX_SUPPORT_CYCLES         Parameter              = 5
	CONSULTATION_REFLECTION_LENGTH          Parameter              = 6
	CONSULTATION_MIN_FEE                    Parameter              = 7
	CONSULTATION_ANSWER_MIN_FEE             Parameter              = 8
	PROPOSAL_MIN_QUORUM                     Parameter              = 9
	PROPOSAL_MIN_ACCEPT                     Parameter              = 10
	PROPOSAL_MIN_REJECT                     Parameter              = 11
	PROPOSAL_MIN_FEE                        Parameter              = 12
	PROPOSAL_MAX_VOTING_CYCLES              Parameter              = 13
	PAYMENT_REQUEST_MIN_QUORUM              Parameter              = 14
	PAYMENT_REQUEST_MIN_ACCEPT              Parameter              = 15
	PAYMENT_REQUEST_MIN_REJECT              Parameter              = 16
	PAYMENT_REQUEST_MIN_FEE                 Parameter              = 17
	PAYMENT_REQUEST_MAX_VOTING_CYCLES       Parameter              = 18
	FUND_SPREAD_ACCUMULATION                Parameter              = 19
	FUND_PERCENT_PER_BLOCK                  Parameter              = 20
	GENERATION_PER_BLOCK                    Parameter              = 21
	NAVNS_FEE                               Parameter              = 22
	CONSENSUS_PARAMS_DAO_VOTE_LIGHT_MIN_FEE Parameter              = 23
)

type ConsensusParameters []*ConsensusParameter

func (cps ConsensusParameters) GetConsensusParameter(parameter Parameter) *ConsensusParameter {
	for idx := range cps {
		if cps[idx].Id == int(parameter) {
			return cps[idx]
		}
	}

	return nil
}

func (cps ConsensusParameters) GetConsensusParameterById(parameter int) *ConsensusParameter {
	for idx := range cps {
		if cps[idx].Id == parameter {
			return cps[idx]
		}
	}

	return nil
}

type ConsensusParameter struct {
	Id             int                    `json:"id"`
	Description    string                 `json:"desc"`
	Type           ConsensusParameterType `json:"type"`
	Value          int                    `json:"value"`
	UpdatedOnBlock uint64                 `json:"updatedOnBlock"`
}

func (cp ConsensusParameter) Slug() string {
	return slug.Make(fmt.Sprintf("consensus-%d", cp.Id))
}
