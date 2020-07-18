package explorer

type RawVouts []RawVout
type Vouts []Vout

func (vouts *Vouts) WithAddress(hash string) []Vout {
	filtered := make([]Vout, 0)

	for _, v := range *vouts {
		if v.HasAddress(hash) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func (vouts *Vouts) HasOutputOfType(txType VoutType) bool {
	for _, vout := range *vouts {
		if vout.ScriptPubKey.Type == txType {
			return true
		}
	}

	return false
}

func (vouts *Vouts) HasAddress(address string) bool {
	for _, vout := range *vouts {
		for _, a := range vout.ScriptPubKey.Addresses {
			if a == address {
				return true
			}
		}
	}

	return false
}

func (vouts *Vouts) GetSpendableAmount() uint64 {
	var amount uint64 = 0
	for _, o := range *vouts {
		if o.ScriptPubKey.Type == VoutCfundContribution {
			continue
		}
		amount += o.ValueSat
	}

	return amount
}

func (vouts *Vouts) GetAmount() uint64 {
	var amount uint64 = 0
	for _, o := range *vouts {
		amount += o.ValueSat
	}

	return amount
}

func (vouts *Vouts) GetAmountByAddress(address string, cold bool) (value float64, valuesat uint64) {
	for _, o := range *vouts {
		if cold {
			if len(o.ScriptPubKey.Addresses) > 1 && o.ScriptPubKey.Addresses[0] == address {
				value += o.Value
				valuesat += o.ValueSat
			}
		} else {
			if (len(o.ScriptPubKey.Addresses) == 1 && o.ScriptPubKey.Addresses[0] == address) ||
				(len(o.ScriptPubKey.Addresses) > 1 && o.ScriptPubKey.Addresses[1] == address) {
				value += o.Value
				valuesat += o.ValueSat
			}
		}
	}

	return
}

func (vouts *Vouts) FilterWithAddresses() Vouts {
	var filtered = make(Vouts, 0)
	for _, o := range *vouts {
		if len(o.ScriptPubKey.Addresses) != 0 {
			filtered = append(filtered, o)
		}
	}
	return filtered
}
