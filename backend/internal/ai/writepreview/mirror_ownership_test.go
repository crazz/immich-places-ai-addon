package writepreview

import (
	"encoding/json"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestMirrorReplacementRequiresVerifiedLocalOwnership(t *testing.T) {
	const id = "aaf12459-abcd-4321-8421-aaccff110033"
	d, doc, provenance := mirrorExportFixture()
	owned, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Languages: []string{"en"}}, provenance, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMirrorOwnership(MirrorBaseline{}, id, nil); err != nil {
		t.Fatal("absent namespace rejected", err)
	}
	var object map[string]any
	if json.Unmarshal(owned, &object) != nil {
		t.Fatal("invalid fixture")
	}
	reordered, _ := json.Marshal(object)
	if err := ValidateMirrorOwnership(MirrorBaseline{Present: true, Value: reordered}, id, owned); err != nil {
		t.Fatal("unchanged owned namespace rejected", err)
	}
	for _, tc := range []struct {
		current  MirrorBaseline
		verified []byte
	}{
		{MirrorBaseline{Present: true, Value: owned}, nil},
		{MirrorBaseline{Present: true, Value: []byte(`{"foreign":true}`)}, owned},
		{MirrorBaseline{Present: true, Value: []byte(strings.Replace(string(owned), "Reviewed description", "External edit", 1))}, owned},
		{MirrorBaseline{Present: true, Value: []byte(strings.Replace(string(owned), `"schemaVersion":1`, `"schemaVersion":2`, 1))}, owned},
		{MirrorBaseline{Present: true, Value: []byte(strings.Replace(string(owned), id, "baf12459-abcd-4321-8421-aaccff110033", 1))}, owned},
		{MirrorBaseline{Present: false, Value: owned}, owned},
		{MirrorBaseline{Present: true}, owned},
	} {
		if err := ValidateMirrorOwnership(tc.current, id, tc.verified); err == nil {
			t.Fatal("namespace adopted without exact verified ownership")
		}
	}
}
