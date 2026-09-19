package selection

import "testing"

func TestNormalizeAllMatchingScopeWithoutExplicitIDs(t *testing.T) {
	request, err := Normalize(Input{Mode: "all-matching", Scope: &Scope{View: "folder", FolderPath: "/Trip/"}}, 500)
	if err != nil || request.Mode != "all-matching" || request.Scope.FolderPath != "/Trip" || request.Scope.GPSFilter != "no-gps" || request.Scope.HiddenFilter != "visible" || request.RequestedCount != 0 || len(request.AssetIDs) != 0 {
		t.Fatalf("all-matching input: %+v %v", request, err)
	}
}

func TestDecodeRejectsAllMatchingAssetIDPresence(t *testing.T) {
	for _, ids := range []string{`null`, `[]`, `["aaaaaaaa-0000-4000-8000-000000000001"]`} {
		if _, err := Decode([]byte(`{"mode":"all-matching","assetIDs":` + ids + `,"scope":{"view":"all"}}`)); err == nil {
			t.Fatalf("accepted mixed mode: %s", ids)
		}
	}
}
