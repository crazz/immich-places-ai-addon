package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestParsePlainLabelsMatchesFixtureForm(t *testing.T) {
	color, shape, err := capabilities.ParsePlainLabels("color: blue\nshape: circle")
	if err != nil || color != "blue" || shape != "circle" {
		t.Fatalf("got %q %q err=%v", color, shape, err)
	}
	if _, _, err := capabilities.ParsePlainLabels("blue circle"); err == nil {
		t.Fatal("expected invalid plain labels to fail")
	}
}
