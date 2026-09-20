package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
)

func TestAIImageConcurrentPreparationNeverWritesPersistenceOrUpstream(t *testing.T) {
	for _, mode := range []string{"success", "failure", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			f := newAIImageFixture(t)
			f.db.db.SetMaxOpenConns(1)
			var before, after int64
			if err := f.db.db.QueryRow("SELECT total_changes()").Scan(&before); err != nil {
				t.Fatal(err)
			}
			if _, err := f.db.db.Exec("PRAGMA query_only=ON"); err != nil {
				t.Fatal(err)
			}
			if mode == "failure" {
				f.handle = func(w http.ResponseWriter, r *http.Request) bool { w.WriteHeader(503); return true }
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "canceled" {
				cancel()
			}
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					p, err := f.service.prepare(ctx, testUserID, f.store.binding, selectionA)
					if mode == "success" {
						if err != nil || p == nil {
							t.Errorf("read-only preparation failed: %v", err)
						}
					} else if err == nil || p != nil {
						t.Error("failed preparation published a copy")
					}
					p.Release()
				}()
			}
			wg.Wait()
			if err := f.db.db.QueryRow("SELECT total_changes()").Scan(&after); err != nil {
				t.Fatal(err)
			}
			calls, bad := f.counts()
			want := 24
			if mode == "failure" {
				want = 8
			}
			if mode == "canceled" {
				want = 0
			}
			if before != after || bad != 0 || calls != want {
				t.Fatalf("unexpected effects: changes=%d->%d calls=%d bad=%d", before, after, calls, bad)
			}
		})
	}
}
