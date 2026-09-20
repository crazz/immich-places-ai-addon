package contextual

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func neighborSources(in Input) []Source {
	_, target := ParseCapture(in.CaptureTime)
	if target.IsZero() {
		return nil
	}
	type nearby struct {
		candidate Candidate
		delta     time.Duration
	}
	eligible := []nearby{}
	seen := map[string]bool{in.Binding.Asset: true}
	for _, c := range in.Candidates {
		_, captured := ParseCapture(c.CaptureTime)
		if seen[c.Asset] || c.Asset == "" || !c.Accessible || c.Lineage == "ai" || captured.IsZero() || !validCoordinates(c.Latitude, c.Longitude) {
			continue
		}
		seen[c.Asset] = true
		delta := captured.Sub(target)
		if captured.Before(target) {
			delta = target.Sub(captured)
		}
		if delta > in.Window {
			continue
		}
		eligible = append(eligible, nearby{c, delta})
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].delta == eligible[j].delta {
			return eligible[i].candidate.Asset < eligible[j].candidate.Asset
		}
		return eligible[i].delta < eligible[j].delta
	})
	if len(eligible) > 6 {
		eligible = eligible[:6]
	}
	sources := []Source{}
	for i, n := range eligible {
		stamp, _ := ParseCapture(n.candidate.CaptureTime)
		sources = append(sources, Source{ID: fmt.Sprintf("neighbor-%d", i+1), Kind: Neighbors, Time: &stamp, Location: &Coordinate{n.candidate.Latitude, n.candidate.Longitude}, Lineage: "unknown", Asset: n.candidate.Asset, SourceDigest: n.candidate.SourceDigest})
	}
	return sources
}

func validCoordinates(lat, lon float64) bool {
	return !math.IsNaN(lat) && !math.IsInf(lat, 0) && !math.IsNaN(lon) && !math.IsInf(lon, 0) && lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}
