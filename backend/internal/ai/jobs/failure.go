package jobs

import (
	"context"
	"errors"
	"time"
)

type FailureCode string

const (
	Budget    FailureCode = "budget"
	Blocked   FailureCode = "blocked"
	Transient FailureCode = "transient"
	Permanent FailureCode = "permanent"
)

type Failure struct {
	Code       FailureCode
	RetryAfter time.Duration
}

func (Failure) Error() string { return "AI job execution failed" }

func AfterFailure(attempts, calls, jobCalls, maxCalls int, failure Failure, jitter time.Duration) Transition {
	if failure.Code == Budget {
		return Transition{State: "failed", Failure: "budget"}
	}
	if failure.Code == Blocked {
		return Transition{State: "blocked", Failure: "blocked"}
	}
	if failure.Code != Transient {
		return Transition{State: "failed", Failure: "execution"}
	}
	if attempts >= 3 || calls >= 3 || jobCalls >= maxCalls {
		return Transition{State: "failed", Failure: "budget"}
	}
	jitter = min(max(jitter, 0), 5*time.Second)
	delay := 5*time.Second*time.Duration(1<<min(max(attempts-1, 0), 2)) + jitter
	delay = max(delay, min(max(failure.RetryAfter, 0), 5*time.Minute))
	return Transition{State: "retry_wait", Failure: "transient", Delay: delay}
}

func ExecutionFailure(err error) Failure {
	var value Failure
	if errors.As(err, &value) {
		return value
	}
	var pointer *Failure
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer
	}
	if errors.Is(err, ErrDenied) {
		return Failure{Code: Blocked}
	}
	if errors.Is(err, ErrBudget) {
		return Failure{Code: Budget}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return Failure{Code: Transient}
	}
	return Failure{Code: Permanent}
}
