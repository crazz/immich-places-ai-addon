package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIWritePreviewConflictUsesProtectedErrorEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	writeAIWritePreviewFailure(rec, writepreview.Failure{Code: "IMMICH_CONFLICT", Conflict: &writepreview.Conflict{}})
	var body map[string]json.RawMessage
	if json.Unmarshal(rec.Body.Bytes(), &body) != nil || rec.Code != 409 {
		t.Fatal("invalid error", rec.Code, rec.Body.String())
	}
	var id string
	if json.Unmarshal(body["requestID"], &id) != nil {
		t.Fatal("missing protected requestID")
	}
	if _, err := uuid.Parse(id); err != nil || string(body["retryable"]) != "false" || len(body["conflict"]) == 0 || len(body["message"]) == 0 {
		t.Fatal("invalid protected error envelope")
	}
}
