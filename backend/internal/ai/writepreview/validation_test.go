package writepreview

import (
	"errors"
	"math"
	"testing"
)

func TestReadinessRequiresOnlyExactStagedGPSAndAcknowledgedBaseline(t *testing.T) {
	valid := Snapshot{Revision: 1, State: "staged", Camera: &Point{Latitude: 0, Longitude: 0}, Fields: []string{"gps"}, BaselineReviewed: true, ImageIdentity: "v1:image"}
	for _, tc := range []struct {
		name, code string
		change     func(*Snapshot)
	}{
		{"valid", "", func(*Snapshot) {}},
		{"old revision", "DRAFT_CONFLICT", func(s *Snapshot) { s.Revision = 2 }},
		{"draft", "DRAFT_NOT_STAGED", func(s *Snapshot) { s.State = "draft" }},
		{"rejected", "DRAFT_NOT_STAGED", func(s *Snapshot) { s.State = "rejected" }},
		{"absent camera", "INVALID_PREVIEW", func(s *Snapshot) { s.Camera = nil }},
		{"latitude", "INVALID_PREVIEW", func(s *Snapshot) { s.Camera = &Point{Latitude: 91} }},
		{"longitude", "INVALID_PREVIEW", func(s *Snapshot) { s.Camera = &Point{Longitude: 181} }},
		{"nan", "INVALID_PREVIEW", func(s *Snapshot) { s.Camera = &Point{Latitude: math.NaN()} }},
		{"infinite", "INVALID_PREVIEW", func(s *Snapshot) { s.Camera = &Point{Longitude: math.Inf(1)} }},
		{"unselected", "INVALID_PREVIEW", func(s *Snapshot) { s.Fields = nil }},
		{"expanded", "INVALID_PREVIEW", func(s *Snapshot) { s.Fields = []string{"gps", "description"} }},
		{"substituted", "INVALID_PREVIEW", func(s *Snapshot) { s.Fields = []string{"heading"} }},
		{"unreviewed", "BASELINE_REVIEW_REQUIRED", func(s *Snapshot) { s.BaselineReviewed = false }},
		{"no identity", "BASELINE_REVIEW_REQUIRED", func(s *Snapshot) { s.ImageIdentity = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := valid
			tc.change(&snapshot)
			err := Validate(snapshot, 1)
			var failure Failure
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !errors.As(err, &failure) || failure.Code != tc.code {
				t.Fatal("unexpected readiness", err)
			}
		})
	}
}
