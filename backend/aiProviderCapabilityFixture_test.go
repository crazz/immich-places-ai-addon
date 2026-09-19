package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

func capabilityFixtureResponse(w http.ResponseWriter, r *http.Request, hits *atomic.Int32) {
	if hits != nil {
		hits.Add(1)
	}
	body, _ := io.ReadAll(r.Body)
	content := "color: blue\nshape: circle"
	if strings.Contains(string(body), `"json_object"`) || strings.Contains(string(body), `"json_schema"`) {
		content = `{"color":"blue","shape":"circle"}`
	}
	contentJSON, _ := json.Marshal(content)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":` + string(contentJSON) + `},"finish_reason":"stop"}]}`))
}
