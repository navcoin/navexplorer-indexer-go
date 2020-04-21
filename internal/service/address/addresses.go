package address

import "github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"

var Addresses = make(addresses)

type addresses map[string]*explorer.Address

func (a *addresses) GetByHash(hash string) *explorer.Address {
	for _, a := range Addresses {
		if a.Hash == hash {
			return a
		}
	}

	return nil
}

func (a *addresses) Delete(hash string) {
	delete(Addresses, hash)
}
