package providerhttp

import (
	"context"
	"time"
)

func responseContext(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc, time.Duration) {
	ctx, cancel := context.WithTimeout(parent, duration)
	deadline, _ := ctx.Deadline()
	wait := time.Until(deadline)
	if duration <= defaultTimeout && wait > time.Minute {
		wait = time.Minute
	}
	return ctx, cancel, wait
}
