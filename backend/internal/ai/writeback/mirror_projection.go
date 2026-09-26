package writeback

func MirrorStandardReady(op Operation) bool {
	for _, target := range op.Targets {
		if target.AssetID == op.Plan.TargetID {
			return target.Status == "succeeded" && target.Verified && target.Settled
		}
	}
	return false
}

func ProjectMirror(op *Operation) {
	m := op.Mirror
	if m == nil {
		return
	}
	op.Verified = op.Verified && m.Verified
	op.Settled = op.Settled && m.Settled
	op.Noop = op.Noop && m.Noop
	if op.Status == "writing" || op.Status == "verifying" || op.Status == "queued" {
		return
	}
	switch m.Status {
	case "writing", "verifying", "queued":
		op.Status = m.Status
		op.Code = m.Code
	case "blocked":
		if MirrorStandardReady(*op) {
			op.Status, op.Code = "queued", "METADATA_PENDING"
		} else {
			op.Status, op.Code = "partial", "STANDARD_INCOMPLETE"
		}
	case "succeeded":
		if !op.Verified || !op.Settled {
			op.Status = "partial"
		}
	default:
		op.Status = "partial"
		op.Code = m.Code
	}
}
