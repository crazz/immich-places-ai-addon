package jobs

import (
	"immich-places-backend/internal/ai/contextual"
	"strings"
	"testing"
)

func TestContextChoicesRejectHiddenOrAmbiguousDisclosure(t *testing.T) {
	for _, tc := range []struct {
		name, mode string
		choice     *ContextChoices
	}{
		{"visual", "visual", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{}}},
		{"absent", "context-assisted", nil},
		{"null classes", "context-assisted", &ContextChoices{Version: "context-v1"}},
		{"duplicate", "context-assisted", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Hint, contextual.Hint}}},
		{"hidden hint", "context-assisted", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{}, Hint: "private"}},
		{"large UTF8 hint", "context-assisted", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Hint}, Hint: strings.Repeat("é", 1001)}},
		{"album scope missing", "context-assisted", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.AlbumLabel}}},
		{"research", "context-assisted", &ContextChoices{Version: "context-v1", Classes: []contextual.Class{"research"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := normalizeContext(tc.mode, tc.choice); err == nil {
				t.Fatal("unsafe disclosure accepted")
			}
		})
	}
}
