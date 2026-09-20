package providers

import (
	"context"
	"time"
)

type dispatchTimeoutKey struct{}

func DispatchTimeout(ctx context.Context) time.Duration {
	if duration, ok := ctx.Value(dispatchTimeoutKey{}).(time.Duration); ok {
		return duration
	}
	return MaxDispatchDuration
}
