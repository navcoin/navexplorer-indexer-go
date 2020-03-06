package explorer

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

type ConsensusParameter struct {
	MetaData MetaData `json:"-"`

	Id          int                    `json:"id"`
	Description string                 `json:"desc"`
	Type        ConsensusParameterType `json:"type"`
	Value       int                    `json:"value"`
}
