package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorCleanupAndDeletionRetainOtherOwnersExactHistoryAndUnknownGuard(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "private-first"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	attempt := &aiMirrorAttempt{standard: &aiStackAttempt{store: f.writer, parent: op, assetID: f.draft.AssetID}}
	if _, err = attempt.read(ctx, false); err != nil {
		t.Fatal(err)
	}
	if err = attempt.reserve(ctx, false); err != nil {
		t.Fatal(err)
	}
	analysis, err := f.f.store.ReadAnalysis(ctx, testUserID, f.draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	const other = "mirror-other-owner"
	if err = f.f.db.createUser(ctx, other, "mirror-other@example.test", "hash"); err != nil {
		t.Fatal(err)
	}
	key := "private-immich-image-key"
	if err = f.f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
		t.Fatal(err)
	}
	if err = f.f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: selectionB, Type: "IMAGE", OriginalFileName: "other.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = f.f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f.f.input.Owner, f.f.input.AssetIDs, f.f.input.MaxCalls = other, []string{selectionB}, 3
	job := completeAIJob(t, f.f, "other-mirror-history")
	var analysisID string
	if err = f.f.db.db.QueryRow(`SELECT id FROM ai_analyses WHERE userID=? AND jobID=?`, other, job.ID).Scan(&analysisID); err != nil {
		t.Fatal(err)
	}
	otherMeta := f.image.metadata()
	otherMeta["id"] = selectionB
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		switch r.URL.Path {
		case "/api/assets/" + selectionB:
			_ = json.NewEncoder(w).Encode(otherMeta)
			return true
		case "/api/assets/" + selectionB + "/metadata":
			_, _ = w.Write([]byte(`[]`))
			return true
		}
		return false
	}
	store := f.writer.drafts
	d, err := store.accept(ctx, other, analysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	observation, err := store.observe(ctx, other, d.ID, d.Revision, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	d, err = store.acknowledge(ctx, other, d.ID, d.Revision, observation.ID, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	text := "Other owner's exact local text"
	d, err = store.edit(ctx, other, d.ID, d.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":1,"longitude":2}`), Fields: json.RawMessage(`["gps"]`), Descriptions: map[string]drafts.DescriptionEdit{"en": {Text: &text}}, Mirror: json.RawMessage(`{"languages":["en"]}`)})
	if err != nil {
		t.Fatal(err)
	}
	d, err = store.edit(ctx, other, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}
	preview, err := writepreview.Create(ctx, session, session, session, other, d.ID, d.Revision, uuid.NewString(), f.f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	otherOp, err := f.writer.confirm(ctx, other, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "private-other"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.f.store.Cancel(ctx, other, job.ID); err != nil {
		t.Fatal(err)
	}
	firstBefore, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	f.f.now = f.f.now.Add(2 * time.Hour)
	for _, owner := range []string{testUserID, other} {
		if n, err := f.f.store.PurgeBefore(ctx, owner, f.f.now.Add(-time.Hour), 100); err != nil || n != 0 {
			t.Fatal("cleanup removed retained mirror history", owner, n, err)
		}
	}
	afterCleanup, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !reflect.DeepEqual(firstBefore, afterCleanup) {
		t.Fatal("cleanup changed active metadata audit", err)
	}
	selectionSQL(t, f.f.db, `DELETE FROM users WHERE ID=?`, testUserID)
	retained, err := f.writer.get(ctx, other, otherOp.ID, false)
	if err != nil || !reflect.DeepEqual(otherOp, retained) {
		t.Fatal("deletion changed another owner's exact approval/outcome", err)
	}
	retainedDraft, err := store.get(ctx, other, d.ID)
	if err != nil || !reflect.DeepEqual(d, retainedDraft) {
		t.Fatal("deletion changed another owner's local content", err)
	}
	for _, table := range []string{"ai_mirror_write_steps", "ai_mirror_write_events", "ai_mirror_records"} {
		var count int
		if err = f.f.db.db.QueryRow(`SELECT count(*) FROM `+table+` WHERE userID=?`, testUserID).Scan(&count); err != nil || count != 0 {
			t.Fatal("deleted private metadata survived", table, err)
		}
	}
	var guards int
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 2 {
		t.Fatal("unknown or unrelated exclusion lost", guards, err)
	}
	if err = attempt.sent(ctx, writeback.Completion{Known: true}); err == nil {
		t.Fatal("late metadata completion recreated private data")
	}
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 1 {
		t.Fatal("late cleanup removed unrelated guard", guards, err)
	}
	retained, err = f.writer.get(ctx, other, otherOp.ID, false)
	if err != nil || !reflect.DeepEqual(otherOp, retained) {
		t.Fatal("late cleanup modified unrelated history", err)
	}
	if f.standardSends != 1 || f.mirrorSends != 0 {
		t.Fatal("cleanup or account deletion sent an upstream mutation")
	}
}
