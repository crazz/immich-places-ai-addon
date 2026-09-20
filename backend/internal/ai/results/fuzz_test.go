package results

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func FuzzValidateBoundedResult(f *testing.F) {
	validator, err := New()
	if err != nil {
		f.Fatal(err)
	}
	for _, name := range []string{"unknown", "synthetic"} {
		data, err := fixtureFS.ReadFile("testdata/ai-analysis-result." + name + ".json")
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data, name == "synthetic", true)
		f.Add(data, name == "synthetic", false)
	}
	for _, data := range [][]byte{[]byte(`{"private-sensitive-key":"https://private.invalid"}`), []byte(`{"x":1,"x":2}`), []byte(`1e999999999`), {0xff}, []byte(`"\uD800"`)} {
		f.Add(data, false, true)
	}
	f.Fuzz(func(t *testing.T, data []byte, withUkrainian, complete bool) {
		ctx := visualContext("en")
		if withUkrainian {
			ctx.Languages = append(ctx.Languages, "uk")
		}
		if !complete {
			ctx.Completion = Truncated
		}
		proposal, err := validator.Validate(data, ctx)
		if err != nil {
			var rejected *Failure
			if !errors.As(err, &rejected) {
				t.Fatal("unsanitized failure type", err)
			}
			encoded, encodeErr := json.Marshal(rejected)
			if encodeErr != nil || len(encoded) > 4096 || len(rejected.Findings) > 20 || strings.Contains(string(encoded), "private-sensitive") {
				t.Fatal("unbounded/private diagnostics")
			}
			if _, err := proposal.MarshalJSON(); err == nil {
				t.Fatal("partial result on failure")
			}
			return
		}
		if !complete {
			t.Fatal("incomplete attempt accepted")
		}
		encoded, err := proposal.MarshalJSON()
		if err != nil || !json.Valid(encoded) {
			t.Fatal("invalid successful serialization", err)
		}
		again, err := validator.Validate(encoded, ctx)
		if err != nil {
			t.Fatal("validated output does not revalidate", err)
		}
		roundtrip, err := again.MarshalJSON()
		if err != nil || !bytes.Equal(encoded, roundtrip) {
			t.Fatal("normalization is not idempotent", err)
		}
	})
}
