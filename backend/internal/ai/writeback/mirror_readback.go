package writeback

import "immich-places-backend/internal/ai/writepreview"

func MirrorReadback(plan writepreview.Plan, fresh writepreview.Metadata, completed bool, attempts int, wasVerified bool) Decision {
	if fresh.ImageIdentity != plan.ImageIdentity {
		return Decision{Status: "conflict", Code: "SOURCE_CHANGED"}
	}
	if plan.Mirror == nil || fresh.Mirror == nil {
		return Decision{Status: "verifying", Code: "METADATA_UNAVAILABLE"}
	}
	if fresh.Mirror.Present && writepreview.EqualMirrorValue(fresh.Mirror.Value, plan.Mirror.Value) {
		if !completed {
			return Decision{Status: "verifying", Code: "METADATA_OBSERVED_UNRESOLVED", Verified: true}
		}
		return Decision{Status: "succeeded", Code: "METADATA_VERIFIED", Verified: true}
	}
	if wasVerified || !EqualMirrorBaseline(*fresh.Mirror, plan.Mirror.Before) {
		return Decision{Status: "conflict", Code: "METADATA_CONFLICT"}
	}
	if !completed {
		return Decision{Status: "verifying", Code: "RECONCILIATION_REQUIRED"}
	}
	if attempts >= 2 {
		return Decision{Status: "failed", Code: "ATTEMPTS_EXHAUSTED"}
	}
	return Decision{Status: "retryable", Code: "METADATA_RETRY_AVAILABLE"}
}

func EqualMirrorBaseline(a, b writepreview.MirrorBaseline) bool {
	if a.Present != b.Present {
		return false
	}
	if !a.Present {
		return len(a.Value) == 0 && len(b.Value) == 0
	}
	return writepreview.EqualMirrorValue(a.Value, b.Value)
}
