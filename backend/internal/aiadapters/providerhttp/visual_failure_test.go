package providerhttp

import (
	"errors"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/providers"
)

func TestVisualFormatFailureClassificationIsExplicit(t *testing.T) {
	for _, status := range []int{400, 401, 403, 422, 429, 500} {
		f := providers.NewTransportFailure(providers.FailureUpstream, "request", status, errors.New("private-marker"))
		f.Code = "unsupported_response_format"
		err := ClassifyVisualFailure(f)
		if errors.Is(err, analysis.ErrUnsupportedFormat) != (status == 400 || status == 422) {
			t.Fatal("wrong downgrade eligibility", status, err)
		}
	}
}

func TestVisualAdmissionLimitRemainsNonTransient(t *testing.T) {
	if got := ClassifyVisualFailure(analysis.ErrLimit); got != analysis.ErrLimit {
		t.Fatalf("admission limit changed category: %v", got)
	}
}
