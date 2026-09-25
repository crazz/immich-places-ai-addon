package main

import (
	"context"
	"testing"
	"time"
)

func TestAIWriteExpiredReservationCannotStartTransport(t *testing.T) {
	w := newAIWriteFixture(t)
	ctx := context.Background()
	a := &aiWriteAttempt{store: w.writer}
	if _, err := a.Read(ctx, w.op); err != nil {
		t.Fatal(err)
	}
	if err := a.Reserve(ctx, w.op, false); err != nil {
		t.Fatal(err)
	}
	w.f.now = w.f.now.Add(45 * time.Second)
	outcome := a.Send(ctx, w.op)
	if !outcome.Known || outcome.Code != "not_sent" {
		t.Fatal("expired sender dispatched", outcome)
	}
	if sends, _ := w.counts(); sends != 0 {
		t.Fatal(sends)
	}
}
