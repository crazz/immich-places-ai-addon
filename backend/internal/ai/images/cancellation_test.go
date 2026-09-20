package images

import (
	"context"
	"errors"
	"testing"
)

type cancelAfterCheck struct {
	context.Context
	checks int
}

func (c *cancelAfterCheck) Err() error {
	c.checks++
	if c.checks > 1 {
		return context.Canceled
	}
	return nil
}

func TestPrepareRasterNeverPublishesAfterCancellation(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, ctx := range []context.Context{canceled, &cancelAfterCheck{Context: context.Background()}} {
		p, err := PrepareRaster(ctx, jpegFixture(t, 8, 4), "image/jpeg", testBinding(), DefaultLimits())
		if !errors.Is(err, context.Canceled) || p != nil {
			t.Fatalf("canceled preparation published: %v", err)
		}
	}
}
