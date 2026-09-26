package writepreview

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestSelectExactStackSubset(t *testing.T) {
	analyzed := "00000000-0000-4000-8000-000000000003"
	a := "00000000-0000-4000-8000-000000000001"
	b := "00000000-0000-4000-8000-000000000002"
	other := "00000000-0000-4000-8000-000000000004"
	snapshot := Snapshot{AssetID: analyzed, Fields: []string{"gps"}, Camera: &Point{Latitude: 0, Longitude: 12}}
	requested := []string{analyzed, b, a, b}
	got, err := SelectStackTargets(snapshot, requested, []string{analyzed, a, b, other})
	if err != nil || !reflect.DeepEqual(got, []string{a, b, analyzed}) {
		t.Fatalf("exact sorted deduplicated selection = %v, %v", got, err)
	}
	if !reflect.DeepEqual(requested, []string{analyzed, b, a, b}) {
		t.Fatal("selection mutated caller input")
	}
}

func TestStackSelectionDefaultsToAnalyzedPhoto(t *testing.T) {
	snapshot := Snapshot{AssetID: "00000000-0000-4000-8000-000000000003", Fields: []string{"description"}}
	got, err := SelectStackTargets(snapshot, nil, []string{snapshot.AssetID, "00000000-0000-4000-8000-000000000004"})
	if err != nil || !reflect.DeepEqual(got, []string{snapshot.AssetID}) {
		t.Fatalf("default expanded or omitted analyzed photo: %v, %v", got, err)
	}
}

func TestStackSelectionRejectsUnauthorizedOrUnreadyExpansion(t *testing.T) {
	root := "00000000-0000-4000-8000-000000000001"
	sibling := "00000000-0000-4000-8000-000000000002"
	base := Snapshot{AssetID: root, Fields: []string{"gps"}, Camera: &Point{Latitude: 0, Longitude: 12}}
	many := []string{root}
	for i := 2; i <= 51; i++ {
		many = append(many, fmt.Sprintf("00000000-0000-4000-8000-%012d", i))
	}
	for _, tc := range []struct {
		name         string
		fields       []string
		camera       *Point
		ids, members []string
	}{
		{"foreign", base.Fields, base.Camera, []string{root, sibling}, []string{root}},
		{"missing analyzed", base.Fields, base.Camera, []string{sibling}, []string{root, sibling}},
		{"malformed ID", base.Fields, base.Camera, []string{root, "bad"}, []string{root, "bad"}},
		{"too many", base.Fields, base.Camera, many, many},
		{"description only", []string{"description"}, base.Camera, []string{root, sibling}, []string{root, sibling}},
		{"no camera", base.Fields, nil, []string{root, sibling}, []string{root, sibling}},
		{"invalid camera", base.Fields, &Point{Latitude: math.NaN()}, []string{root, sibling}, []string{root, sibling}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := base
			snapshot.Fields, snapshot.Camera = tc.fields, tc.camera
			got, err := SelectStackTargets(snapshot, tc.ids, tc.members)
			if err == nil || len(got) != 0 {
				t.Fatalf("unusable scope accepted: %v %v", got, err)
			}
		})
	}
}
