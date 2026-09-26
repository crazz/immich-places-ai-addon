package writeback

func TargetStep(parent Operation, asset string) (Operation, error) {
	if (parent.Plan.Version != "stack-preview-v3" && parent.Plan.Version != "mirror-preview-v4") || parent.Plan.Manifest == nil || len(parent.Targets) != len(parent.Plan.Manifest.Targets) {
		return Operation{}, Failure("WRITE_UNAVAILABLE")
	}
	for i, target := range parent.Plan.Manifest.Targets {
		if target.AssetID != asset {
			continue
		}
		state := parent.Targets[i]
		if state.AssetID != asset {
			return Operation{}, Failure("WRITE_UNAVAILABLE")
		}
		plan := parent.Plan
		plan.Manifest = nil
		plan.Mirror = nil
		plan.Version = "stack-preview-v3"
		plan.TargetID, plan.ImageIdentity = target.AssetID, target.ImageIdentity
		plan.Fields, plan.Before, plan.Intended, plan.Description = target.Fields, target.Before, target.Intended, target.Description
		return Operation{ID: parent.ID, Plan: plan, Digest: parent.Digest, ApprovedAt: parent.ApprovedAt, Status: state.Status, Code: state.Code, Attempts: state.Attempts, Generation: state.Generation, Observed: state.Observed, Verified: state.Verified, Refreshed: state.Refreshed, Noop: state.Noop, Settled: state.Settled, Fields: state.Fields, Events: state.Events}, nil
	}
	return Operation{}, Failure("WRITE_UNAVAILABLE")
}

func ProjectTargets(op *Operation) {
	if len(op.Targets) == 0 {
		return
	}
	op.Status = op.Targets[0].Status
	op.Verified, op.Refreshed, op.Noop, op.Settled = true, true, true, true
	active := ""
	for _, target := range op.Targets {
		op.Verified = op.Verified && target.Verified
		op.Refreshed = op.Refreshed && target.Refreshed
		op.Noop = op.Noop && target.Noop
		op.Settled = op.Settled && target.Settled
		if op.Status != target.Status {
			op.Status = "partial"
		}
		switch target.Status {
		case "writing":
			active = "writing"
		case "verifying":
			if active != "writing" {
				active = "verifying"
			}
		case "queued":
			if active == "" {
				active = "queued"
			}
		}
	}
	if active != "" {
		op.Status = active
	}
}
