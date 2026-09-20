package contextual

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func contextInput() Input {
	b := Binding{Owner: "owner", Installation: "installation", Asset: "target", Selection: "selection", Profile: "profile", Revision: 1, SourceDigest: strings.Repeat("a", 64)}
	return Input{Binding: b, Consent: Consent{Binding: b, Version: Version, Classes: []Class{Capture, Hint}}, CaptureTime: "2026-09-20T08:30:00+02:00", Hint: "Near a bridge", Window: 6 * time.Hour}
}

func TestOnlyConsentedClassesAreDisclosed(t *testing.T) {
	in := contextInput()
	in.Album = &Album{ID: "private-album", Label: "Private trip", Member: true}
	in.Candidates = []Candidate{{Asset: "private-neighbor", Accessible: true, CaptureTime: in.CaptureTime, Latitude: 1, Longitude: 2}}
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, err := b.Projection()
	if err != nil {
		t.Fatal(err)
	}
	var p Projection
	if json.Unmarshal(data, &p) != nil || len(p.Sources) != 2 {
		t.Fatalf("unexpected projection: %s", data)
	}
	if p.Sources[0].Kind != Capture || p.Sources[1].Kind != Hint || p.Sources[1].Text != in.Hint {
		t.Fatal("consented content missing")
	}
	for _, private := range []string{"Private trip", "private-album", "private-neighbor", "owner", "installation", "profile", "selection"} {
		if strings.Contains(string(data), private) {
			t.Fatalf("unconsented disclosure: %s", private)
		}
	}
}

func TestMissingOrMismatchedConsentPreventsPreparation(t *testing.T) {
	for _, change := range []func(*Input){
		func(in *Input) { in.Consent.Version = "" },
		func(in *Input) { in.Consent.Binding.Owner = "foreign" },
		func(in *Input) { in.Consent.Binding.Installation = "other" },
		func(in *Input) { in.Consent.Binding.Asset = "other" },
		func(in *Input) { in.Consent.Binding.Selection = "other" },
		func(in *Input) { in.Consent.Binding.Profile = "other" },
		func(in *Input) { in.Consent.Binding.Revision++ },
		func(in *Input) { in.Consent.Classes = []Class{"private_paths"} },
		func(in *Input) { in.Consent.Classes = []Class{Hint, Hint} },
		func(in *Input) { in.Binding.Owner = ""; in.Consent.Binding = in.Binding },
	} {
		in := contextInput()
		change(&in)
		if b, err := Build(in); err == nil || b != nil {
			t.Fatal("invalid consent admitted")
		}
	}
}

func TestCaptureTimePreservesOffsetAndMissingSemantics(t *testing.T) {
	for _, tc := range []struct{ raw, status, day string }{
		{"2026-01-01T00:30:00+14:00", "offset", "2026-01-01"},
		{"2026-01-01T00:30:00", "local_unknown", "2026-01-01"},
		{"2026-01-01", "local_unknown", "2026-01-01"},
		{"", "missing", ""}, {"2026-02-30T12:00:00Z", "invalid", ""},
	} {
		in := contextInput()
		in.CaptureTime = tc.raw
		in.Consent.Classes = []Class{Capture}
		b, err := Build(in)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := b.Projection()
		var p struct {
			Sources []struct {
				Time struct{ Status, Day, Value string }
			}
		}
		if json.Unmarshal(data, &p) != nil || len(p.Sources) != 1 || p.Sources[0].Time.Status != tc.status || p.Sources[0].Time.Day != tc.day {
			t.Fatalf("wrong capture semantics: %s", data)
		}
		if tc.day != "" && p.Sources[0].Time.Value != tc.raw {
			t.Fatal("capture timestamp changed")
		}
	}
}

func TestInvalidContextBoundsFailWithoutPartialBundle(t *testing.T) {
	for _, change := range []func(*Input){
		func(in *Input) { in.Hint = strings.Repeat("a", 2001) },
		func(in *Input) { in.Hint = string([]byte{0xff}) },
		func(in *Input) { in.Window = 0 },
		func(in *Input) { in.Window = 25 * time.Hour },
		func(in *Input) { in.Candidates = make([]Candidate, 65) },
		func(in *Input) {
			in.Consent.Classes = []Class{AlbumLabel}
			in.Consent.AlbumID = "album"
			in.Album = &Album{ID: "album", Member: true, Label: strings.Repeat("x", 257)}
		},
	} {
		in := contextInput()
		change(&in)
		if b, err := Build(in); err == nil || b != nil {
			t.Fatal("invalid bounds admitted")
		}
	}
}

func TestSelectedAlbumCannotBroadenScope(t *testing.T) {
	in := contextInput()
	in.Consent.Classes = []Class{AlbumLabel}
	in.Consent.AlbumID = "album"
	in.Album = &Album{ID: "album", Label: "Selected trip", Member: true}
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := b.Projection()
	var p Projection
	_ = json.Unmarshal(data, &p)
	if len(p.Sources) != 1 || p.Sources[0].Kind != AlbumLabel || p.Sources[0].Text != "Selected trip" {
		t.Fatal("selected album missing")
	}
	in.Album.ID = "foreign"
	if b, err = Build(in); err == nil || b != nil {
		t.Fatal("foreign album admitted")
	}
	in.Album.ID = "album"
	in.Album.Member = false
	b, err = Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = b.Projection()
	_ = json.Unmarshal(data, &p)
	if len(p.Sources) != 0 {
		t.Fatal("former membership disclosed album")
	}
}

func TestEmptyAuthorizedContextRemainsExplicit(t *testing.T) {
	in := contextInput()
	in.Consent.Classes = []Class{Neighbors}
	in.CaptureTime = ""
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := b.Projection()
	var p struct {
		Status  string
		Sources []Source
	}
	if json.Unmarshal(data, &p) != nil || p.Status != "empty" || len(p.Sources) != 0 {
		t.Fatalf("empty context not explicit: %s", data)
	}
}
