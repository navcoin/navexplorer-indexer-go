package explorer

type RawVins []RawVin

type Vins []Vin

func (vins *Vins) Empty() bool {
	return len(*vins) == 0
}

func (vins *Vins) First() Vin {
	return (*vins)[0]
}

func (vins *Vins) HasAddress(address string) bool {
	for _, vin := range *vins {
		for _, a := range vin.Addresses {
			if a == address {
				return true
			}
		}
	}

	return false
}

func (vins *Vins) GetAmount() uint64 {
	var amount uint64 = 0
	for _, i := range *vins {
		amount += i.ValueSat
	}
	return amount
}

func (vins *Vins) GetAmountByAddress(address string, cold bool) (value float64, valuesat uint64) {
	for _, i := range *vins {
		if cold {
			if len(i.Addresses) > 1 && i.Addresses[0] == address {
				value += i.Value
				valuesat += i.ValueSat
			}
		} else {
			if (len(i.Addresses) == 1 && i.Addresses[0] == address) ||
				(len(i.Addresses) > 1 && i.Addresses[1] == address) {
				value += i.Value
				valuesat += i.ValueSat
			}
		}
	}

	return
}

func (vins *Vins) FilterWithAddresses() Vins {
	var filtered = make(Vins, 0)
	for _, i := range *vins {
		if len(i.Addresses) != 0 {
			filtered = append(filtered, i)
		}
	}

	return filtered
}
