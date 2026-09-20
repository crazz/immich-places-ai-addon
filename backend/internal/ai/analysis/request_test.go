package analysis_test

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/analysis"
)

func TestVisualRejectsInvalidBindingAndMissingAuthorityBeforeDispatch(t *testing.T) {
	cases := map[string]func(*analysis.Request){
		"owner": func(r *analysis.Request) { r.Owner = "other" }, "installation": func(r *analysis.Request) { r.Installation = "other" }, "asset": func(r *analysis.Request) { r.Asset = "other" }, "image": func(r *analysis.Request) { r.Image = nil }, "released": func(r *analysis.Request) { r.Image.Release() }, "guard": func(r *analysis.Request) { r.Guard.Authorize = nil }, "budget": func(r *analysis.Request) { r.Guard.Reserve = nil }, "revision": func(r *analysis.Request) { r.Revision = 0 }, "profile": func(r *analysis.Request) { r.ProfileID = "" }, "json permission": func(r *analysis.Request) { r.Format = "json" }, "format": func(r *analysis.Request) { r.Format = "text" }, "languages": func(r *analysis.Request) { r.Languages = []string{"en", "EN"} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			runner, req, calls, reservations := visualFixture(t)
			mutate(&req)
			result, err := runner.RunVisual(context.Background(), req)
			if err == nil || result != nil || *calls != 0 || *reservations != 0 {
				t.Fatal("invalid input dispatched or published")
			}
		})
	}
}
