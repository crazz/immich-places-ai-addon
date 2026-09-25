package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIWritePreviewPrivateReferencesCannotSubstituteScope(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	preview := savedWritePreview(t, f, image, draft)
	ctx := context.Background()
	if err := f.db.createUser(ctx, "other-preview-owner", "preview-other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte("other-preview-session"))
	if err := f.db.createSession(ctx, hex.EncodeToString(hash[:]), "other-preview-owner", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	h := writePreviewHandler(f, image)
	for _, id := range []string{preview.Plan.ID, "missing"} {
		req := httptest.NewRequest("GET", "/ai/write-previews/"+id, nil)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "other-preview-session"})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != 404 || strings.Contains(rec.Body.String(), draft.ID) || strings.Contains(rec.Body.String(), draft.AssetID) || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("foreign reference disclosed", rec.Code, rec.Body.String())
		}
	}
	for _, extra := range []string{`"targetId":"other"`, `"intended":{"latitude":7,"longitude":8}`, `"fields":["description"]`, `"digest":"` + preview.Digest + `"`, `"plan":{}`} {
		body := `{"draftId":"` + draft.ID + `","draftRevision":3,` + extra + `}`
		if rec := aiRequest(h, "POST", "/ai/write-previews", body, aiTestOrigin, true); rec.Code != 400 {
			t.Fatal("scope substitution accepted", rec.Code, rec.Body.String())
		}
	}
	if rec := aiRequest(h, "POST", "/ai/write-previews", `{}`, "", true); rec.Code != 403 {
		t.Fatal("origin missing", rec.Code)
	}
	if rec := aiRequest(h, "GET", "/ai/write-previews/"+preview.Plan.ID, "", "", false); rec.Code != 401 {
		t.Fatal("anonymous preview", rec.Code)
	}
	var restored writepreview.Preview
	rec := aiRequest(h, "GET", "/ai/write-previews/"+preview.Plan.ID, "", "", true)
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &restored) != nil || restored.Digest != preview.Digest {
		t.Fatal("substitution changed stored plan")
	}
}

func TestAIWritePreviewLifecycleKeepsOtherOwnersAndDisabledReads(t *testing.T) {
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
	const other = "other-preview-owner"
	if err = f.db.createUser(ctx, other, "preview-other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	key := "private-immich-image-key"
	if err = f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
		t.Fatal(err)
	}
	if err = f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: selectionA, Type: "IMAGE", OriginalFileName: "synthetic.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
		t.Fatal(err)
	}
	f.input.Owner = other
	job := completeAIJob(t, f, "other-preview")
	var id string
	if err = f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", other, job.ID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	value, err := store.accept(ctx, other, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	observation, err := store.observe(ctx, other, value.ID, value.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.acknowledge(ctx, other, value.ID, value.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.edit(ctx, other, value.ID, value.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":0}`), Fields: json.RawMessage(`["gps"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	f.store.enabled = false
	session := &aiWritePreviewSession{drafts: store, images: image.service}
	second, err := writepreview.Create(ctx, session, session, session, other, value.ID, value.Revision, selectionB, f.store.now)
	if err != nil {
		t.Fatal("disabled execution blocked preview", err)
	}
	selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
	if _, err = session.get(ctx, testUserID, first.Plan.ID); err == nil {
		t.Fatal("deleted preview survived")
	}
	if _, err = session.get(ctx, other, second.Plan.ID); err != nil {
		t.Fatal("other preview lost", err)
	}
	selectionSQL(t, f.db, "DELETE FROM assets WHERE userID=?", other)
	if _, err = session.get(ctx, other, second.Plan.ID); err != nil {
		t.Fatal("catalog reset removed stored comparison", err)
	}
	selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionID(999))
	if _, err = session.get(ctx, other, second.Plan.ID); err == nil {
		t.Fatal("old installation usable")
	}
	session.drafts = &aiDraftStore{results: &aiResultStore{jobs: newAIJobStore(f.db, selectionID(999), false, f.store.now)}}
	if _, err = session.get(ctx, other, second.Plan.ID); err == nil {
		t.Fatal("new installation adopted old preview")
	}
	var count int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews WHERE userID=?", testUserID).Scan(&count); err != nil || count != 0 {
		t.Fatal("deleted private rows retained", count, err)
	}
}
