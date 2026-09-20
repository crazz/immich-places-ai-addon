package jobs

import "time"

type Transition struct {
	State, Failure string
	Delay          time.Duration
}

func Interrupted(attempts, calls, jobCalls, maxCalls int, canceled bool) Transition {
	if canceled {
		return Transition{State: "canceled", Failure: "canceled"}
	}
	if attempts >= 3 || calls >= 3 || jobCalls >= maxCalls {
		return Transition{State: "failed", Failure: "budget"}
	}
	return Transition{State: "retry_wait", Failure: "interrupted", Delay: time.Second}
}
