package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type ConsensusParameterType int

var (
	NumberType  ConsensusParameterType = 0
	PercentType ConsensusParameterType = 1
	NavType     ConsensusParameterType = 2
	BoolType    ConsensusParameterType = 3
)

type ConsensusParameters struct {
	parameters []*ConsensusParameter
}

func (p *ConsensusParameters) Add(c *ConsensusParameter) {
	p.parameters = append(p.parameters, c)
}

func (p *ConsensusParameters) Get(id int) *ConsensusParameter {
	for _, p := range p.parameters {
		if p.Id == id {
			return p
		}
	}

	return nil
}

func (p *ConsensusParameters) All() []*ConsensusParameter {
	return p.parameters
}

type ConsensusParameter struct {
	Id             int                    `json:"id"`
	Description    string                 `json:"desc"`
	Type           ConsensusParameterType `json:"type"`
	Value          int                    `json:"value"`
	UpdatedOnBlock uint64                 `json:"updatedOnBlock"`
}

func (cp *ConsensusParameter) Slug() string {
	return slug.Make(fmt.Sprintf("consensus-%d", cp.Id))
}
