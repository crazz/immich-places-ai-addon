package capabilities

import (
	"errors"
	"sync"
)

var (
	ErrBusy        = errors.New("capability test is busy")
	ErrDisabled    = errors.New("provider profile is disabled")
	ErrUnavailable = errors.New("provider profile is unavailable")
)

type Limiter struct {
	mu        sync.Mutex
	global    int
	perUser   map[string]int
	maxGlobal int
	maxUser   int
}

func NewLimiter(maxGlobal, maxUser int) *Limiter {
	return &Limiter{perUser: make(map[string]int), maxGlobal: maxGlobal, maxUser: maxUser}
}

func (l *Limiter) Acquire(userID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.global >= l.maxGlobal || l.perUser[userID] >= l.maxUser {
		return ErrBusy
	}
	l.global++
	l.perUser[userID]++
	return nil
}

func (l *Limiter) Release(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.perUser[userID] > 0 {
		l.perUser[userID]--
		if l.perUser[userID] == 0 {
			delete(l.perUser, userID)
		}
	}
	if l.global > 0 {
		l.global--
	}
}
