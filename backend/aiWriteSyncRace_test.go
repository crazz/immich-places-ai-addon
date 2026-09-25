package main

import (
	"context"
	"sync"
	"testing"
)

func TestAIWriteDrainsStaleSyncBeforePublishingVerifiedGPS(t *testing.T) {
	w := newAIWriteFixture(t)
	canceled, drained := make(chan struct{}), make(chan struct{})
	var once sync.Once
	if _, ok := w.writer.sync.tryStartUserSync(testUserID, func() { once.Do(func() { close(canceled) }) }); !ok {
		t.Fatal("sync fixture")
	}
	go func() {
		<-canceled
		_, err := w.f.db.db.ExecContext(context.Background(), "UPDATE assets SET latitude=41,longitude=42 WHERE userID=? AND immichID=?", testUserID, w.draft.AssetID)
		if err != nil {
			t.Error(err)
		}
		w.writer.sync.clearUserCancel(testUserID)
		w.writer.sync.releaseUserSyncLock(testUserID)
		close(drained)
	}()
	w.run(t)
	<-drained
	var lat, lon float64
	if err := w.f.db.db.QueryRow("SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?", testUserID, w.draft.AssetID).Scan(&lat, &lon); err != nil || lat != 0 || lon != 12 {
		t.Fatal("stale sync won", lat, lon, err)
	}
	if op := w.status(t); !op.Refreshed || !op.Verified {
		t.Fatal(op)
	}
	if w.writer.sync.isUserSyncing(testUserID) {
		t.Fatal("drain lock leaked")
	}
}
