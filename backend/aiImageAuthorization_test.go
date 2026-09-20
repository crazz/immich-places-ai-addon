package main

import (
	"context"
	"testing"
)

func TestAIImageLocalDenialsNeverReachImmich(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		change                     func(*testing.T, *aiImageFixture)
		owner, installation, asset string
	}{
		{name: "disabled", change: func(t *testing.T, f *aiImageFixture) { f.store.enabled = false }},
		{name: "missing account", owner: "absent"},
		{name: "wrong installation", installation: selectionID(90)},
		{name: "malformed installation", installation: "../instance"},
		{name: "missing asset", asset: selectionB},
		{name: "source URL", asset: "http://private.example/image"},
		{name: "path traversal", asset: "../settings"},
		{name: "missing key", change: func(t *testing.T, f *aiImageFixture) {
			selectionSQL(t, f.db, "UPDATE users SET immichAPIKey=NULL WHERE ID=?", testUserID)
		}},
		{name: "empty key", change: func(t *testing.T, f *aiImageFixture) {
			key := ""
			if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "hidden", change: func(t *testing.T, f *aiImageFixture) { selectionSQL(t, f.db, "UPDATE assets SET isHidden=1") }},
		{name: "video", change: func(t *testing.T, f *aiImageFixture) { selectionSQL(t, f.db, "UPDATE assets SET type='VIDEO'") }},
		{name: "stack child", change: func(t *testing.T, f *aiImageFixture) {
			selectionSQL(t, f.db, "UPDATE assets SET stackPrimaryAssetID=?", selectionB)
		}},
		{name: "hidden library", change: func(t *testing.T, f *aiImageFixture) {
			selectionSQL(t, f.db, "INSERT INTO libraries (libraryID,isHidden) VALUES ('private',1)")
			selectionSQL(t, f.db, "UPDATE assets SET libraryID='private'")
		}},
		{name: "foreign owner", change: func(t *testing.T, f *aiImageFixture) {
			if err := f.db.createUser(context.Background(), "other", "other@example.com", "hash"); err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, f.db, "UPDATE assets SET userID='other'")
		}},
		{name: "persisted installation replaced", change: func(t *testing.T, f *aiImageFixture) {
			selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionID(99))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAIImageFixture(t)
			if tc.change != nil {
				tc.change(t, f)
			}
			owner, installation, asset := testUserID, f.store.binding, selectionA
			if tc.owner != "" {
				owner = tc.owner
			}
			if tc.installation != "" {
				installation = tc.installation
			}
			if tc.asset != "" {
				asset = tc.asset
			}
			p, err := f.service.prepare(context.Background(), owner, installation, asset)
			calls, _ := f.counts()
			if err == nil || p != nil || calls != 0 {
				t.Fatalf("denial leaked to upstream: calls=%d err=%v", calls, err)
			}
		})
	}
}
