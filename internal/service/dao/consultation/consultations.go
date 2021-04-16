package consultation

import (
	"github.com/NavExplorer/navexplorer-indexer-go/v2/pkg/explorer"
)

var Consultations consultations

type consultations map[string]explorer.Consultation

func (p *consultations) Add(c explorer.Consultation) {
	Consultations[c.Hash] = c
}

func (p *consultations) Delete(hash string) {
	for i := range Consultations {
		if Consultations[i].Hash == hash {
			delete(Consultations, hash)
			return
		}
	}
}
