package review

import (
	"net/url"
	"testing"
)

func TestQueryHasFiniteDefaultsAndRejectsInvalidBoundaries(t *testing.T) {
	got, err := ParseQuery(url.Values{})
	if err != nil || got.Limit != 30 {
		t.Fatal("default page unavailable", got, err)
	}
	for _, raw := range []string{"limit=0", "limit=101", "limit=x", "limit=30&limit=40", "startDate=2026-02-02&endDate=2026-02-01", "startDate=2026-02-30", "state=running", "outcome=failure", "gps=missing", "assetId=bad", "jobId=bad", "undated=maybe", "undated=true&startDate=2026-01-01"} {
		values, _ := url.ParseQuery(raw)
		if _, err = ParseQuery(values); err == nil {
			t.Errorf("accepted invalid query %s", raw)
		}
	}
}
