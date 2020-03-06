package explorer

type ConsensusParameterType int

var (
	NumberType  ConsensusParameterType = 0
	PercentType ConsensusParameterType = 1
	NavType     ConsensusParameterType = 2
	BoolType    ConsensusParameterType = 3
)

type ConsensusParameters []*ConsensusParameter

type ConsensusParameter struct {
	MetaData MetaData `json:"-"`

	Id          int                    `json:"id"`
	Description string                 `json:"desc"`
	Type        ConsensusParameterType `json:"type"`
	Value       int                    `json:"value"`
}
