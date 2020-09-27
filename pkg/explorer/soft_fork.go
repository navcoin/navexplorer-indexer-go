package explorer

import (
	"fmt"
	"github.com/gosimple/slug"
	"log"
	"time"
)

type SoftForks []*SoftFork

type SoftFork struct {
	Name             string         `json:"name"`
	SignalBit        uint           `json:"signalBit"`
	StartTime        time.Time      `json:"startTime"`
	Timeout          time.Time      `json:"timeout"`
	State            SoftForkState  `json:"state"`
	LockedInHeight   uint64         `json:"lockedinheight,omitempty"`
	ActivationHeight uint64         `json:"activationheight,omitempty"`
	SignalHeight     uint64         `json:"signalheight,omitempty"`
	Cycles           SoftForkCycles `json:"cycles,omitempty"`
}

func (s *SoftFork) Slug() string {
	return slug.Make(fmt.Sprintf("softfork-%s", s.Name))
}

type SoftForkCycles []SoftForkCycle

type SoftForkCycle struct {
	Cycle            uint `json:"cycle"`
	BlocksSignalling int  `json:"blocks"`
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

func (s *SoftFork) IsActive() bool {
	if s.State == "" {
		log.Fatal("State cannot be null")
	}
	return s.State == SoftForkActive
}

func (s *SoftFork) GetCycle(cycle uint) *SoftForkCycle {
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

func (s SoftForks) HasSoftFork(name string) bool {
	for i, _ := range s {
		if s[i].Name == name {
			return true
		}
	}

	return false
}
