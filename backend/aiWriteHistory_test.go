package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIWritePrivateHistorySurvivesReloadAndRejectsForeignDraft(t *testing.T) {
	w := newAIWriteFixture(t)
	h := newAIResultHandler(w.writer.drafts.results, w.image.service)
	rec := aiRequest(h, "GET", "/ai/write-operations?draftId="+w.draft.ID, "", "", true)
	var page struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &page) != nil || len(page.Items) != 1 || page.Items[0].ID != w.op.ID {
		t.Fatal(rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), w.preview.Digest) || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private summary boundary")
	}
	if rec = aiRequest(h, "GET", "/ai/write-operations?draftId="+selectionB, "", "", true); rec.Code != 404 {
		t.Fatal(rec.Code)
	}
}
