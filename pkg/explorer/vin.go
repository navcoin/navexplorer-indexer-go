package explorer

type RawVin struct {
	Coinbase  string     `json:"coinbase,omitempty"`
	Txid      *string    `json:"txid,omitempty"`
	Vout      *int       `json:"vout,omitempty"`
	ScriptSig *ScriptSig `json:"scriptSig,omitempty"`
	Sequence  uint32     `json:"sequence"`
}

type Vin struct {
	RawVin
	Value            float64        `json:"value,omitempty"`
	ValueSat         uint64         `json:"valuesat,omitempty"`
	Addresses        []string       `json:"addresses,omitempty"`
	PreviousOutput   PreviousOutput `json:"previousOutput,omitempty"`
	Private          bool           `json:"private"`
	Wrapped          bool           `json:"wrapped"`
	WrappedAddresses []string       `json:"wrappedAddresses,omitempty"`
}

type PreviousOutput struct {
	Height uint64   `json:"height"`
	Type   VoutType `json:"type"`
}

func (i *Vin) HasAddress(address string) bool {
	for idx := range i.Addresses {
		if i.Addresses[idx] == address {
			return true
		}
	}

	return false
}

func (i *Vin) IsCoinbase() bool {
	return i.Coinbase != ""
}

func (i *Vin) IsColdStakingAddress(address string) bool {
	return len(i.Addresses) == 2 && i.Addresses[0] == address
}

func (i *Vin) IsColdSpendingAddress(address string) bool {
	return len(i.Addresses) == 2 && i.Addresses[0] == address
}

func (i *Vin) IsPrivate() bool {
	return i.PreviousOutput.Type == VoutNonstandard && len(i.Addresses) == 0
}
