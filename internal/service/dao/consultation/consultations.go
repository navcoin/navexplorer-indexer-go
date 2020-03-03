package consultation

import "github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"

var Consultations consultations

type consultations []*explorer.Consultation

func (p *consultations) Delete(hash string) {
	for i := range Consultations {
		if Consultations[i].Hash == hash {
			Consultations[i] = Consultations[len(Consultations)-1]
			Consultations[len(Consultations)-1] = nil
			Consultations = append([]*explorer.Consultation(nil), Consultations[:len(Consultations)-1]...)
			break
		}
	}
}
