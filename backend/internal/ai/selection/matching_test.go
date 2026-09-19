package selection

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCollectMatchingClassifiesWholeStream(t *testing.T) {
	result, err := CollectMatching(500, func(yield func(Match) error) error {
		for _, candidate := range []Match{
			{AssetID: "newer", Candidate: Candidate{Available: true, Type: "IMAGE", InScope: true}, Facts: "first"},
			{AssetID: "video", Candidate: Candidate{Available: true, Type: "VIDEO", InScope: true}},
			{AssetID: "hidden", Candidate: Candidate{Available: true, Type: "IMAGE", Hidden: true, InScope: true}},
			{AssetID: "older", Candidate: Candidate{Available: true, Type: "IMAGE", InScope: true}, Facts: "last"},
		} {
			if err := yield(candidate); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || result.MatchedCount != 4 || result.EligibleCount != 2 || result.ExcludedCount != 2 ||
		result.ExclusionCounts["unsupported_type"] != 1 || result.ExclusionCounts["hidden_by_policy"] != 1 ||
		len(result.AssetIDs) != 2 || result.AssetIDs[0] != "newer" || result.AssetIDs[1] != "older" || result.Facts[1] != "last" {
		t.Fatalf("whole-stream classification: %+v %v", result, err)
	}
}

func TestCollectMatchingLimitsRetentionButFinishesCounts(t *testing.T) {
	result, err := CollectMatching(2, func(yield func(Match) error) error {
		for i := 0; i < 10000; i++ {
			if err := yield(Match{AssetID: "excluded", Candidate: Candidate{Available: true, Type: "VIDEO", InScope: true}}); err != nil {
				return err
			}
		}
		for i := 0; i < 3; i++ {
			if err := yield(Match{AssetID: "eligible", Candidate: Candidate{Available: true, Type: "IMAGE", InScope: true}, Facts: "fact"}); err != nil {
				return err
			}
		}
		return nil
	})
	var limit *MatchingLimitError
	if !errors.As(err, &limit) || limit.MatchedCount != 10003 || limit.EligibleCount != 3 || limit.ExcludedCount != 10000 || len(result.AssetIDs) != 0 {
		t.Fatalf("inexact or partial overflow: %+v %v", result, err)
	}
}

func TestCollectMatchingDropsProvisionalCountsOnFailure(t *testing.T) {
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded, errors.New("private database failure")} {
		result, err := CollectMatching(1, func(yield func(Match) error) error {
			if err := yield(Match{AssetID: "one", Candidate: Candidate{Available: true, Type: "IMAGE", InScope: true}, Facts: "fact"}); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) || result.MatchedCount != 0 || result.AssetIDs != nil {
			t.Fatalf("partial success: %+v %v", result, err)
		}
	}
}

func TestCollectMatchingBoundsRetainedFactBytes(t *testing.T) {
	result, err := CollectMatching(500, func(yield func(Match) error) error {
		return yield(Match{AssetID: "one", Candidate: Candidate{Available: true, Type: "IMAGE", InScope: true}, Facts: strings.Repeat("x", MaxBytes+1)})
	})
	if !errors.Is(err, ErrLimit) || result.MatchedCount != 0 || result.AssetIDs != nil {
		t.Fatalf("unbounded facts: count=%d err=%v", result.MatchedCount, err)
	}
}
