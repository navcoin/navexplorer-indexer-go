package explorer

import (
	"log"
)

type SoftForks []*SoftFork

type SoftFork struct {
	MetaData MetaData `json:"-"`

	Name             string         `json:"name"`
	SignalBit        uint           `json:"signalBit"`
	State            SoftForkState  `json:"state"`
	LockedInHeight   uint64         `json:"lockedinheight,omitempty"`
	ActivationHeight uint64         `json:"activationheight,omitempty"`
	SignalHeight     uint64         `json:"signalheight,omitempty"`
	Cycles           SoftForkCycles `json:"cycles,omitempty"`
}

type SoftForkCycles []SoftForkCycle

type SoftForkCycle struct {
	Cycle            int `json:"cycle"`
	BlocksSignalling int `json:"blocks"`
}

func (s *SoftFork) LatestCycle() *SoftForkCycle {
	if len(s.Cycles) == 0 {
		return nil
	}

	return &(s.Cycles)[len(s.Cycles)-1]
}

func (s *SoftFork) IsOpen() bool {
	if s.State == "" {
		log.Fatal("State cannot be null")
	}
	return s.State == SoftForkDefined || s.State == SoftForkStarted || s.State == SoftForkFailed
}

func (s *SoftFork) GetCycle(cycle int) *SoftForkCycle {
	for i, c := range s.Cycles {
		if c.Cycle == cycle {
			return &s.Cycles[i]
		}
	}
	return nil
}

func (s SoftForks) GetSoftFork(name string) *SoftFork {
	for i, _ := range s {
		if s[i].Name == name {
			return s[i]
		}
	}

	return nil
}
