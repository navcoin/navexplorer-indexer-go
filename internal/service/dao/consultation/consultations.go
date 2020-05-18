package consultation

import (
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
)

var Consultations consultations

type consultations []*explorer.Consultation

func (p *consultations) Add(c *explorer.Consultation) {
	Consultations = append(Consultations, c)
}

func (p *consultations) Delete(hash string) {
	for i := range Consultations {
		if Consultations[i].Hash == hash {
			Consultations[i] = Consultations[len(Consultations)-1]
			Consultations = Consultations[:len(Consultations)-1]
			return
		}
	}
}
