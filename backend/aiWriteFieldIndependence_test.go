package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardMalformedGPSDoesNotHideDescriptionEvidence(t *testing.T) {
	for _, fields := range [][]string{{"description"}, {"gps", "description"}} {
		t.Run(fields[0], func(t *testing.T) {
			w := standardWriteFixture(t, fields)
			ctx := context.Background()
			op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "independent-fields"})
			if err != nil {
				t.Fatal(err)
			}
			attempt := &aiWriteAttempt{store: w.writer}
			w.meta["exifInfo"] = map[string]any{"latitude": "invalid", "longitude": 12, "description": op.Plan.Description.Before.Value}
			fresh, err := attempt.Read(ctx, op)
			if err != nil {
				t.Fatal("valid description hidden by malformed GPS", err)
			}
			code := writeback.CompareBefore(op.Plan, fresh)
			if len(fields) == 1 && code != "" || len(fields) == 2 && code != "SOURCE_UNAVAILABLE" {
				t.Fatal("selected-field baseline decision", code)
			}
			w.meta["exifInfo"] = map[string]any{"latitude": 1000, "longitude": 12, "description": op.Plan.Description.Intended}
			fresh, err = attempt.Read(ctx, op)
			if err != nil {
				t.Fatal(err)
			}
			decision := writeback.Readback(op, fresh, true)
			last := decision.Fields[len(decision.Fields)-1]
			if last.Field != "description" || last.Status != "verified" || !last.WasVerified {
				t.Fatal("description success discarded", decision)
			}
			if len(fields) == 1 && !decision.Verified || len(fields) == 2 && (decision.Status != "verifying" || decision.Fields[0].Status != "unavailable") {
				t.Fatal("malformed selected GPS treated as valid", decision)
			}
		})
	}
}
