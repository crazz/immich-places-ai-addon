package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJobPolicyBoundsAndSafeFailureClassification(t *testing.T) {
	if !DefaultPolicy().Valid() {
		t.Fatal("default policy invalid")
	}
	for _, policy := range []Policy{{}, {Global: 17, PerOwner: 1, LeaseDuration: time.Minute}, {Global: 1, PerOwner: 2, LeaseDuration: time.Minute}, {Global: 1, PerOwner: 1, LeaseDuration: time.Millisecond}, {Global: 1, PerOwner: 1, LeaseDuration: time.Hour}} {
		if policy.Valid() {
			t.Fatal("unbounded policy accepted", policy)
		}
	}
	cases := []struct {
		failure     Failure
		attempts    int
		jitter      time.Duration
		state, code string
		delay       time.Duration
	}{
		{Failure{Code: Transient, RetryAfter: -time.Hour}, 1, -time.Hour, "retry_wait", "transient", 5 * time.Second},
		{Failure{Code: Transient}, 2, time.Hour, "retry_wait", "transient", 15 * time.Second},
		{Failure{Code: Transient, RetryAfter: time.Hour}, 1, time.Hour, "retry_wait", "transient", 5 * time.Minute},
		{Failure{Code: Transient}, 3, 0, "failed", "budget", 0},
		{Failure{Code: Blocked}, 1, 0, "blocked", "blocked", 0},
		{Failure{Code: FailureCode("private-input")}, 1, 0, "failed", "execution", 0},
	}
	for _, tc := range cases {
		got := AfterFailure(tc.attempts, 0, 0, 3, tc.failure, tc.jitter)
		if got.State != tc.state || got.Failure != tc.code || got.Delay != tc.delay {
			t.Fatal("wrong bounded transition", got)
		}
	}
	for _, tc := range []struct {
		err  error
		code FailureCode
	}{
		{context.DeadlineExceeded, Transient}, {ErrBudget, Budget}, {errors.New("private-input"), Permanent}, {Failure{Code: Transient}, Transient}, {&Failure{Code: Blocked}, Blocked},
	} {
		if got := ExecutionFailure(tc.err); got.Code != tc.code || strings.Contains(got.Error(), "private") {
			t.Fatal("unsafe classification", got.Code)
		}
	}
}
