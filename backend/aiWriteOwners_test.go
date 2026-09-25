package main

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIWriteConcurrentOwnersShareTargetExclusionWithoutPrivateDisclosure(t *testing.T) {
	f, image, store, draft, _ := writePreviewFixture(t)
	first := savedWritePreview(t, f, image, draft)
	ctx := context.Background()
	analysis, err := f.store.ReadAnalysis(ctx, testUserID, draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	other := "other-writer"
	if err = f.db.createUser(ctx, other, "other-writer@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	key := "private-immich-image-key"
	if err = f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
		t.Fatal(err)
	}
	if err = f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: selectionA, Type: "IMAGE", OriginalFileName: "synthetic.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f.input.Owner = other
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	job := completeAIJob(t, f, "other-write-plan")
	var analysisID string
	if err = f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", other, job.ID).Scan(&analysisID); err != nil {
		t.Fatal(err)
	}
	secondDraft, err := store.accept(ctx, other, analysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := store.observe(ctx, other, secondDraft.ID, secondDraft.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	secondDraft, err = store.acknowledge(ctx, other, secondDraft.ID, secondDraft.Revision, observed.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	secondDraft, err = store.edit(ctx, other, secondDraft.ID, secondDraft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":1,"longitude":2}`), Fields: json.RawMessage(`["gps"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: image.service}
	second, err := writepreview.Create(ctx, session, session, session, other, secondDraft.ID, secondDraft.Revision, uuid.NewString(), f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	writer := &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2"}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wins := make(chan writeback.Operation, 2)
	losses := make(chan error, 2)
	for _, preview := range []writepreview.Preview{first, second} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			op, err := writer.confirm(ctx, preview.Plan.Owner, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "same-key"})
			if err != nil {
				losses <- err
			} else {
				wins <- op
			}
		}()
	}
	close(start)
	wg.Wait()
	close(wins)
	close(losses)
	if len(wins) != 1 || len(losses) != 1 {
		t.Fatal("target was not serialized", len(wins), len(losses))
	}
	if err := <-losses; err.Error() != "TARGET_BUSY" {
		t.Fatal("non-generic contention", err)
	}
	winner := <-wins
	loser := other
	if winner.Plan.Owner == other {
		loser = testUserID
	}
	if _, err = writer.get(ctx, loser, winner.ID, false); err == nil {
		t.Fatal("foreign outcome disclosed")
	}
	if !writer.enter(winner) {
		t.Fatal("winner not admitted")
	}
	independent := winner
	independent.Plan.Owner = loser
	independent.Plan.TargetID = selectionB
	if !writer.enter(independent) {
		t.Fatal("unrelated target blocked")
	}
	third := independent
	third.Plan.Owner = "third"
	third.Plan.TargetID = uuid.NewString()
	if writer.enter(third) {
		t.Fatal("global bound exceeded")
	}
	writer.leave(winner)
	writer.leave(independent)
}
