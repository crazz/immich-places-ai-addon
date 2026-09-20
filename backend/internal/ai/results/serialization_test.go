package results

import (
	"bytes"
	"errors"
	"testing"
)

func TestSerializedProposalCannotExceedThePayloadBudget(t *testing.T) {
	data, ctx := maximumWorkload(t)
	// Valid literal prose can expand sixfold when JSON serialization escapes HTML.
	data = bytes.ReplaceAll(data, bytes.Repeat([]byte("w"), 8192), bytes.Repeat([]byte("<"), 8192))
	if len(data) != maxBytes {
		t.Fatal("fixture must exercise the exact raw-byte ceiling")
	}
	proposal, err := testValidator(t).Validate(data, ctx)
	var rejected *Failure
	if !errors.As(err, &rejected) || rejected.Category != "limit_exceeded" {
		t.Fatalf("serialized byte overflow must fail the whole proposal: %v", err)
	}
	if _, err := proposal.MarshalJSON(); err == nil {
		t.Fatal("serialization overflow returned a partial proposal")
	}
}
