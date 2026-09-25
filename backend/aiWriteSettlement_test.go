package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAIWriteReportsObservedGPSSeparatelyFromSenderSettlement(t *testing.T) {
	w := newAIWriteFixture(t)
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		w.mu.Lock()
		w.meta["exifInfo"] = map[string]float64{"latitude": 0, "longitude": 12}
		w.mu.Unlock()
		conn, _, _ := out.(http.Hijacker).Hijack()
		conn.Close()
		return true
	}
	w.run(t)
	rec := aiRequest(newAIResultHandler(w.writer.drafts.results, w.image.service), "GET", "/ai/write-operations/"+w.op.ID, "", "", true)
	var body map[string]any
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &body) != nil || body["verified"] != true || body["settled"] != false {
		t.Fatal("sender settlement hidden", rec.Body.String())
	}
}
