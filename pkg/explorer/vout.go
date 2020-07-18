package explorer

type RawVout struct {
	Value        float64      `json:"value"`
	ValueSat     uint64       `json:"valuesat"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}
type Vout struct {
	RawVout
	RedeemedIn *RedeemedIn `json:"redeemedIn,omitempty"`
}

type RedeemedIn struct {
	Hash   string `json:"hash,omitempty"`
	Height uint64 `json:"height,omitempty"`
}

func (o *Vout) HasAddress(hash string) bool {
	for _, a := range o.ScriptPubKey.Addresses {
		if a == hash {
			return true
		}
	}

	return false
}

func (o *Vout) IsColdStaking() bool {
	return o.ScriptPubKey.Type == VoutColdStaking || o.ScriptPubKey.Type == VoutColdStakingV2
}

func (o *Vout) IsProposalVote() bool {
	return o.ScriptPubKey.Type == VoutProposalYesVote || o.ScriptPubKey.Type == VoutProposalNoVote
}

func (o *Vout) IsPaymentRequestVote() bool {
	return o.ScriptPubKey.Type == VoutPaymentRequestYesVote || o.ScriptPubKey.Type == VoutPaymentRequestNoVote
}

func (o *Vout) IsColdStakingAddress(address string) bool {
	return len(o.ScriptPubKey.Addresses) == 2 && o.ScriptPubKey.Addresses[0] == address
}

func (o *Vout) IsColdSpendingAddress(address string) bool {
	return len(o.ScriptPubKey.Addresses) == 2 && o.ScriptPubKey.Addresses[1] == address
}

func (o *Vout) IsColdVotingAddress(address string) bool {
	return len(o.ScriptPubKey.Addresses) == 3 && o.ScriptPubKey.Addresses[2] == address
}
