package jobs

import "time"

type Policy struct {
	Global, PerOwner int
	LeaseDuration    time.Duration
	ResearchDuration time.Duration
}

func DefaultPolicy() Policy { return Policy{Global: 2, PerOwner: 1, LeaseDuration: 180 * time.Second} }

func (p Policy) Valid() bool {
	return p.Global >= 1 && p.Global <= 16 && p.PerOwner >= 1 && p.PerOwner <= p.Global && p.LeaseDuration >= time.Second && p.LeaseDuration <= 10*time.Minute && (p.ResearchDuration == 0 || p.ResearchDuration >= time.Second && p.ResearchDuration <= 10*time.Minute)
}

type Lease struct {
	Owner, Installation, JobID, ItemID, Token, Asset, Model string
	Input                                                   Submission
	Attempts, Calls                                         int
	ExpiresAt                                               time.Time
}
