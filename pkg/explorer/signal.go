package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
)

type Signal struct {
	Address   string   `json:"address"`
	Height    uint64   `json:"height"`
	SoftForks []string `json:"softforks"`
}

func (s *Signal) Slug() string {
	return slug.Make(fmt.Sprintf("signal-%s-%d", s.Address, s.Height))
}

func (s *Signal) IsSignalling() bool {
	return len(s.SoftForks) > 0
}

func (s *Signal) DeleteSoftFork(name string) {
	softForks := make([]string, 0)

	for i := range s.SoftForks {
		if s.SoftForks[i] != name {
			softForks = append(softForks, s.SoftForks[i])
		}
	}

	s.SoftForks = softForks
}
