package jobs

import "time"

func CleanupAllowed(owner string, cutoff, now time.Time, limit int) bool {
	return boundedIdentity(owner) && !cutoff.IsZero() && !cutoff.After(now) && limit >= 1 && limit <= 100
}
