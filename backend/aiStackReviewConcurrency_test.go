package main

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestAIStackReviewBoundsConcurrentReadsAndHonorsCancellation(t *testing.T) {
	f := stackWriteFixture(t)
	for i := 4; i <= 8; i++ {
		id := selectionID(i)
		seedAsset(t, f.f.db, id, nil, nil, "2026-09-20")
		if _, err := f.f.db.db.Exec(`UPDATE assets SET stackID=?,stackPrimaryAssetID=? WHERE userID=? AND immichID=?`, f.stackID, f.draft.AssetID, testUserID, id); err != nil {
			t.Fatal(err)
		}
		meta := f.image.metadata()
		meta["id"] = id
		f.members = append(f.members, id)
		f.metadata[id] = meta
	}
	for _, meta := range f.metadata {
		meta["stack"] = map[string]any{"id": f.stackID, "primaryAssetId": f.draft.AssetID, "assetCount": len(f.members)}
	}
	arrived := make(chan struct{}, len(f.members))
	var active, maximum, primaryReads atomic.Int32
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path == "/api/stacks/"+f.stackID {
			return false
		}
		if r.URL.Path == "/api/assets/"+f.draft.AssetID && primaryReads.Add(1) == 1 {
			return false
		}
		current := active.Add(1)
		defer active.Add(-1)
		for old := maximum.Load(); current > old; old = maximum.Load() {
			if maximum.CompareAndSwap(old, current) {
				break
			}
		}
		arrived <- struct{}{}
		<-r.Context().Done()
		return true
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		_, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
		finished <- err
	}()
	for range 4 {
		select {
		case <-arrived:
		case <-time.After(3 * time.Second):
			t.Fatal("four independent reads did not reach barrier")
		}
	}
	cancel()
	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("canceled review became usable")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("canceled reads did not stop")
	}
	if maximum.Load() > 4 {
		t.Fatal("metadata fan-out exceeded four", maximum.Load())
	}
	var rows int
	if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_reviews`).Scan(&rows); err != nil || rows != 0 {
		t.Fatal("canceled observation persisted", rows, err)
	}
}
