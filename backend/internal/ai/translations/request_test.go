package translations

import (
	"reflect"
	"strings"
	"testing"
)

func approvedRequest() Request {
	return Request{Key: "submission", DraftID: "draft", Revision: 2, FactsRevision: 2,
		ProfileID: "profile", ProfileRevision: 1, Basis: "An uncertain stone bridge.\nNo identified place.",
		BasisKind: "scene", Languages: []string{"uk", "en"}, Confirmed: true}
}

func TestNormalizeRejectsUnapprovedOrUnboundedInputs(t *testing.T) {
	for name, alter := range map[string]func(*Request){
		"consent":           func(r *Request) { r.Confirmed = false },
		"empty":             func(r *Request) { r.Basis = "  " },
		"oversize":          func(r *Request) { r.Basis = strings.Repeat("x", (16<<10)+1) },
		"invalid utf8":      func(r *Request) { r.Basis = string([]byte{255}) },
		"basis kind":        func(r *Request) { r.BasisKind = "hidden" },
		"languages empty":   func(r *Request) { r.Languages = nil },
		"languages large":   func(r *Request) { r.Languages = []string{"en", "uk", "fr", "de", "es", "it", "pt", "nl", "pl"} },
		"duplicate":         func(r *Request) { r.Languages = []string{"en", "EN"} },
		"invalid tag":       func(r *Request) { r.Languages = []string{"not_a_language"} },
		"empty tag":         func(r *Request) { r.Languages = []string{""} },
		"revision":          func(r *Request) { r.Revision = 0 },
		"facts":             func(r *Request) { r.FactsRevision = 0 },
		"provider revision": func(r *Request) { r.ProfileRevision = 0 },
		"identity":          func(r *Request) { r.Key = "" },
		"draft":             func(r *Request) { r.DraftID = "" },
		"profile":           func(r *Request) { r.ProfileID = "" },
	} {
		t.Run(name, func(t *testing.T) {
			r := approvedRequest()
			alter(&r)
			if _, _, err := Normalize(r); err != ErrInvalid {
				t.Fatalf("invalid input accepted: %v", err)
			}
		})
	}
}

func TestNormalizeFreezesOnlyReviewedTextAndLanguages(t *testing.T) {
	req := approvedRequest()
	normal, digest, err := Normalize(req)
	if err != nil || normal.Basis != req.Basis || normal.BasisKind != "scene" || len(digest) != 64 || !reflect.DeepEqual(normal.Languages, []string{"en", "uk"}) {
		t.Fatalf("reviewed text request not normalized: %#v %q %v", normal, digest, err)
	}
	if !reflect.DeepEqual(req.Languages, []string{"uk", "en"}) {
		t.Fatal("mutated caller languages")
	}
	_, again, err := Normalize(normal)
	if err != nil || again != digest {
		t.Fatal("unstable request identity")
	}
}

func TestNormalizeRejectsInvalidBudgetsAndRetryIdentity(t *testing.T) {
	for _, alter := range []func(*Request){
		func(r *Request) { r.MaxTokens = -1 }, func(r *Request) { r.MaxTokens = 8000000001 },
		func(r *Request) { v := int64(-1); r.MaxEstimatedMicros = &v },
		func(r *Request) { r.ParentID = strings.Repeat("x", 129) },
	} {
		r := approvedRequest()
		alter(&r)
		if _, _, err := Normalize(r); err != ErrInvalid {
			t.Fatal("invalid budget or retry identity accepted", err)
		}
	}
}
