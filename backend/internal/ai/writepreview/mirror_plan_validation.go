package writepreview

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"
	"time"

	"golang.org/x/text/language"
)

func ValidMirrorPlan(plan MirrorPlan, revision int) bool {
	if plan.Key != MirrorNamespace || plan.Disclosure != MirrorDisclosure || ValidateMirrorOwnership(plan.Before, plan.RecordID, plan.Before.Value) != nil {
		return false
	}
	if _, err := parseMirrorJSON(plan.Value, MirrorValueLimit); err != nil {
		return false
	}
	var value MirrorExport
	decoder := json.NewDecoder(bytes.NewReader(plan.Value))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || value.SchemaVersion != 1 || value.Application != MirrorNamespace || value.RecordID != plan.RecordID || value.Review.DraftRevision != revision || value.Review.FactsRevision < 1 || value.Review.FactsRevision > revision {
		return false
	}
	canonical, err := json.Marshal(value)
	if err != nil || !bytes.Equal(plan.Value, canonical) {
		return false
	}
	choice := plan.Selection
	if choice.Direction != (value.Direction != nil) || choice.Precision != (value.Precision != nil) || choice.Place != (value.Place != nil) || choice.Provenance != (value.Provenance != nil) || (choice.Model && !choice.Provenance) || len(choice.Languages) > 8 || len(choice.Languages) != len(value.Descriptions) {
		return false
	}
	if !choice.Direction && !choice.Precision && !choice.Place && !choice.Provenance && len(choice.Languages) == 0 {
		return false
	}
	for i, tag := range choice.Languages {
		parsed, err := language.Parse(tag)
		if err != nil || parsed.String() != tag || slices.Contains(choice.Languages[:i], tag) || !mirrorText(value.Descriptions[tag], 16<<10) {
			return false
		}
	}
	if value.Place != nil && !mirrorText(*value.Place, 8192) {
		return false
	}
	if !validMirrorGeometry(value) {
		return false
	}
	if origin := value.Provenance; origin != nil {
		if (origin.Mode != "visual" && origin.Mode != "context-assisted" && origin.Mode != "research") || choice.Model != (origin.Model != "") || (choice.Model && !mirrorText(origin.Model, 256)) {
			return false
		}
		if _, err := time.Parse(time.RFC3339Nano, origin.ReviewedAt); err != nil {
			return false
		}
		fields := append([]string{}, choice.Languages...)
		for i := range fields {
			fields[i] = "description:" + fields[i]
		}
		for name, selected := range map[string]bool{"direction": choice.Direction, "precision": choice.Precision, "place": choice.Place} {
			if selected {
				fields = append(fields, name)
			}
		}
		if len(origin.Fields) != len(fields) {
			return false
		}
		for _, field := range fields {
			if origin.Fields[field] != "user" && origin.Fields[field] != "model" {
				return false
			}
		}
	}
	return true
}

func validMirrorGeometry(value MirrorExport) bool {
	if direction := value.Direction; direction != nil {
		if !slices.Contains([]string{"unknown", "user_supplied", "visual_estimate", "known_viewpoint_alignment"}, direction.Method) {
			return false
		}
		if direction.Heading == nil {
			if direction.Method != "unknown" || direction.Uncertainty != nil {
				return false
			}
		} else if !finite(*direction.Heading, 360) || *direction.Heading < 0 || *direction.Heading >= 360 {
			return false
		}
		if direction.Uncertainty != nil && (!finite(*direction.Uncertainty, 180) || *direction.Uncertainty < 0 || direction.Method == "unknown" || direction.Method == "user_supplied") {
			return false
		}
	}
	if precision := value.Precision; precision != nil {
		if !slices.Contains([]string{"unknown", "visual_estimate", "context_extent", "source_reported", "model_estimate"}, precision.Basis) {
			return false
		}
		if precision.Radius == nil {
			return precision.Basis == "unknown"
		}
		if math.IsNaN(*precision.Radius) || math.IsInf(*precision.Radius, 0) || *precision.Radius < 0 || precision.Basis == "unknown" {
			return false
		}
	}
	return true
}
