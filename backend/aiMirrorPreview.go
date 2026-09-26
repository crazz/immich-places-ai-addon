package main

import (
	"context"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiWritePreviewSession) mirrorInput(ctx context.Context, owner string, draft drafts.Draft) (*writepreview.MirrorInput, error) {
	detail, err := s.drafts.results.detail(ctx, owner, draft.AnalysisID, "", "")
	if err != nil || detail.Proposal == nil {
		return nil, drafts.ErrUnavailable
	}
	id, owned, err := s.drafts.mirrorRecord(ctx, owner, draft.AssetID)
	if err != nil {
		return nil, err
	}
	value, err := writepreview.BuildMirrorExport(draft, *detail.Proposal, *draft.Mirror, writepreview.MirrorProvenance{Mode: detail.Entry.Mode, Model: detail.Entry.Model}, id)
	if err != nil {
		return nil, err
	}
	return &writepreview.MirrorInput{RecordID: id, Export: value, Owned: owned, Selection: *draft.Mirror, PolicyID: s.policyIdentity(), Disclosure: s.mirrorDisclosure}, nil
}

func (s *aiWritePreviewSession) readMirror(ctx context.Context, owner, asset string, authority aiImageAuthority) (*writepreview.MirrorBaseline, error) {
	value, current, err := s.drafts.mirrorSource(ctx, owner, asset, s.images)
	if err != nil {
		return nil, err
	}
	if current != authority {
		return nil, drafts.ErrUnavailable
	}
	return &value, nil
}
