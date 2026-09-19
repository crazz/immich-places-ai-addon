package providers

import (
	"context"
	"testing"
	"time"
)

type deadlineTransport struct {
	t          *testing.T
	wantLatest time.Time
	deadline   time.Time
}

func (d *deadlineTransport) Pin(ctx context.Context, _ EgressRule) (PinnedDestination, error) {
	deadline, ok := ctx.Deadline()
	if !ok || deadline.After(d.wantLatest) {
		d.t.Fatal("DNS must share a bounded total dispatch deadline")
	}
	d.deadline = deadline
	return PinnedDestination{}, nil
}

func (d *deadlineTransport) Transmit(ctx context.Context, _ PinnedDestination, _ []byte, _ string) (DispatchResult, error) {
	deadline, ok := ctx.Deadline()
	if !ok || !deadline.Equal(d.deadline) {
		d.t.Fatal("transmission must retain the DNS deadline")
	}
	return DispatchResult{}, nil
}

func TestDispatchBoundsDNSAndRetainsEarlierCallerDeadline(t *testing.T) {
	for _, callerLimit := range []time.Duration{0, time.Second} {
		ctx := context.Background()
		latest := time.Now().Add(120*time.Second + time.Second)
		if callerLimit != 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, callerLimit)
			defer cancel()
			latest, _ = ctx.Deadline()
		}
		transport := &deadlineTransport{t: t, wantLatest: latest}
		d := Dispatcher{Enabled: true, Store: &countingStore{}, Transport: transport,
			Policy: EgressPolicy{Rules: []EgressRule{{BaseURL: "http://127.0.0.1:9/v1"}}}}
		if _, err := d.Dispatch(ctx, DispatchRequest{OwnerID: "o", ProfileID: "p", Revision: 1}); err != nil {
			t.Fatal(err)
		}
	}
}
