package jobs

import (
	"slices"
	"unicode/utf8"

	"immich-places-backend/internal/ai/contextual"
)

type ContextChoices struct {
	Version string             `json:"version"`
	Classes []contextual.Class `json:"classes"`
	Hint    string             `json:"hint,omitempty"`
	AlbumID string             `json:"albumId,omitempty"`
}

func normalizeContext(mode string, choice *ContextChoices) (*ContextChoices, error) {
	if mode == "visual" {
		if choice != nil {
			return nil, ErrInvalid
		}
		return nil, nil
	}
	if choice == nil || choice.Version != contextual.Version || choice.Classes == nil || len(choice.Classes) > 4 || len(choice.Hint) > 2000 || !utf8.ValidString(choice.Hint) {
		return nil, ErrInvalid
	}
	c := *choice
	c.Classes = slices.Clone(choice.Classes)
	slices.Sort(c.Classes)
	for i, class := range c.Classes {
		if (i > 0 && c.Classes[i-1] == class) || !slices.Contains([]contextual.Class{contextual.Capture, contextual.AlbumLabel, contextual.Hint, contextual.Neighbors}, class) {
			return nil, ErrInvalid
		}
	}
	if (c.Hint != "" && !slices.Contains(c.Classes, contextual.Hint)) || (c.AlbumID != "" && !slices.Contains(c.Classes, contextual.AlbumLabel)) || (slices.Contains(c.Classes, contextual.AlbumLabel) && !boundedIdentity(c.AlbumID)) {
		return nil, ErrInvalid
	}
	return &c, nil
}
