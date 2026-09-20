package contextual

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestBundleRoundTripRetainsExactPrivateProvenance(t *testing.T) {
	binding := Binding{Owner: "owner", Installation: "installation", Asset: "asset", Selection: "selection", Profile: "profile", SourceDigest: "digest", Revision: 1}
	bundle, err := Build(Input{Binding: binding, Consent: Consent{Binding: binding, Version: Version, Classes: []Class{Hint}}, Hint: "private hint", Window: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var restored Bundle
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	original, _ := bundle.Projection()
	actual, err := restored.Projection()
	if err != nil || !bytes.Equal(original, actual) || restored.Info().Digest != bundle.Info().Digest {
		t.Fatal("durable bundle lost projection or authority", err)
	}
}
