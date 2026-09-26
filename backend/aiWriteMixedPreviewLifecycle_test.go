package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIStandardPreviewQuotaAndExpiryAreSharedAcrossVersionsAndReopen(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	session := &aiWritePreviewSession{drafts: w.writer.drafts, images: w.image.service}
	snapshot, err := session.ReadDraft(ctx, testUserID, w.draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	current, err := session.ReadMetadata(ctx, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Description, snapshot.Fields = nil, []string{"gps"}
	var retained writepreview.Preview
	for range 5 {
		preview, raw, err := writepreview.Build(snapshot, current, uuid.NewString(), w.f.now)
		if err != nil {
			t.Fatal(err)
		}
		_, err = w.f.db.db.Exec(`INSERT INTO ai_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) VALUES(?,?,?,?,?,?,?,?,?,?)`, testUserID, w.f.store.binding, preview.Plan.ID, w.draft.ID, w.draft.Revision, preview.Plan.Version, string(raw), preview.Digest, w.f.now.UnixNano(), w.f.now.Add(5*time.Minute).UnixNano())
		if err != nil {
			t.Fatal(err)
		}
		retained = preview
	}
	for range 4 {
		if _, err := writepreview.Create(ctx, session, session, session, testUserID, w.draft.ID, w.draft.Revision, uuid.NewString(), w.f.store.now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := writepreview.Create(ctx, session, session, session, testUserID, w.draft.ID, w.draft.Revision, uuid.NewString(), w.f.store.now); err == nil || err.Error() != "PREVIEW_CAPACITY" {
		t.Fatal("mixed formats bypassed shared capacity", err)
	}
	selectionSQL(t, w.f.db, "UPDATE ai_write_previews SET protected=1 WHERE id=?", retained.Plan.ID)
	if _, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "protected-v2"}); err != nil {
		t.Fatal(err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.f.now = w.f.now.Add(5 * time.Minute)
	if _, err := writepreview.Create(ctx, session, session, session, testUserID, w.draft.ID, w.draft.Revision, uuid.NewString(), w.f.store.now); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_all_write_previews`).Scan(&count); err != nil || count != 3 {
		t.Fatal("expired mixed records not pruned or protected history removed", count, err)
	}
	for _, preview := range []writepreview.Preview{retained, w.preview} {
		stored, err := session.get(ctx, testUserID, preview.Plan.ID)
		if err != nil || stored.Digest != preview.Digest || stored.Status != "expired" || stored.Plan.ExpiresAt != preview.Plan.ExpiresAt {
			t.Fatal("retained comparison renewed or altered", err)
		}
	}
}
