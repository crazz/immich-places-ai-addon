package writepreview

import (
	"strings"
	"testing"
)

func TestMirrorNamespaceRequiresCompleteUnambiguousObjectList(t *testing.T) {
	for _, raw := range []string{`[]`, `[{"key":"other","value":{"private":"unrelated"}}]`} {
		got, err := ParseMirrorNamespace([]byte(raw))
		if err != nil || got.Present || len(got.Value) != 0 {
			t.Fatalf("valid complete absence rejected: %+v %v", got, err)
		}
	}
	got, err := ParseMirrorNamespace([]byte(`[{"key":"immich-places-ai-addon","value":{"a":1,"b":[true,null,"text"]},"updatedAt":"2026-09-26T00:00:00Z"},{"key":"other","value":{}}]`))
	if err != nil || !got.Present || string(got.Value) != `{"a":1,"b":[true,null,"text"]}` {
		t.Fatalf("namespace not extracted exactly: %+v %v", got, err)
	}
	for _, invalid := range []string{
		`null`, `{}`, `[] {}`, `[`, `[{}]`, `[{"key":"","value":{}}]`,
		`[{"key":"x","value":null}]`, `[{"key":"x","value":[]}]`,
		`[{"key":"x","key":"y","value":{}}]`, `[{"key":"x","value":{"a":1,"a":2}}]`,
		`[{"key":"x","value":{}},{"key":"x","value":{}}]`,
		`[{"key":"x","value":{"n":1e999}}]`, `[{"key":"x","value":{"n":NaN}}]`,
		`[{"key":"x","value":{"text":"\ud800"}}]`,
		`[{"key":"x","value":{"text":"` + string([]byte{0xff}) + `"}}]`,
		strings.Repeat(" ", (1<<20)+1) + `[]`,
	} {
		if _, err := ParseMirrorNamespace([]byte(invalid)); err == nil {
			t.Fatal("malformed/ambiguous metadata treated as available")
		}
	}
}

func TestMirrorSemanticComparisonPreservesArraysTextAndNumericPrecision(t *testing.T) {
	for _, tc := range []struct {
		left, right string
		equal       bool
	}{
		{`{"a":1,"b":["x",true,null]}`, `{"b":["x",true,null],"a":1.0}`, true},
		{`{"n":0}`, `{"n":-0.0}`, true},
		{`{"n":0.01}`, `{"n":1e-2}`, true},
		{`{"n":9007199254740992}`, `{"n":9007199254740993}`, false},
		{`{"a":[1,2]}`, `{"a":[2,1]}`, false},
		{`{"text":"é"}`, `{"text":"é"}`, false},
		{`{"a":1}`, `{"a":"1"}`, false},
		{`{"a":null}`, `{}`, false},
		{`{"a":1,"a":1}`, `{"a":1}`, false},
		{`null`, `null`, false},
		{`[]`, `[]`, false},
	} {
		if EqualMirrorValue([]byte(tc.left), []byte(tc.right)) != tc.equal {
			t.Fatalf("incorrect semantic equality for %s / %s", tc.left, tc.right)
		}
	}
}

func TestMirrorNamespaceRejectsUnboundedNumericWork(t *testing.T) {
	for _, number := range []string{"1e-1000000000", "0e-1000000000", "0." + strings.Repeat("1", 129)} {
		if _, err := ParseMirrorNamespace([]byte(`[{"key":"other","value":{"n":` + number + `}}]`)); err == nil {
			t.Fatal("unbounded numeric normalization accepted")
		}
	}
}
