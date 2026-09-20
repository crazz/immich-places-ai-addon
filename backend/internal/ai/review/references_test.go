package review

import (
	"testing"

	"immich-places-backend/internal/ai/results"
)

func TestResearchReferencesAreInertAndExcludeUnsafeDestinations(t *testing.T) {
	for _, raw := range []string{"https://user:secret@example.org/photo", "javascript:alert(1)", "http://127.0.0.1/a", "http://[::1]/a", "http://10.0.0.1/a", "http://192.168.1.1/a", "http://169.254.169.254/latest", "http://100.100.100.100/a", "http://localhost./a", "http://nas.local/a", "http://pt-nas/a", "http://127.1/a", "http://0177.0.0.1/a", "https://example.org/%0aheader", "not a URL", "//example.org/photo"} {
		sources := []results.AnswerSource{{ID: "ref", URL: raw, Relevance: "<script>inert text</script>"}}
		presented := PresentSources(sources)
		if len(presented) != 1 || presented[0].URL != "" || presented[0].Relevance != sources[0].Relevance || sources[0].URL != raw {
			t.Fatal("unsafe reference or mutated retained answer")
		}
	}
	sources := []results.AnswerSource{{ID: "ref", URL: "https://example.org/reference?q=photo#detail", Relevance: "Facade comparison."}}
	if got := PresentSources(sources); len(got) != 1 || got[0].URL != sources[0].URL {
		t.Fatal("public answer link discarded")
	}
}
