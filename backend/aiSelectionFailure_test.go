package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAISelectionFailureEnvelopeIsSanitizedAndRetryable(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		retry  bool
	}{
		{context.DeadlineExceeded, 503, true}, {context.Canceled, 503, true},
		{errAISelectionCapacity, 503, true}, {errors.New("secret-private-path"), 500, true},
		{errAISelectionStale, 409, false}, {errAISelectionOwnerQuota, 429, false},
	} {
		rec := httptest.NewRecorder()
		writeAISelectionFailure(rec, tc.err)
		var result struct {
			Retryable                bool
			Code, Message, RequestID string
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if rec.Code != tc.status || result.Retryable != tc.retry || result.Code == "" || result.Message == "" || result.RequestID == "" || strings.Contains(rec.Body.String(), "secret-private") {
			t.Fatalf("unsafe failure: %d %s", rec.Code, rec.Body.String())
		}
	}
}
