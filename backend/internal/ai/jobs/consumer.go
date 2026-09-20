package jobs

import (
	"context"
	"sync"
	"time"
)

type Consumer struct {
	Worker   Worker
	Settings ConsumerSettings
}

func (c Consumer) Run(ctx context.Context, onFailure func()) error {
	if !c.Settings.Valid() || c.Worker.Store == nil || c.Worker.Execute == nil {
		return ErrInvalid
	}
	c.Worker.Policy = c.Settings.Policy
	var workers sync.WaitGroup
	for i := 0; i < c.Settings.Policy.Global; i++ {
		workers.Add(1)
		go func() { defer workers.Done(); c.loop(ctx, onFailure) }()
	}
	workers.Wait()
	return nil
}
func (c Consumer) loop(ctx context.Context, onFailure func()) {
	ticker := time.NewTicker(c.Settings.Heartbeat)
	defer ticker.Stop()
	for ctx.Err() == nil {
		worked, err := c.Worker.RunOne(ctx, ticker.C)
		if ctx.Err() != nil {
			return
		}
		if err != nil && onFailure != nil {
			onFailure()
		}
		if worked && err == nil {
			continue
		}
		timer := time.NewTimer(c.Settings.Idle)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
