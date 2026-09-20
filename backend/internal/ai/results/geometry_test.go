package results

import (
	"strings"
	"testing"
)

func camera(m map[string]any) map[string]any {
	return firstCandidate(m)["camera_location"].(map[string]any)
}

func TestRadiusClaimsNeedConsistentGranularityAndBasis(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) { camera(m)["radius_basis"] = "unknown" },
		func(m map[string]any) { camera(m)["estimated_radius_m"] = nil },
		func(m map[string]any) { camera(m)["estimated_radius_m"] = 0 },
		func(m map[string]any) { camera(m)["granularity"] = "city" },
		func(m map[string]any) { camera(m)["granularity"] = "region" },
		func(m map[string]any) {
			firstCandidate(m)["evidence_refs"] = []any{}
			firstCandidate(m)["camera_direction"] = nil
		},
		func(m map[string]any) { camera(m)["radius_basis"] = "context_extent" },
		func(m map[string]any) { camera(m)["radius_basis"] = "source_reported" },
	} {
		requireFailure(t, v, mutateFixture(t, "synthetic", change), visualContext("en", "uk"), "semantic_violation")
	}
}

func TestSupportedRadiusStatesAreRetained(t *testing.T) {
	v := testValidator(t)
	for _, tc := range []struct {
		granularity, basis string
		radius             any
		source             *Source
	}{
		{"area", "visual_estimate", 250, nil}, {"point", "visual_estimate", 0, nil},
		{"site", "unknown", nil, nil}, {"city", "unknown", nil, nil}, {"region", "unknown", nil, nil},
		{"area", "context_extent", 250, &Source{ID: "authorized", ContextExtent: true}},
		{"area", "source_reported", 250, &Source{ID: "authorized", SourceReportedRadius: true}},
	} {
		ctx := visualContext("en", "uk")
		if tc.source != nil {
			ctx.Mode = ContextAssisted
			ctx.Sources = []Source{*tc.source}
		}
		data := mutateFixture(t, "synthetic", func(m map[string]any) {
			camera(m)["granularity"] = tc.granularity
			camera(m)["radius_basis"] = tc.basis
			camera(m)["estimated_radius_m"] = tc.radius
			if tc.source != nil {
				firstCandidate(m)["source_refs"] = []any{"authorized"}
			}
		})
		proposal, err := v.Validate(data, ctx)
		if err != nil {
			t.Fatalf("supported radius %+v: %v", tc, err)
		}
		doc, err := proposal.Data()
		if err != nil {
			t.Fatal(err)
		}
		if doc.Candidates[0].CameraLocation.RadiusBasis != tc.basis || (doc.Candidates[0].CameraLocation.EstimatedRadiusM == nil) != (tc.radius == nil) {
			t.Fatal("radius repaired")
		}
		if tc.source != nil {
			ctx.Sources[0] = Source{ID: "authorized"}
			requireFailure(t, v, data, ctx, "semantic_violation")
		}
	}
}

func TestDirectionNeedsAViewpointAndMethodSpecificEvidence(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) {
			m["outcome"] = "unknown"
			m["selected_candidate_id"] = nil
			firstCandidate(m)["camera_location"] = nil
		},
		func(m map[string]any) {
			firstCandidate(m)["camera_direction"].(map[string]any)["method"] = "known_viewpoint_alignment"
		},
		func(m map[string]any) {
			camera(m)["estimated_radius_m"] = nil
			camera(m)["radius_basis"] = "unknown"
			firstCandidate(m)["evidence_refs"] = []any{}
		},
	} {
		rejected := requireFailure(t, v, mutateFixture(t, "synthetic", change), visualContext("en", "uk"), "semantic_violation")
		found := false
		for _, finding := range rejected.Findings {
			if finding.Code == "direction_evidence" {
				found = true
			}
		}
		if !found {
			t.Fatal("failure did not identify unsupported direction")
		}
	}
}

func TestHeadingBoundsAndSupportedMethodsRemainExact(t *testing.T) {
	v := testValidator(t)
	for _, azimuth := range []string{"0", "359.9999999999999999999999999"} {
		data := []byte(strings.Replace(string(fixture(t, "synthetic")), `"azimuth_deg": 40`, `"azimuth_deg": `+azimuth, 1))
		proposal, err := v.Validate(data, visualContext("en", "uk"))
		if err != nil {
			t.Fatal(err)
		}
		typed, err := proposal.Data()
		if err != nil {
			t.Fatal(err)
		}
		if typed.Candidates[0].CameraDirection.AzimuthDeg.String() != azimuth {
			t.Fatal("azimuth rounded")
		}
	}
	for _, change := range []func(map[string]any){
		func(m map[string]any) { firstCandidate(m)["camera_direction"].(map[string]any)["azimuth_deg"] = 360 },
		func(m map[string]any) { firstCandidate(m)["camera_direction"].(map[string]any)["azimuth_deg"] = -0.01 },
		func(m map[string]any) {
			firstCandidate(m)["camera_direction"].(map[string]any)["uncertainty_deg"] = 180.01
		},
		func(m map[string]any) {
			firstCandidate(m)["camera_direction"].(map[string]any)["uncertainty_deg"] = -0.01
		},
		func(m map[string]any) {
			firstCandidate(m)["camera_direction"].(map[string]any)["reference"] = "magnetic_north"
		},
	} {
		requireFailure(t, v, mutateFixture(t, "synthetic", change), visualContext("en", "uk"), "schema_violation")
	}
	for _, uncertainty := range []any{nil, 0, 180} {
		ctx := visualContext("en", "uk")
		ctx.Mode = ContextAssisted
		ctx.Sources = []Source{{ID: "alignment", ViewpointAlignment: true}}
		data := mutateFixture(t, "synthetic", func(m map[string]any) {
			firstCandidate(m)["source_refs"] = []any{"alignment"}
			direction := firstCandidate(m)["camera_direction"].(map[string]any)
			direction["method"] = "known_viewpoint_alignment"
			direction["uncertainty_deg"] = uncertainty
		})
		if _, err := v.Validate(data, ctx); err != nil {
			t.Fatal("supported alignment rejected", err)
		}
		ctx.Sources[0].ViewpointAlignment = false
		requireFailure(t, v, data, ctx, "semantic_violation")
	}
	data := mutateFixture(t, "synthetic", func(m map[string]any) { firstCandidate(m)["camera_direction"] = nil })
	if _, err := v.Validate(data, visualContext("en", "uk")); err != nil {
		t.Fatal("nullable heading rejected", err)
	}
}

func TestCoordinateBoundariesAreExactForBothPairs(t *testing.T) {
	v := testValidator(t)
	for _, pair := range [][2]float64{{-90, -180}, {90, 180}, {0, 0}} {
		data := mutateFixture(t, "synthetic", func(m map[string]any) {
			camera(m)["latitude"] = pair[0]
			camera(m)["longitude"] = pair[1]
			subject := firstCandidate(m)["subject"].(map[string]any)["location"].(map[string]any)
			subject["latitude"] = pair[0]
			subject["longitude"] = pair[1]
		})
		if _, err := v.Validate(data, visualContext("en", "uk")); err != nil {
			t.Fatal("boundary rejected", pair, err)
		}
	}
	for _, target := range []string{"camera", "subject"} {
		for _, axis := range []string{"latitude", "longitude"} {
			for _, value := range []any{nil, 180.0000001, -180.0000001} {
				data := mutateFixture(t, "synthetic", func(m map[string]any) {
					point := camera(m)
					if target == "subject" {
						point = firstCandidate(m)["subject"].(map[string]any)["location"].(map[string]any)
					}
					point[axis] = value
				})
				requireFailure(t, v, data, visualContext("en", "uk"), "schema_violation")
			}
		}
	}
}
