package drafts

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func validFields(fields []string) bool {
	if fields == nil || len(fields) > 2 {
		return false
	}
	seen := make(map[string]bool, len(fields))
	for _, field := range fields {
		if (field != "gps" && field != "description") || seen[field] {
			return false
		}
		seen[field] = true
	}
	return true
}

func SelectedDescription(draft Draft) (string, bool) {
	if !slices.Contains(draft.Fields, "description") || (draft.DescriptionPolicy != "replace" && draft.DescriptionPolicy != "managed_append") {
		return "", false
	}
	for _, description := range draft.Descriptions {
		if description.Language != draft.PrimaryLanguage {
			continue
		}
		if description.Status != "complete" || description.Text == nil || description.Stale || (description.Basis != "scene_only" && description.FactsRevision != draft.FactsRevision) {
			return "", false
		}
		text := *description.Text
		return text, len(text) <= 16<<10 && utf8.ValidString(text) && strings.TrimSpace(text) != ""
	}
	return "", false
}

func Ready(draft Draft) bool {
	if !validFields(draft.Fields) || len(draft.Fields) == 0 {
		return false
	}
	if slices.Contains(draft.Fields, "gps") && (draft.Camera == nil || !validNumber(draft.Camera.Latitude, 90) || !validNumber(draft.Camera.Longitude, 180)) {
		return false
	}
	if slices.Contains(draft.Fields, "description") {
		_, ok := SelectedDescription(draft)
		return ok
	}
	return true
}
