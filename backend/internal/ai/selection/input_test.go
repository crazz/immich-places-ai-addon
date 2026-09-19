package selection

import (
	"errors"
	"reflect"
	"testing"
)

func TestExplicitDuplicatesRetainFirstOccurrence(t *testing.T) {
	ids := []string{"aaaaaaaa-0000-4000-8000-000000000001", "bbbbbbbb-0000-4000-8000-000000000002"}
	input := Input{Mode: "explicit", AssetIDs: []string{ids[0], ids[1], "AAAAAAAA-0000-4000-8000-000000000001"}, Scope: &Scope{View: "all"}}
	got, err := Normalize(input, 500)
	if err != nil || !reflect.DeepEqual(got.AssetIDs, ids) || got.RequestedCount != 3 || got.DuplicateCount != 1 {
		t.Fatalf("normalization = %+v, error = %v", got, err)
	}
	if got.Scope.GPSFilter != "no-gps" || got.Scope.HiddenFilter != "visible" {
		t.Fatalf("effective defaults missing: %+v", got.Scope)
	}
}

func TestInvalidScopeCannotBroadenSelection(t *testing.T) {
	for _, scope := range []Scope{
		{View: "search"}, {View: ""}, {View: "album"}, {View: "all", AlbumID: "album"},
		{View: "folder", FolderPath: "/"}, {View: "folder", FolderPath: "/Trip", AlbumID: "album"},
		{View: "all", GPSFilter: "missing"}, {View: "all", HiddenFilter: "yes"},
		{View: "all", StartDate: "2026-02-30"}, {View: "all", StartDate: "2026-09-20", EndDate: "2026-09-19"},
	} {
		input := Input{Mode: "explicit", AssetIDs: []string{"aaaaaaaa-0000-4000-8000-000000000001"}, Scope: &scope}
		if _, err := Normalize(input, 500); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid scope accepted: %+v (%v)", scope, err)
		}
	}
}

func TestExplicitLimitsRejectWholeRequest(t *testing.T) {
	id := "aaaaaaaa-0000-4000-8000-000000000001"
	input := Input{Mode: "explicit", AssetIDs: []string{id, "bbbbbbbb-0000-4000-8000-000000000002"}, Scope: &Scope{View: "all"}}
	if result, err := Normalize(input, 1); !errors.Is(err, ErrLimit) || len(result.AssetIDs) != 0 {
		t.Fatalf("oversized batch returned partial success: %+v, %v", result, err)
	}
	input.AssetIDs = make([]string, 10001)
	for i := range input.AssetIDs {
		input.AssetIDs[i] = id
	}
	if _, err := Normalize(input, 500); !errors.Is(err, ErrLimit) {
		t.Fatalf("raw IDs bypassed cap: %v", err)
	}
	input.AssetIDs = []string{id, id}
	if got, err := Normalize(input, 1); err != nil || len(got.AssetIDs) != 1 {
		t.Fatalf("exact unique cap: %+v %v", got, err)
	}
}

func TestExplicitRejectsInvalidInput(t *testing.T) {
	id := "aaaaaaaa-0000-4000-8000-000000000001"
	for _, input := range []Input{
		{Mode: "all-matching", AssetIDs: []string{id}, Scope: &Scope{View: "all"}},
		{Mode: "explicit", Scope: &Scope{View: "all"}},
		{Mode: "explicit", AssetIDs: []string{id}},
		{Mode: "explicit", AssetIDs: []string{"invalid"}, Scope: &Scope{View: "all"}},
	} {
		if _, err := Normalize(input, 500); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid input accepted: %+v, error = %v", input, err)
		}
	}
}
