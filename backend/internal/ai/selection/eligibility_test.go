package selection

import "testing"

func TestMixedSelectionUsesDeterministicExclusions(t *testing.T) {
	for _, tc := range []struct {
		candidate Candidate
		reason    string
	}{
		{Candidate{}, "unavailable"},
		{Candidate{Available: true, Type: "VIDEO", Hidden: true, StackChild: true}, "unsupported_type"},
		{Candidate{Available: true, Type: "IMAGE", Hidden: true, StackChild: true}, "hidden_by_policy"},
		{Candidate{Available: true, Type: "IMAGE", StackChild: true}, "stack_child"},
		{Candidate{Available: true, Type: "IMAGE"}, "outside_scope"},
		{Candidate{Available: true, Type: "IMAGE", InScope: true}, ""},
	} {
		if got := ExclusionReason(tc.candidate); got != tc.reason {
			t.Fatalf("candidate %+v: reason %q, want %q", tc.candidate, got, tc.reason)
		}
	}
}
