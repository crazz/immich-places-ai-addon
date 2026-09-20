package main

import (
	"encoding/json"
	"testing"
)

func TestAIProductionHTTPSessionStorageFailureIsRetryableAndSafe(t *testing.T) {
	f, p, _ := productionFixture(t)
	productionSession(t, f)
	h := newAIJobHandler(p, aiTestOrigin)
	f.image.db.close()
	rec := aiRequest(h, "GET", "/ai/jobs", "", "", true)
	var body struct {
		Code      string
		Retryable bool
	}
	if rec.Code != 500 || json.Unmarshal(rec.Body.Bytes(), &body) != nil || body.Code != "STORAGE_ERROR" || !body.Retryable {
		t.Fatal("storage failure misclassified", rec.Code, rec.Body.String())
	}
}
