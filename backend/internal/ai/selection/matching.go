package selection

// Match is one distinct, authorized catalog candidate in stable catalog order.
// The adapter owns discovery and supplies opaque, scope-relevant facts.
type Match struct {
	AssetID   string
	Candidate Candidate
	Facts     string
}

type Enumerate func(yield func(Match) error) error

type MatchingCounts struct {
	MatchedCount    int            `json:"matchedCount"`
	EligibleCount   int            `json:"eligibleCount"`
	ExcludedCount   int            `json:"excludedCount"`
	ExclusionCounts map[string]int `json:"exclusionCounts"`
}

type MatchingResult struct {
	MatchingCounts
	AssetIDs   []string
	Facts      []string
	FactsBytes int
}

type MatchingLimitError struct{ MatchingCounts }

func (*MatchingLimitError) Error() string { return "matching selection exceeds eligible asset limit" }

func CollectMatching(maxAssets int, enumerate Enumerate) (MatchingResult, error) {
	result := MatchingResult{MatchingCounts: MatchingCounts{ExclusionCounts: map[string]int{}}, AssetIDs: []string{}, Facts: []string{}}
	err := enumerate(func(match Match) error {
		result.MatchedCount++
		if reason := ExclusionReason(match.Candidate); reason != "" {
			result.ExcludedCount++
			result.ExclusionCounts[reason]++
			return nil
		}
		result.EligibleCount++
		if result.EligibleCount > maxAssets {
			return nil
		}
		if result.FactsBytes+len(match.Facts) > MaxBytes {
			return ErrLimit
		}
		result.AssetIDs = append(result.AssetIDs, match.AssetID)
		result.Facts = append(result.Facts, match.Facts)
		result.FactsBytes += len(match.Facts)
		return nil
	})
	if err != nil {
		return MatchingResult{}, err
	}
	if result.EligibleCount > maxAssets {
		return MatchingResult{}, &MatchingLimitError{result.MatchingCounts}
	}
	return result, nil
}
