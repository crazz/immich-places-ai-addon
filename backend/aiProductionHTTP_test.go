package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

func productionSession(t *testing.T, f *aiVisualFixture) {
	t.Helper()
	hash := sha256.Sum256([]byte("ai-session"))
	if err := f.image.db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
}
func TestAIProductionHTTPSubmitReadAndDisabledCancel(t *testing.T) {
	f, p, req := productionFixture(t)
	productionSession(t, f)
	h := newAIJobHandler(p, aiTestOrigin)
	data, _ := json.Marshal(req)
	rec := aiRequest(h, "POST", "/ai/jobs", string(data), aiTestOrigin, true)
	if rec.Code != 202 {
		t.Fatalf("admit %d %s", rec.Code, rec.Body.String())
	}
	var job struct {
		ID     string         `json:"id"`
		Counts map[string]int `json:"counts"`
		Usage  struct {
			ReportedStatus string `json:"reportedStatus"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &job); err != nil || job.ID == "" || job.Counts["queued"] != 1 || job.Usage.ReportedStatus != "unknown" {
		t.Fatal("incomplete progress", rec.Body.String(), err)
	}
	p.store.enabled = false
	rec = aiRequest(h, "GET", "/ai/jobs/"+job.ID, "", "", true)
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("disabled read failed", rec.Code, rec.Body.String())
	}
	for i := 0; i < 2; i++ {
		rec = aiRequest(h, "POST", "/ai/jobs/"+job.ID+"/cancel", "{}", aiTestOrigin, true)
		if rec.Code != 200 {
			t.Fatal("cancel failed", rec.Code, rec.Body.String())
		}
	}
	rec = aiRequest(h, "GET", "/ai/jobs/"+job.ID, "", "", true)
	if err := json.Unmarshal(rec.Body.Bytes(), &job); err != nil || job.Counts["canceled"] != 1 {
		t.Fatal("cancel not durable", rec.Body.String())
	}
	if f.hits.Load() != 0 {
		t.Fatal("progress dispatched work")
	}
}
