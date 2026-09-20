package jobs

import "time"

type ConsumerSettings struct {
	Policy    Policy
	Heartbeat time.Duration
	Idle      time.Duration
}

func (s ConsumerSettings) Valid() bool {
	return s.Policy.Valid() && s.Heartbeat >= time.Second && s.Heartbeat <= time.Minute && s.Policy.LeaseDuration > 120*time.Second+s.Heartbeat && s.Idle >= 100*time.Millisecond && s.Idle <= 30*time.Second
}
