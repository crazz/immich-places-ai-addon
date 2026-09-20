package results

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

func maximumWorkload(t *testing.T) ([]byte, Context) {
	t.Helper()
	var doc Document
	if err := json.Unmarshal(fixture(t, "synthetic"), &doc); err != nil {
		t.Fatal(err)
	}
	ctx := visualContext("en", "uk", "pt", "fr", "de", "it", "es", "ja", "ko", "zh")
	ctx.Mode = ContextAssisted
	evidence := []string{}
	sources := []string{}
	doc.Observations = nil
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("%03d-", i) + strings.Repeat("x", 124)
		evidence = append(evidence, id)
		doc.Observations = append(doc.Observations, Observation{ID: id, Kind: "visual", Text: strings.Repeat("v", 1024)})
		source := fmt.Sprintf("source-%03d", i)
		sources = append(sources, source)
		ctx.Sources = append(ctx.Sources, Source{ID: source, ContextExtent: true, SourceReportedRadius: true, ViewpointAlignment: true})
	}
	candidate := doc.Candidates[0]
	candidate.CameraLocation.Latitude = json.Number("0." + strings.Repeat("0", 125) + "1")
	doc.Candidates = make([]Candidate, 20)
	for i := range doc.Candidates {
		next := candidate
		next.ID = fmt.Sprintf("candidate-%03d", i)
		next.EvidenceRefs = evidence
		next.SourceRefs = sources
		next.UncertaintyNotes = make([]string, 20)
		for j := range next.UncertaintyNotes {
			next.UncertaintyNotes[j] = strings.Repeat("n", 64)
		}
		doc.Candidates[i] = next
	}
	doc.SelectedCandidateID = &doc.Candidates[0].ID
	doc.Descriptions = nil
	for _, tag := range ctx.Languages {
		text := strings.Repeat("d", 8192)
		doc.Descriptions = append(doc.Descriptions, Description{Language: tag, Status: "complete", Text: &text, Basis: "candidate", CandidateID: doc.SelectedCandidateID})
	}
	doc.Warnings = make([]string, 50)
	for i := range doc.Warnings {
		doc.Warnings[i] = strings.Repeat("w", 8192)
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > maxBytes {
		t.Fatalf("invalid maximum fixture: %d bytes", len(data))
	}
	data = append(data, []byte(strings.Repeat(" ", maxBytes-len(data)))...)
	return data, ctx
}

func TestRepresentativeAndMaximumBoundValidation(t *testing.T) {
	validator := testValidator(t)
	maximum, ctx := maximumWorkload(t)
	for _, tc := range []struct {
		name       string
		data       []byte
		ctx        Context
		candidates int
	}{
		{"representative", fixture(t, "synthetic"), visualContext("en", "uk"), 1},
		{"maximum", maximum, ctx, 20},
	} {
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		start := time.Now()
		proposal, err := validator.Validate(tc.data, tc.ctx)
		elapsed := time.Since(start)
		runtime.ReadMemStats(&after)
		if err != nil {
			t.Fatal(tc.name, err)
		}
		doc, err := proposal.Data()
		if err != nil || len(doc.Candidates) != tc.candidates {
			t.Fatal("maximum result truncated", err)
		}
		t.Logf("%s: bytes=%d elapsed=%s allocated=%d; local synthetic, not hardware acceptance", tc.name, len(tc.data), elapsed, after.TotalAlloc-before.TotalAlloc)
	}
}
