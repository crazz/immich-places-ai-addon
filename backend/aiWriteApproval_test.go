package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"strings"
	"testing"
	"time"
)

func TestAIWriteDisabledKeepsLocalPreviewReadable(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	p := savedWritePreview(t, f, image, draft)
	h := writePreviewHandler(f, image)
	rec := aiRequest(h, "POST", "/ai/write-operations", `{"previewId":"`+p.Plan.ID+`","digest":"`+p.Digest+`","idempotencyKey":"disabled-key"}`, aiTestOrigin, true)
	if rec.Code != 503 || !strings.Contains(rec.Body.String(), "WRITE_DISABLED") {
		t.Fatal(rec.Code, rec.Body.String())
	}
	if rec = aiRequest(h, "GET", "/ai/write-previews/"+p.Plan.ID, "", "", true); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
}

func TestAIWriteConfirmationSurvivesReopenAndConsumesExactPlan(t *testing.T) {
	f, image, store, draft, _ := writePreviewFixture(t)
	p := savedWritePreview(t, f, image, draft)
	results := store.results
	results.origin = aiTestOrigin
	results.writer = &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2"}
	body := `{"previewId":"` + p.Plan.ID + `","digest":"` + p.Digest + `","idempotencyKey":"confirm-key"}`
	h := newAIResultHandler(results, image.service)
	rec := aiRequest(h, "POST", "/ai/write-operations", body, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	var first map[string]any
	if json.Unmarshal(rec.Body.Bytes(), &first) != nil || first["status"] != "queued" {
		t.Fatal(rec.Body.String())
	}
	var protected, operations, events int
	if err := f.db.db.QueryRow(`SELECT protected FROM ai_write_previews WHERE id=?`, p.Plan.ID).Scan(&protected); err != nil || protected != 1 {
		t.Fatal(protected, err)
	}
	f.reopen(t)
	store.results.jobs = f.store
	h = newAIResultHandler(results, image.service)
	for _, path := range []string{"/ai/write-operations/" + first["id"].(string), "/ai/write-operations/by-key/confirm-key"} {
		got := aiRequest(h, "GET", path, "", "", true)
		var recovered map[string]any
		if got.Code != 200 || json.Unmarshal(got.Body.Bytes(), &recovered) != nil || recovered["id"] != first["id"] {
			t.Fatal(got.Code, got.Body.String())
		}
	}
	duplicate := aiRequest(h, "POST", "/ai/write-operations", body, aiTestOrigin, true)
	if duplicate.Code != 200 || duplicate.Body.String() != rec.Body.String() {
		t.Fatal(duplicate.Code, duplicate.Body.String())
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_operations").Scan(&operations); err != nil || operations != 1 {
		t.Fatal(operations, err)
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_events").Scan(&events); err != nil || events != 1 {
		t.Fatal(events, err)
	}
}

func TestAIWriteRejectsSubstitutedStaleOrConsumedApproval(t *testing.T) {
	for _, kind := range []string{"digest", "expired", "edited", "rejected", "foreign", "corrupt", "substituted", "consumed", "reused-key", "unsupported", "unknown", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			f, image, store, draft, _ := writePreviewFixture(t)
			p := savedWritePreview(t, f, image, draft)
			store.results.origin = aiTestOrigin
			store.results.writer = &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2"}
			h := newAIResultHandler(store.results, image.service)
			body := `{"previewId":"` + p.Plan.ID + `","digest":"` + p.Digest + `","idempotencyKey":"key"}`
			expected := 409
			switch kind {
			case "digest":
				body = strings.ReplaceAll(body, p.Digest, strings.Repeat("a", 64))
			case "expired":
				f.now = f.now.Add(5 * time.Minute)
				expected = 410
			case "edited", "rejected":
				state := "draft"
				if kind == "rejected" {
					state = "rejected"
				}
				if _, err := store.edit(context.Background(), testUserID, draft.ID, draft.Revision, drafts.Edit{State: state}); err != nil {
					t.Fatal(err)
				}
			case "foreign":
				body = strings.ReplaceAll(body, p.Plan.ID, selectionB)
				expected = 404
			case "corrupt", "substituted":
				selectionSQL(t, f.db, "DROP TRIGGER ai_write_previews_immutable")
				if kind == "corrupt" {
					selectionSQL(t, f.db, "UPDATE ai_write_previews SET payload='{}' WHERE id=?", p.Plan.ID)
				} else {
					raw, _ := json.Marshal(p.Plan)
					selectionSQL(t, f.db, "UPDATE ai_write_previews SET payload=? WHERE id=?", strings.ReplaceAll(string(raw), draft.AssetID, selectionB), p.Plan.ID)
				}
				expected = 503
			case "consumed", "reused-key":
				if rec := aiRequest(h, "POST", "/ai/write-operations", body, aiTestOrigin, true); rec.Code != 200 {
					t.Fatal(rec.Code)
				}
				if kind == "consumed" {
					body = strings.ReplaceAll(body, `"key"`, `"another-key"`)
				} else {
					body = strings.ReplaceAll(body, p.Digest, strings.Repeat("b", 64))
				}
			case "unsupported":
				store.results.writer.profile = "unknown"
				expected = 503
			case "unknown":
				body = strings.TrimSuffix(body, "}") + `,"latitude":1}`
				expected = 400
			case "oversize":
				body = strings.Repeat(" ", 4097) + body
				expected = 400
			}
			rec := aiRequest(h, "POST", "/ai/write-operations", body, aiTestOrigin, true)
			if rec.Code != expected {
				t.Fatal(kind, rec.Code, rec.Body.String())
			}
			var count int
			want := 0
			if kind == "consumed" || kind == "reused-key" {
				want = 1
			}
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_operations").Scan(&count); err != nil || count != want {
				t.Fatal(count, err)
			}
		})
	}
}

func TestAIWriteApprovalFailureRollsBackAllAuthority(t *testing.T) {
	f, image, store, draft, _ := writePreviewFixture(t)
	p := savedWritePreview(t, f, image, draft)
	writer := &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2"}
	selectionSQL(t, f.db, `CREATE TRIGGER fail_write_audit BEFORE INSERT ON ai_write_events BEGIN SELECT RAISE(ABORT,'synthetic storage failure'); END`)
	_, err := writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: p.Plan.ID, Digest: p.Digest, Key: "key"})
	if err == nil {
		t.Fatal("approval committed without audit")
	}
	f.reopen(t)
	store.results.jobs = f.store
	for _, table := range []string{"ai_write_operations", "ai_write_targets", "ai_write_events", "ai_write_target_guards"} {
		var n int
		if err := f.db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
			t.Fatal(table, n, err)
		}
	}
	var protected int
	if err := f.db.db.QueryRow("SELECT protected FROM ai_write_previews WHERE id=?", p.Plan.ID).Scan(&protected); err != nil || protected != 0 {
		t.Fatal(protected, err)
	}
}
