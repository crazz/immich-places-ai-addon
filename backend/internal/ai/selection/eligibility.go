package selection

type Candidate struct {
	Available  bool
	Type       string
	Hidden     bool
	StackChild bool
	InScope    bool
}

func ExclusionReason(candidate Candidate) string {
	switch {
	case !candidate.Available:
		return "unavailable"
	case candidate.Type != "IMAGE":
		return "unsupported_type"
	case candidate.Hidden:
		return "hidden_by_policy"
	case candidate.StackChild:
		return "stack_child"
	case !candidate.InScope:
		return "outside_scope"
	default:
		return ""
	}
}
