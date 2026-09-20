package results

import "testing"

func firstCandidate(m map[string]any) map[string]any {
	return m["candidates"].([]any)[0].(map[string]any)
}

func TestContradictoryOutcomesAreRejectedWithoutRepair(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) { m["selected_candidate_id"] = nil },
		func(m map[string]any) { m["selected_candidate_id"] = "absent" },
		func(m map[string]any) {
			firstCandidate(m)["camera_location"] = nil
			firstCandidate(m)["camera_direction"] = nil
		},
		func(m map[string]any) { m["outcome"] = "ambiguous"; m["selected_candidate_id"] = nil },
		func(m map[string]any) { m["outcome"] = "ambiguous" },
		func(m map[string]any) { m["outcome"] = "unknown" },
	} {
		data := mutateFixture(t, "synthetic", change)
		requireFailure(t, v, data, visualContext("en", "uk"), "semantic_violation")
	}
}

func TestTypedCameraAndSubjectCoordinatesStaySeparate(t *testing.T) {
	v := testValidator(t)
	data := mutateFixture(t, "synthetic", func(m map[string]any) {
		firstCandidate(m)["camera_location"].(map[string]any)["latitude"] = 0
		firstCandidate(m)["camera_location"].(map[string]any)["longitude"] = 0
	})
	proposal, err := v.Validate(data, visualContext("en", "uk"))
	if err != nil {
		t.Fatal(err)
	}
	typed, err := proposal.Data()
	if err != nil {
		t.Fatal(err)
	}
	if typed.Candidates[0].CameraLocation.Latitude.String() != "0" || typed.Candidates[0].CameraLocation.Longitude.String() != "0" || typed.Candidates[0].Subject.Location.Latitude.String() != "50.001" {
		t.Fatalf("coordinate pairs conflated: %+v", typed.Candidates[0])
	}
}

func TestUnknownSubjectAndAmbiguousAlternativesRemainUnselected(t *testing.T) {
	v := testValidator(t)
	for _, outcome := range []string{"unknown", "ambiguous"} {
		data := mutateFixture(t, "synthetic", func(m map[string]any) {
			m["outcome"] = outcome
			m["selected_candidate_id"] = nil
			if outcome == "unknown" {
				firstCandidate(m)["camera_location"] = nil
				firstCandidate(m)["camera_direction"] = nil
			} else {
				copy := map[string]any{}
				for k, value := range firstCandidate(m) {
					copy[k] = value
				}
				copy["id"] = "alternative"
				m["candidates"] = append(m["candidates"].([]any), copy)
			}
		})
		proposal, err := v.Validate(data, visualContext("en", "uk"))
		if err != nil {
			t.Fatal(err)
		}
		typed, err := proposal.Data()
		if err != nil {
			t.Fatal(err)
		}
		if typed.Outcome != outcome || typed.SelectedCandidateID != nil {
			t.Fatalf("outcome changed: %+v", typed)
		}
		if outcome == "unknown" && (typed.Candidates[0].CameraLocation != nil || typed.Candidates[0].Subject.Location == nil) {
			t.Fatal("subject promoted to camera")
		}
		if outcome == "ambiguous" && (len(typed.Candidates) != 2 || typed.Candidates[1].ID != "alternative") {
			t.Fatal("alternative/order lost")
		}
	}
}
