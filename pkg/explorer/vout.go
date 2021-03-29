package explorer

type RawVout struct {
	Value        float64      `json:"value"`
	ValueSat     uint64       `json:"valuesat"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
	SpendingKey  string       `json:"spendingKey,omitempty"`
	OutputKey    string       `json:"outputKey,omitempty"`
	EphemeralKey string       `json:"ephemeralKey,omitempty"`
	RangeProof   bool         `json:"rangeProof,omitempty"`
	SpentTxId    string       `json:"spentTxId,omitempty"`
	SpentIndex   int          `json:"spentIndex,omitempty"`
	SpentHeight  uint64       `json:"spentHeight,omitempty"`
}

type Vout struct {
	RawVout
	Redeemed   bool        `json:"redeemed"`
	RedeemedIn *RedeemedIn `json:"redeemedIn,omitempty"`
	MultiSig   *MultiSig   `json:"multisig,omitempty"`
	Private    bool        `json:"private"`
	Wrapped    bool        `json:"wrapped"`
}

type RedeemedIn struct {
	Hash   string `json:"hash,omitempty"`
	Index  int    `json:"index,omitempty"`
	Height uint64 `json:"height,omitempty"`
}

type MultiSig struct {
	Hash       string   `json:"hash,omitempty"`
	Signatures []string `json:"signatures"`
	Required   int      `json:"required"`
	Total      int      `json:"total"`
}

func (o *Vout) HasAddress(hash string) bool {
	for _, a := range o.ScriptPubKey.Addresses {
		if a == hash {
			return true
		}
	}

	return false
}

func (o *Vout) IsMultiSig() bool {
	return o.MultiSig != nil
}

func (o *Vout) IsPrivateFee() bool {
	return o.ScriptPubKey.Type == VoutNulldata || o.ScriptPubKey.Asm == "OP_RETURN"
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
