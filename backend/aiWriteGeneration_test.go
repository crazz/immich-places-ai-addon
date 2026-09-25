package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteReopenEveryReservedPhaseFencesLateWorkers(t *testing.T) {
	for _, phase := range []string{"reserved", "sent", "verifying"} {
		t.Run(phase, func(t *testing.T) {
			w := newAIWriteFixture(t)
			ctx := context.Background()
			old := &aiWriteAttempt{store: w.writer}
			if _, err := old.Read(ctx, w.op); err != nil {
				t.Fatal(err)
			}
			if err := old.Reserve(ctx, w.op, false); err != nil {
				t.Fatal(err)
			}
			outcome := writeback.Completion{}
			if phase != "reserved" {
				outcome = old.Send(ctx, w.op)
			}
			if phase == "verifying" {
				if err := old.Sent(ctx, w.op, outcome); err != nil {
					t.Fatal(err)
				}
			}
			w.f.reopen(t)
			w.writer.drafts.results.jobs = w.f.store
			w.writer.sync = newSyncService(w.f.db, nil, nil)
			w.f.now = w.f.now.Add(time.Minute)
			w.run(t)
			current := w.status(t)
			want := "succeeded"
			sends := 1
			if phase == "reserved" {
				want = "verifying"
				sends = 0
			}
			if current.Status != want || current.Attempts != 1 || current.Generation != 2 {
				t.Fatal(phase, current)
			}
			if err := old.Sent(ctx, w.op, writeback.Completion{Known: true, Code: "late"}); err == nil {
				t.Fatal("late completion published")
			}
			old.Send(ctx, w.op)
			if actual, _ := w.counts(); actual != sends {
				t.Fatal("recovery or old worker resent", actual, sends)
			}
			after := w.status(t)
			if after.Status != current.Status || after.Generation != current.Generation || after.Code != current.Code {
				t.Fatal("late worker changed state", after)
			}
		})
	}
}
