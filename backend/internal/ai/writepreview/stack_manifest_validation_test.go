package writepreview

import (
	"strings"
	"testing"
	"time"
)

func TestStackManifestRejectsStaleIncompleteOrChangedReview(t *testing.T) {
	primary, sibling := "00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"
	for _, tc := range []struct {
		name   string
		mutate func(*Snapshot, *StackReview, map[string]Metadata)
	}{
		{"owner", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.Owner = "other" }},
		{"installation", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.Installation = "other" }},
		{"revision", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.DraftRevision++ }},
		{"draft", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.DraftID = "other" }},
		{"analyzed source", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.ImageIdentity = "other" }},
		{"stack identity", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.StackID = "" }},
		{"review identity", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.ID = "" }},
		{"expired", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			r.ExpiresAt = time.Unix(1000, 0).Format(time.RFC3339Nano)
		}},
		{"missing member read", func(s *Snapshot, r *StackReview, m map[string]Metadata) { delete(m, sibling) }},
		{"changed member image", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			v := m[sibling]
			v.ImageIdentity = "replacement"
			m[sibling] = v
		}},
		{"changed member GPS", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			v := m[sibling]
			z := 0.0
			v.GPS.Latitude = &z
			m[sibling] = v
		}},
		{"unreadable GPS", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			v := m[sibling]
			v.GPSUnavailable = true
			m[sibling] = v
		}},
		{"description only", func(s *Snapshot, r *StackReview, m map[string]Metadata) { s.Fields = []string{"description"} }},
		{"no camera", func(s *Snapshot, r *StackReview, m map[string]Metadata) { s.Camera = nil }},
		{"single target", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.Targets = r.Targets[:1] }},
		{"duplicate", func(s *Snapshot, r *StackReview, m map[string]Metadata) { r.Targets[1] = r.Targets[0] }},
		{"order", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			r.Targets[0], r.Targets[1] = r.Targets[1], r.Targets[0]
		}},
		{"oversized", func(s *Snapshot, r *StackReview, m map[string]Metadata) {
			r.Targets[1].ImageIdentity = strings.Repeat("a", 1<<20)
			v := m[sibling]
			v.ImageIdentity = r.Targets[1].ImageIdentity
			m[sibling] = v
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Unix(1000, 0)
			snapshot := Snapshot{Owner: "owner", Installation: "installation", ID: "draft", Revision: 2, AssetID: primary, AnalysisID: "analysis", State: "staged", Fields: []string{"gps"}, Camera: &Point{Latitude: 0, Longitude: 12}, BaselineReviewed: true, ImageIdentity: "image1", PolicyID: strings.Repeat("a", 64)}
			review := StackReview{ID: "00000000-0000-4000-8000-000000000004", Owner: snapshot.Owner, Installation: snapshot.Installation, DraftID: snapshot.ID, DraftRevision: 2, AnalyzedID: primary, ImageIdentity: "image1", StackID: "00000000-0000-4000-8000-000000000003", ExpiresAt: now.Add(time.Minute).Format(time.RFC3339Nano), Targets: []ReviewedTarget{{AssetID: primary, ImageIdentity: "image1"}, {AssetID: sibling, ImageIdentity: "image2"}}}
			current := map[string]Metadata{primary: {ImageIdentity: "image1"}, sibling: {ImageIdentity: "image2"}}
			tc.mutate(&snapshot, &review, current)
			preview, raw, err := BuildStack(snapshot, review, current, "preview", now)
			if err == nil || len(raw) != 0 || preview.Status == "usable" {
				t.Fatalf("invalid review produced usable authority: %v", err)
			}
		})
	}
}
