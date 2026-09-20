package main

import (
	"context"
	"fmt"
	"testing"
)

func TestAIJobRecoveryIsBoundedAndKeepsSystemicBlocks(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		f := newAIJobFixture(t)
		ctx := context.Background()
		f.input.AssetIDs = nil
		f.input.MaxCalls = 303
		for i := 0; i < 101; i++ {
			f.input.AssetIDs = append(f.input.AssetIDs, fmt.Sprintf("00000000-0000-4000-8000-%012d", i))
		}
		job, err := f.store.Submit(ctx, f.input)
		if err != nil {
			t.Fatal(err)
		}
		selectionSQL(t, f.db, "UPDATE ai_job_items SET state='running',attempts=1,leaseToken=id,leaseExpiresAt=? WHERE jobID=?", f.now.UnixNano(), job.ID)
		if blocked {
			selectionSQL(t, f.db, "UPDATE ai_jobs SET blocked=1 WHERE id=?", job.ID)
		}
		if n, err := f.store.Recover(ctx); err != nil || n != 100 {
			t.Fatal("unbounded recovery", n, err)
		}
		got, err := f.store.Get(ctx, testUserID, job.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := "retry_wait"
		if blocked {
			want = "blocked"
		}
		running, recovered := 0, 0
		for _, item := range got.Items {
			if item.State == "running" {
				running++
			}
			if item.State == want {
				recovered++
			}
			if item.Calls != 0 {
				t.Fatal("invented reservation")
			}
		}
		if running != 1 || recovered != 100 {
			t.Fatal("wrong recovery batch", running, recovered)
		}
		if n, err := f.store.Recover(ctx); err != nil || n != 1 {
			t.Fatal("remaining recovery", n, err)
		}
	}
}
