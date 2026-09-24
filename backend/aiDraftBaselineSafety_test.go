package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
)

func draftBaselineFixture(t *testing.T) (*aiJobFixture, *aiImageFixture, *aiDraftStore, drafts.Draft, map[string]any) {
	t.Helper()
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	image := newAIImageFixture(t)
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
	key := "private-immich-image-key"
	if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
		t.Fatal(err)
	}
	meta := image.metadata()
	meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil}
	image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/assets/"+selectionA {
			return false
		}
		if err := json.NewEncoder(w).Encode(meta); err != nil {
			t.Error(err)
		}
		return true
	}
	return f, image, &aiDraftStore{results: &aiResultStore{jobs: f.store}}, value, meta
}
func TestAIDraftStaleBaselineCannotReplaceSavedDecision(t *testing.T) {
	for _, change := range []string{"gps", "source", "hidden", "key", "installation", "revision", "expired", "deleted", "storage"} {
		t.Run(change, func(t *testing.T) {
			f, image, store, value, meta := draftBaselineFixture(t)
			ctx := context.Background()
			observation, err := store.observe(ctx, testUserID, value.ID, 1, image.service)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "gps":
				meta["exifInfo"] = map[string]any{"latitude": 1, "longitude": 2}
			case "source":
				meta["updatedAt"] = "2026-09-20T12:01:00Z"
			case "hidden":
				selectionSQL(t, f.db, "UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
			case "key":
				key := "revoked"
				if err = f.db.updateImmichAPIKey(ctx, testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "installation":
				selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionB)
			case "revision":
				if _, err = store.edit(ctx, testUserID, value.ID, 1, drafts.Edit{State: "rejected"}); err != nil {
					t.Fatal(err)
				}
			case "expired":
				f.now = f.now.Add(5 * time.Minute)
			case "deleted":
				selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
			case "storage":
				selectionSQL(t, f.db, `CREATE TRIGGER draft_fail BEFORE INSERT ON ai_draft_revisions BEGIN SELECT RAISE(ABORT,'private storage detail'); END`)
			}
			if _, err = store.acknowledge(ctx, testUserID, value.ID, 1, observation.ID, image.service); err == nil {
				t.Fatal("stale observation accepted")
			}
			var count int
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_draft_revisions WHERE draftID=?", value.ID).Scan(&count); err != nil {
				t.Fatal(err)
			}
			expected := 1
			if change == "revision" {
				expected = 2
			}
			if change == "deleted" {
				expected = 0
			}
			if count != expected {
				t.Fatal("failed baseline created revision", count)
			}
		})
	}
}

func TestAIDraftObservationStorageIsBounded(t *testing.T) {
	f, image, store, value, _ := draftBaselineFixture(t)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if _, err := store.observe(ctx, testUserID, value.ID, 1, image.service); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.observe(ctx, testUserID, value.ID, 1, image.service); err == nil {
		t.Fatal("unbounded active observations")
	}
	f.now = f.now.Add(5 * time.Minute)
	if _, err := store.observe(ctx, testUserID, value.ID, 1, image.service); err != nil {
		t.Fatal("expired observations prevented review", err)
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_draft_baseline_observations WHERE draftID=?", value.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("expired observations retained", count, err)
	}
}

func TestAIDraftBaselineUsesItsExistingSQLiteConnection(t *testing.T) {
	f, image, store, value, _ := draftBaselineFixture(t)
	f.db.db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	observation, err := store.observe(ctx, testUserID, value.ID, 1, image.service)
	if err != nil {
		t.Fatal("baseline blocked on nested pool acquisition", err)
	}
	if _, err = store.acknowledge(ctx, testUserID, value.ID, 1, observation.ID, image.service); err != nil {
		t.Fatal(err)
	}
}

func TestAIDraftAccountDeletionFencesInflightObservation(t *testing.T) {
	f, image, store, value, meta := draftBaselineFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		close(entered)
		<-release
		_ = json.NewEncoder(w).Encode(meta)
		return true
	}
	done := make(chan error, 1)
	go func() {
		_, err := store.observe(context.Background(), testUserID, value.ID, 1, image.service)
		done <- err
	}()
	<-entered
	selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
	close(release)
	if err := <-done; err == nil {
		t.Fatal("late observation published")
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_draft_baseline_observations").Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
}

func TestAIDraftBaselineAcknowledgementInvalidatesStaging(t *testing.T) {
	_, image, store, value, _ := draftBaselineFixture(t)
	fields := json.RawMessage(`["gps"]`)
	ctx := context.Background()
	staged, err := store.edit(ctx, testUserID, value.ID, 1, drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":0}`), Fields: fields, State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	observation, err := store.observe(ctx, testUserID, value.ID, staged.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.acknowledge(ctx, testUserID, value.ID, staged.Revision, observation.ID, image.service)
	if err != nil || updated.State != "draft" || updated.Revision != 3 {
		t.Fatal("baseline revision retained stage", updated.State, updated.Revision, err)
	}
	saved, err := store.get(ctx, testUserID, value.ID)
	if err != nil || saved.State != "draft" {
		t.Fatal(saved.State, err)
	}
}
