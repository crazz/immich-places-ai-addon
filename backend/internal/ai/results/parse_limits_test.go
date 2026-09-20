package results

import (
	"errors"
	"strings"
	"testing"
)

func TestParseBoundsBytesDepthNodesAndStrings(t *testing.T) {
	for _, tc := range []struct {
		name        string
		exact, over []byte
	}{
		{"bytes", []byte(strings.Repeat(" ", (1<<20)-2) + `{}`), []byte(strings.Repeat(" ", (1<<20)-1) + `{}`)},
		{"depth", []byte(strings.Repeat("[", 32) + `0` + strings.Repeat("]", 32)), []byte(strings.Repeat("[", 33) + `0` + strings.Repeat("]", 33))},
		{"nodes", []byte(`[` + strings.Repeat(`0,`, 19998) + `0]`), []byte(`[` + strings.Repeat(`0,`, 19999) + `0]`)},
		{"string", []byte(`"` + strings.Repeat("x", 8192) + `"`), []byte(`"` + strings.Repeat("x", 8193) + `"`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parse(tc.exact); err != nil {
				t.Fatal("exact boundary rejected", err)
			}
			_, err := parse(tc.over)
			var failure *Failure
			if !errors.As(err, &failure) || failure.Category != "limit_exceeded" {
				t.Fatalf("over boundary accepted: %v", err)
			}
		})
	}
}

func TestClosedSchemaFieldsAreRequiredAndTyped(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) { delete(m, "selected_candidate_id") },
		func(m map[string]any) { m["approval"] = true },
		func(m map[string]any) { m["outcome"] = true },
		func(m map[string]any) { m["descriptions"].([]any)[0].(map[string]any)["text"] = 17 },
	} {
		requireFailure(t, v, mutateFixture(t, "unknown", change), visualContext("en"), "schema_violation")
	}

}

func TestNumbersAreBoundedWithoutLosingExactRanges(t *testing.T) {
	for _, number := range []string{strings.Repeat("1", 128), "0e324", "0e-324", "1e-323", "-0", "90.000000000000000000000000001"} {
		if _, err := parse([]byte(number)); err != nil {
			t.Fatalf("bounded token %s: %v", number, err)
		}
	}
	for _, number := range []string{strings.Repeat("1", 129), "0e325", "0e-325", "1e9999999999999999999", "1e324", "1e-324", "-1e-324"} {
		if _, err := parse([]byte(number)); err == nil {
			t.Fatalf("unbounded/underflow token accepted: %s", number)
		}
	}
	v := testValidator(t)
	for _, number := range []string{"90.000000000000000000000000001", "-90.000000000000000000000000001"} {
		data := []byte(strings.Replace(string(fixture(t, "synthetic")), `"latitude": 50.0`, `"latitude": `+number, 1))
		requireFailure(t, v, data, visualContext("en", "uk"), "schema_violation")
	}
}

func TestCollectionsAndIdentifiersHaveIndividualBounds(t *testing.T) {
	for field, limit := range map[string]int{"observations": 100, "candidates": 20, "descriptions": 10, "evidence_refs": 100, "source_refs": 100, "uncertainty_notes": 20, "warnings": 50} {
		t.Run(field, func(t *testing.T) {
			exact := []byte(`{"` + field + `":[` + strings.Repeat(`null,`, limit-1) + `null]}`)
			if _, err := parse(exact); err != nil {
				t.Fatal("exact count rejected", err)
			}
			over := []byte(`{"` + field + `":[` + strings.Repeat(`null,`, limit) + `null]}`)
			_, err := parse(over)
			var f *Failure
			if !errors.As(err, &f) || f.Category != "limit_exceeded" {
				t.Fatal("over-limit count accepted", err)
			}
		})
	}
	for _, field := range []string{"id", "selected_candidate_id", "candidate_id", "language", "evidence_refs", "source_refs"} {
		exact := []byte(`{"` + field + `":"` + strings.Repeat("x", 128) + `"}`)
		if _, err := parse(exact); err != nil {
			t.Fatal("exact identifier length rejected", err)
		}
		if _, err := parse([]byte(`{"` + field + `":"` + strings.Repeat("x", 129) + `"}`)); err == nil {
			t.Fatal("over-limit identifier accepted", field)
		}
	}
}
