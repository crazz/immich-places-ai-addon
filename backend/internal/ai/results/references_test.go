package results

import "testing"

func firstObservation(m map[string]any) map[string]any {
	return m["observations"].([]any)[0].(map[string]any)
}

func TestInvalidIdentitiesAndLocalReferencesCannotBeInvented(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) { firstObservation(m)["id"] = " " },
		func(m map[string]any) { firstObservation(m)["text"] = "\n\t" },
		func(m map[string]any) { m["observations"] = append(m["observations"].([]any), firstObservation(m)) },
		func(m map[string]any) { m["candidates"] = append(m["candidates"].([]any), firstCandidate(m)) },
		func(m map[string]any) { firstCandidate(m)["place_name"] = " " },
		func(m map[string]any) { firstCandidate(m)["support_summary"] = " " },
		func(m map[string]any) { firstCandidate(m)["evidence_refs"] = []any{"missing"} },
		func(m map[string]any) { firstCandidate(m)["evidence_refs"] = []any{"obs-1", "obs-1"} },
		func(m map[string]any) { m["descriptions"].([]any)[0].(map[string]any)["candidate_id"] = "missing" },
	} {
		requireFailure(t, v, mutateFixture(t, "synthetic", change), visualContext("en", "uk"), "semantic_violation")
	}
}

func TestCountryCodesAreAssignedISOAlpha2Only(t *testing.T) {
	v := testValidator(t)
	for _, country := range []string{"UA", "PT", "US", "AQ", "AX"} {
		data := mutateFixture(t, "synthetic", func(m map[string]any) { firstCandidate(m)["country_code"] = country })
		if _, err := v.Validate(data, visualContext("en", "uk")); err != nil {
			t.Fatal(country, err)
		}
	}
	for _, country := range []string{"", "UK", "EU", "ZZ", "XK", "AC", "DG", "TA", "AA", "QZ", "USA", "001", "uk"} {
		data := mutateFixture(t, "synthetic", func(m map[string]any) { firstCandidate(m)["country_code"] = country })
		requireFailure(t, v, data, visualContext("en", "uk"), "semantic_violation")
	}
}

func TestSourcesMustBelongToThisAuthorizedBundle(t *testing.T) {
	v := testValidator(t)
	ctx := visualContext("en", "uk")
	ctx.Mode = ContextAssisted
	ctx.Sources = []Source{{ID: "allowed"}}
	for _, refs := range [][]any{{"foreign-source"}, {"https://private.invalid/secret"}, {"allowed", "allowed"}, {" "}} {
		data := mutateFixture(t, "synthetic", func(m map[string]any) { firstCandidate(m)["source_refs"] = refs })
		requireFailure(t, v, data, ctx, "semantic_violation")
	}
	good := mutateFixture(t, "synthetic", func(m map[string]any) { firstCandidate(m)["source_refs"] = []any{"allowed"} })
	if _, err := v.Validate(good, ctx); err != nil {
		t.Fatal("authorized ref rejected", err)
	}
	requireFailure(t, v, good, visualContext("en", "uk"), "semantic_violation")
}

func TestProvidedContextRetainsAnAuthorizedSourceConnection(t *testing.T) {
	v := testValidator(t)
	data := mutateFixture(t, "synthetic", func(m map[string]any) {
		m["observations"] = append(m["observations"].([]any), map[string]any{"id": "context-observation", "kind": "provided_context", "text": "Synthetic authorized context"})
		firstCandidate(m)["evidence_refs"] = []any{"obs-1", "context-observation"}
	})
	requireFailure(t, v, data, visualContext("en", "uk"), "semantic_violation")
	ctx := visualContext("en", "uk")
	ctx.Mode = ContextAssisted
	requireFailure(t, v, data, ctx, "semantic_violation")
	ctx.Sources = []Source{{ID: "authorized"}}
	requireFailure(t, v, data, ctx, "semantic_violation")
	good := mutateFixture(t, "synthetic", func(m map[string]any) {
		m["observations"] = append(m["observations"].([]any), map[string]any{"id": "context-observation", "kind": "provided_context", "text": "Synthetic authorized context"})
		firstCandidate(m)["evidence_refs"] = []any{"obs-1", "context-observation"}
		firstCandidate(m)["source_refs"] = []any{"authorized"}
	})
	if _, err := v.Validate(good, ctx); err != nil {
		t.Fatal("valid context connection rejected", err)
	}
}
