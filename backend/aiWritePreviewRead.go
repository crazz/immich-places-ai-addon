package main

import (
	"context"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

type aiWritePreviewSession struct {
	drafts    *aiDraftStore
	images    *aiImagePreparer
	authority aiImageAuthority
}

func (s *aiWritePreviewSession) ReadDraft(ctx context.Context, owner, id string) (writepreview.Snapshot, error) {
	draft, err := s.drafts.get(ctx, owner, id)
	if err != nil {
		return writepreview.Snapshot{}, err
	}
	value := writepreview.Snapshot{Owner: owner, Installation: s.drafts.results.jobs.binding, ID: draft.ID, AnalysisID: draft.AnalysisID, AssetID: draft.AssetID, State: draft.State, Revision: draft.Revision, Fields: draft.Fields, BaselineReviewed: draft.Baseline.Status == "reviewed", ImageIdentity: draft.Baseline.ImageIdentity, Baseline: writepreview.GPS{Latitude: draft.Baseline.Latitude, Longitude: draft.Baseline.Longitude}}
	if draft.Camera != nil {
		value.Camera = &writepreview.Point{Latitude: draft.Camera.Latitude, Longitude: draft.Camera.Longitude}
	}
	return value, nil
}

func (s *aiWritePreviewSession) ReadMetadata(ctx context.Context, draft writepreview.Snapshot) (writepreview.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := s.drafts.results.imageAuthority(ctx, draft.Owner, draft.AssetID); err != nil {
		return writepreview.Metadata{}, drafts.ErrUnavailable
	}
	baseline, authority, err := s.drafts.source(ctx, draft.Owner, draft.AssetID, s.images)
	if err != nil {
		return writepreview.Metadata{}, writepreview.Failure{Code: "SOURCE_UNAVAILABLE"}
	}
	s.authority = authority
	return writepreview.Metadata{ImageIdentity: baseline.ImageIdentity, GPS: writepreview.GPS{Latitude: baseline.Latitude, Longitude: baseline.Longitude}}, nil
}
