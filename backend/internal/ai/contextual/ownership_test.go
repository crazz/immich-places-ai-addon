package contextual

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestFrozenBundleAndProvenanceOwnTheirData(t *testing.T) {
	in := contextInput()
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	original, _ := b.Projection()
	info := b.Info()
	if info.Binding != in.Binding || info.Version != Version || len(info.Digest) != 64 || len(info.Sources) != 2 || info.Consent.Version != Version {
		t.Fatal("missing exact provenance")
	}
	in.Consent.Classes[0] = Neighbors
	in.Hint = "changed"
	info.Consent.Classes[0] = Neighbors
	info.Sources[0].ID = "changed"
	copy, _ := b.Projection()
	clear(copy)
	again, _ := b.Projection()
	if !bytes.Equal(original, again) || b.Info().Sources[0].ID == "changed" || b.Info().Consent.Classes[0] == Neighbors {
		t.Fatal("caller widened disclosure")
	}
	if strings.Contains(fmt.Sprintf("%v %#v", b, b), "Near a bridge") || strings.Contains(fmt.Sprintf("%v %#v", b, b), "owner") {
		t.Fatal("diagnostics exposed private context")
	}
}

func TestNeighborProvenanceStaysInPrivateEnvelope(t *testing.T) {
	in := contextInput()
	in.Consent.Classes = []Class{Neighbors}
	var c Candidate
	if err := json.Unmarshal([]byte(`{"Asset":"private-neighbor","SourceDigest":"private-fingerprint","Accessible":true,"CaptureTime":"2026-09-20T06:31:00Z","Latitude":1,"Longitude":2}`), &c); err != nil {
		t.Fatal(err)
	}
	in.Candidates = []Candidate{c}
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	m := b.Info()
	if len(m.Sources) != 1 || m.Sources[0].Asset != "private-neighbor" || m.Sources[0].SourceDigest != "private-fingerprint" {
		t.Fatal("private source binding lost")
	}
	p, _ := b.Projection()
	if bytes.Contains(p, []byte("private-")) {
		t.Fatal("private binding in provider projection")
	}
}
