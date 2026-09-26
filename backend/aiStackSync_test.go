package main

import (
	"context"
	"sync"
	"testing"
)

func TestAIStackDrainsOlderSyncBeforePublishingEachVerifiedTarget(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	m := stackMutationServer(t, f)
	canceled, drained := make(chan struct{}), make(chan struct{})
	var once sync.Once
	if _, ok := f.writer.sync.tryStartUserSync(testUserID, func() { once.Do(func() { close(canceled) }) }); !ok {
		t.Fatal("sync fixture")
	}
	go func() {
		<-canceled
		for _, target := range op.Targets {
			if _, err := f.f.db.db.Exec(`UPDATE assets SET latitude=41,longitude=42 WHERE userID=? AND immichID=?`, testUserID, target.AssetID); err != nil {
				t.Error(err)
			}
		}
		f.writer.sync.clearUserCancel(testUserID)
		f.writer.sync.releaseUserSyncLock(testUserID)
		close(drained)
	}()
	for range op.Targets {
		if err := f.writer.runOne(context.Background(), testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	<-drained
	saved, err := f.writer.get(context.Background(), testUserID, op.ID, false)
	if err != nil || !saved.Refreshed || !saved.Verified {
		t.Fatal("target publication did not complete", err)
	}
	for _, target := range saved.Targets {
		var lat, lon float64
		if err := f.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, target.AssetID).Scan(&lat, &lon); err != nil || lat != 0 || lon != 12 {
			t.Fatal("older sync replaced verified GPS", err)
		}
		if m.sent(target.AssetID) != 1 {
			t.Fatal("sync repair resent mutation")
		}
	}
	if f.writer.sync.isUserSyncing(testUserID) {
		t.Fatal("publication leaked sync drain lock")
	}
}
