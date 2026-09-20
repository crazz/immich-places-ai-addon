package contextual

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNeighborLimitAndOrderingAreDeterministic(t *testing.T) {
	in := contextInput()
	in.Consent.Classes = []Class{Neighbors}
	_, start := ParseCapture(in.CaptureTime)
	for i := 9; i > 0; i-- {
		in.Candidates = append(in.Candidates, Candidate{Asset: fmt.Sprintf("private-%02d", i), Accessible: true, CaptureTime: start.Add(time.Duration(i) * time.Minute).Format(time.RFC3339), Latitude: float64(i), Longitude: 0})
	}
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := b.Projection()
	var p struct {
		Sources []struct {
			Kind     Class
			Lineage  string
			Location struct{ Latitude, Longitude float64 }
		}
	}
	if json.Unmarshal(data, &p) != nil || len(p.Sources) != 6 {
		t.Fatalf("expected six nearest metadata sources: %s", data)
	}
	for i, s := range p.Sources {
		if s.Kind != Neighbors || s.Location.Latitude != float64(i+1) || s.Lineage != "unknown" {
			t.Fatal("wrong order or invented lineage", p)
		}
	}
	if strings.Contains(string(data), "private-") {
		t.Fatal("source asset identity disclosed")
	}
}

func TestIneligibleAndAIOriginNeighborsAreExcluded(t *testing.T) {
	in := contextInput()
	in.Consent.Classes = []Class{Neighbors}
	for _, raw := range []string{
		`{"Asset":"known-ai","Accessible":true,"CaptureTime":"2026-09-20T06:31:00Z","Latitude":1,"Longitude":2,"Lineage":"ai"}`,
		`{"Asset":"allowed","Accessible":true,"CaptureTime":"2026-09-20T06:32:00Z","Latitude":0,"Longitude":0}`,
		`{"Asset":"hidden","Accessible":false,"CaptureTime":"2026-09-20T06:31:00Z","Latitude":1,"Longitude":2}`,
		`{"Asset":"target","Accessible":true,"CaptureTime":"2026-09-20T06:31:00Z","Latitude":1,"Longitude":2}`,
		`{"Asset":"offsetless","Accessible":true,"CaptureTime":"2026-09-20T06:31:00","Latitude":1,"Longitude":2}`,
		`{"Asset":"old","Accessible":true,"CaptureTime":"0001-01-02T00:00:00Z","Latitude":1,"Longitude":2}`,
		`{"Asset":"outside","Accessible":true,"CaptureTime":"2026-09-21T06:31:00Z","Latitude":1,"Longitude":2}`,
	} {
		var c Candidate
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatal(err)
		}
		in.Candidates = append(in.Candidates, c)
	}
	in.Candidates = append(in.Candidates, in.Candidates[1])
	b, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := b.Projection()
	var p Projection
	_ = json.Unmarshal(data, &p)
	if len(p.Sources) != 1 || p.Sources[0].Location.Latitude != 0 || p.Sources[0].Lineage != "unknown" {
		t.Fatalf("ineligible source disclosed: %s", data)
	}
	in.CaptureTime = "2026-09-20T08:30:00"
	b, err = Build(in)
	if err != nil {
		t.Fatal(err)
	}
	data, _ = b.Projection()
	_ = json.Unmarshal(data, &p)
	if len(p.Sources) != 0 {
		t.Fatal("offsetless target admitted temporal neighbors")
	}
}
