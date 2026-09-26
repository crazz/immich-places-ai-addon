package writepreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"immich-places-backend/internal/ai/drafts"
)

const MirrorDisclosure = "asset-readers-v1"

type MirrorInput struct {
	RecordID   string
	Export     json.RawMessage
	Owned      json.RawMessage
	Selection  drafts.MirrorSelection
	PolicyID   string
	Disclosure string
}

type MirrorPlan struct {
	Key        string                 `json:"key"`
	RecordID   string                 `json:"recordId"`
	Before     MirrorBaseline         `json:"before"`
	Value      json.RawMessage        `json:"value"`
	Selection  drafts.MirrorSelection `json:"selection"`
	Disclosure string                 `json:"disclosure"`
}

func AttachMirror(base Preview, input MirrorInput, before MirrorBaseline) (Preview, []byte, error) {
	plan := base.Plan
	policy, err := hex.DecodeString(input.PolicyID)
	if err != nil || len(policy) != 32 || hex.EncodeToString(policy) != input.PolicyID || plan.Mirror != nil || len(plan.Fields) == 0 || len(plan.Fields) > 2 {
		return Preview{}, nil, Failure{Code: "INVALID_MIRROR"}
	}
	for i, field := range plan.Fields {
		if (field != "gps" && field != "description") || slices.Contains(plan.Fields[:i], field) {
			return Preview{}, nil, Failure{Code: "INVALID_MIRROR"}
		}
	}
	if err := ValidateMirrorOwnership(before, input.RecordID, input.Owned); err != nil {
		return Preview{}, nil, err
	}
	plan.Version, plan.ComparisonPolicy, plan.PolicyID = "mirror-preview-v4", "standard-then-metadata-v4", input.PolicyID
	if !slices.Contains(plan.Fields, "gps") {
		plan.Before, plan.Intended = GPS{}, Point{}
	}
	if plan.Manifest == nil {
		plan.Manifest = &TargetManifest{Targets: []Target{{AssetID: plan.TargetID, ImageIdentity: plan.ImageIdentity, Fields: slices.Clone(plan.Fields), Before: plan.Before, Intended: plan.Intended, Description: plan.Description}}}
	}
	choice := input.Selection
	choice.Languages = slices.Clone(choice.Languages)
	before.Value = bytes.Clone(before.Value)
	plan.Mirror = &MirrorPlan{Key: MirrorNamespace, RecordID: input.RecordID, Before: before, Value: bytes.Clone(input.Export), Selection: choice, Disclosure: input.Disclosure}
	if !ValidMirrorPlan(*plan.Mirror, plan.DraftRevision) {
		return Preview{}, nil, Failure{Code: "INVALID_MIRROR"}
	}
	raw, err := json.Marshal(plan)
	if err != nil || len(raw) > 1<<20 {
		return Preview{}, nil, Failure{Code: "INVALID_PREVIEW"}
	}
	hash := sha256.Sum256(raw)
	return Preview{Plan: plan, Digest: hex.EncodeToString(hash[:]), Status: "usable", Diff: Diff(plan)}, raw, nil
}
