package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAIMirrorRecordIsStablePrivateAndUnverifiedUntilPublication(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	id, owned, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID)
	parsed, parseErr := uuid.Parse(id)
	if err != nil || parseErr != nil || parsed.Version() != 4 || len(owned) != 0 {
		t.Fatal("new record is not an unverified random identity", err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	again, owned, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID)
	if err != nil || again != id || len(owned) != 0 {
		t.Fatal("identity did not survive reopen", err)
	}
	other, _, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.members[1])
	if err != nil || other == id {
		t.Fatal("different asset reused identity", err)
	}
	if _, _, err := f.writer.drafts.mirrorRecord(ctx, uuid.NewString(), f.draft.AssetID); err == nil {
		t.Fatal("foreign owner accessed record")
	}
	var count int
	if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_mirror_records`).Scan(&count); err != nil || count != 2 {
		t.Fatal("foreign owner created a record", err)
	}
}
