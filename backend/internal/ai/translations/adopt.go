package translations

import (
	"immich-places-backend/internal/ai/drafts"
	"strings"
	"unicode/utf8"
)

func Adopt(current drafts.Draft, request Request, revision int, selected []Outcome) (drafts.Draft, error) {
	if current.Revision != revision || request.Revision != revision || current.FactsRevision != request.FactsRevision {
		return drafts.Draft{}, drafts.ErrConflict
	}
	if len(selected) < 1 || len(selected) > 8 {
		return drafts.Draft{}, drafts.ErrInvalid
	}
	seen := map[string]bool{}
	for _, out := range selected {
		if seen[out.Language] || out.Status != "complete" || out.Text == nil || strings.TrimSpace(*out.Text) == "" || len(*out.Text) > 16<<10 || !utf8.ValidString(*out.Text) {
			return drafts.Draft{}, drafts.ErrInvalid
		}
		seen[out.Language] = true
	}
	current.Descriptions = append([]drafts.Description{}, current.Descriptions...)
	basis := request.BasisKind
	if basis == "scene" {
		basis = "scene_only"
	}
	for _, out := range selected {
		value := drafts.Description{Language: out.Language, Status: "complete", Text: out.Text, Basis: basis, FactsRevision: current.FactsRevision}
		found := false
		for i := range current.Descriptions {
			if current.Descriptions[i].Language == out.Language {
				current.Descriptions[i] = value
				found = true
				break
			}
		}
		if !found {
			current.Descriptions = append(current.Descriptions, value)
		}
	}
	current.Revision++
	current.State = "draft"
	if len(current.Descriptions) > 20 {
		return drafts.Draft{}, drafts.ErrInvalid
	}
	return current, nil
}
