package main

import (
	"context"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestAIStackDisableAndShutdownPreserveSuccessAndUnknownSender(t *testing.T) {
	for _, mode := range []string{"disable", "shutdown"} {
		t.Run(mode, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			m := stackMutationServer(t, f)
			if err := f.writer.runOne(context.Background(), testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			var disabled atomic.Bool
			f.writer.enabled = func() bool { return !disabled.Load() }
			entered, release, returned := make(chan struct{}), make(chan struct{}), make(chan struct{})
			m.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PATCH" {
					close(entered)
					<-release
					defer close(returned)
					if mode == "shutdown" {
						return true
					}
				}
				return false
			}
			lock, err := acquireAIWriteLock(filepath.Join(t.TempDir(), "stack-shutdown.lock"))
			if err != nil {
				t.Fatal(err)
			}
			runtime := &aiWriteRuntime{store: f.writer, lock: lock}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan struct{})
			go func() {
				if mode == "shutdown" {
					runtime.run(ctx)
				} else {
					if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
						t.Error(err)
					}
					_ = lock.Close()
				}
				close(done)
			}()
			<-entered
			disabled.Store(true)
			if mode == "shutdown" {
				cancel()
				select {
				case <-done:
				case <-time.After(3 * time.Second):
					close(release)
					t.Fatal("shutdown failed to cancel active sender")
				}
			}
			close(release)
			<-returned
			if mode == "disable" {
				<-done
				cancel()
			}
			for range 3 {
				_ = f.writer.runOne(context.Background(), testUserID, op.ID)
			}
			saved, err := f.writer.get(context.Background(), testUserID, op.ID, false)
			if err != nil || saved.Targets[0].Status != "succeeded" || !saved.Targets[0].Verified || m.sent(saved.Targets[0].AssetID) != 1 {
				t.Fatal("lifecycle stop rewrote prior success", err)
			}
			if saved.Targets[1].Attempts != 1 || m.sent(saved.Targets[1].AssetID) != 1 || saved.Targets[1].Status == "canceled" {
				t.Fatal("possibly acting member lost durable evidence")
			}
			if saved.Targets[2].Status != "canceled" || saved.Targets[2].Attempts != 0 || m.sent(saved.Targets[2].AssetID) != 0 {
				t.Fatal("pending member sent after disablement")
			}
			if mode == "shutdown" && (saved.Targets[1].Settled || saved.Targets[1].Verified) {
				t.Fatal("shutdown invented sender completion or success")
			}
			var held, wantGuard int
			if mode == "shutdown" {
				wantGuard = 1
			} else if saved.Targets[1].Status != "succeeded" || !saved.Targets[1].Settled {
				t.Fatal("disablement prevented safe readback of already sent member")
			}
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE assetID=?`, saved.Targets[1].AssetID).Scan(&held); err != nil || held != wantGuard {
				t.Fatal("target exclusion does not reflect sender evidence", err)
			}
		})
	}
}
