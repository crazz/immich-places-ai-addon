package writepreview

import (
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/language"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
)

func validateMirrorExport(d drafts.Draft, document results.Document, choice drafts.MirrorSelection, provenance MirrorProvenance, recordID string) error {
	invalid := Failure{Code: "INVALID_MIRROR"}
	id, err := uuid.Parse(recordID)
	if err != nil || id.Version() != 4 || id.String() != recordID || d.Revision < 1 || d.FactsRevision < 1 {
		return invalid
	}
	if (!choice.Direction && !choice.Precision && !choice.Place && !choice.Provenance && len(choice.Languages) == 0) || (choice.Model && !choice.Provenance) || len(choice.Languages) > 8 {
		return invalid
	}
	if choice.Direction && (d.HeadingStale || d.HeadingFactsRevision != d.FactsRevision || (d.Heading != nil && (!finite(*d.Heading, 360) || *d.Heading < 0 || *d.Heading >= 360))) {
		return invalid
	}
	if choice.Precision && (d.RadiusStale || (d.Radius != nil && (math.IsNaN(*d.Radius) || math.IsInf(*d.Radius, 0) || *d.Radius < 0))) {
		return invalid
	}
	if choice.Place {
		found := false
		for _, c := range document.Candidates {
			if d.CandidateID != nil && c.ID == *d.CandidateID {
				found = mirrorText(c.PlaceName, 8192)
			}
		}
		if !found {
			return invalid
		}
	}
	for i, tag := range choice.Languages {
		parsed, err := language.Parse(tag)
		if err != nil || parsed.String() != tag || slices.Contains(choice.Languages[:i], tag) {
			return invalid
		}
		matches := 0
		for _, description := range d.Descriptions {
			if description.Language == tag {
				matches++
				if description.Status != "complete" || description.Text == nil || description.Stale || (description.Basis != "scene_only" && description.FactsRevision != d.FactsRevision) || !mirrorText(*description.Text, 16<<10) {
					return invalid
				}
			}
		}
		if matches != 1 {
			return invalid
		}
	}
	if choice.Provenance {
		if provenance.Mode != "visual" && provenance.Mode != "context-assisted" && provenance.Mode != "research" {
			return invalid
		}
		if _, err := time.Parse(time.RFC3339Nano, d.UpdatedAt); err != nil || (choice.Model && !mirrorText(provenance.Model, 256)) {
			return invalid
		}
	}
	return nil
}

func mirrorText(text string, limit int) bool {
	return len(text) <= limit && utf8.ValidString(text) && strings.TrimSpace(text) != ""
}
