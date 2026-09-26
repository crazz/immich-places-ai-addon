package main

import (
	"context"
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestAIStackRetryAffectsOneTargetAndReplaysLocallyAfterSuccess(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	second := op.Plan.Manifest.Targets[1].AssetID
	m := stackMutationServer(t, f)
	var requests atomic.Int32
	var unavailable atomic.Bool
	m.handle = func(w http.ResponseWriter, r *http.Request) bool {
		requests.Add(1)
		if unavailable.Load() {
			w.WriteHeader(503)
			return true
		}
		if r.Method == "PATCH" && r.URL.Path == "/api/assets/"+second && m.sent(second) == 1 {
			w.WriteHeader(400)
			return true
		}
		return false
	}
	ctx := context.Background()
	for range 3 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "partial" || saved.Targets[1].Status != "retryable" {
		t.Fatal("fixture did not reach partial retry", saved.Status, err)
	}
	generation := saved.Targets[1].Generation
	accepted, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, second, generation)
	if err != nil || accepted.Targets[1].Status != "queued" || accepted.Targets[1].Generation != generation+1 {
		t.Fatal("eligible exact retry rejected", err)
	}
	if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	success, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || success.Status != "succeeded" || success.Targets[1].Attempts != 2 {
		t.Fatal("exact retry did not finish", err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.f.now = f.f.now.Add(6 * time.Minute)
	f.writer.enabled = func() bool { return false }
	unavailable.Store(true)
	before := requests.Load()
	replayed, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, second, generation)
	if err != nil || !reflect.DeepEqual(replayed, success) || requests.Load() != before {
		t.Fatal("accepted retry replay consulted upstream or altered state", err)
	}
	if _, err := f.writer.retryStackTarget(ctx, "foreign", op.ID, second, generation); err == nil {
		t.Fatal("foreign owner replayed target")
	}
	for _, target := range success.Targets {
		want := 1
		if target.AssetID == second {
			want = 2
		}
		if m.sent(target.AssetID) != want {
			t.Fatal("retry changed another target", target.AssetID)
		}
	}
}
