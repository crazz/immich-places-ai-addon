package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

type aiWritePreviewSession struct {
	mirrorDisclosure  string
	targetAuthorities map[string]aiImageAuthority
	drafts            *aiDraftStore
	images            *aiImagePreparer
	authority         aiImageAuthority
}

func (s *aiWritePreviewSession) ReadDraft(ctx context.Context, owner, id string) (writepreview.Snapshot, error) {
	draft, err := s.drafts.get(ctx, owner, id)
	if err != nil {
		return writepreview.Snapshot{}, err
	}
	if draft.Mirror != nil {
		writer := s.drafts.results.writer
		if writer == nil || !writer.capabilities.Allows(s.drafts.results.jobs.binding, writer.profile, "metadata") {
			return writepreview.Snapshot{}, writepreview.Failure{Code: "METADATA_UNSUPPORTED"}
		}
	}
	value := writepreview.Snapshot{Owner: owner, Installation: s.drafts.results.jobs.binding, ID: draft.ID, AnalysisID: draft.AnalysisID, AssetID: draft.AssetID, State: draft.State, Revision: draft.Revision, Fields: draft.Fields, BaselineReviewed: draft.Baseline.Status == "reviewed", ImageIdentity: draft.Baseline.ImageIdentity, Baseline: writepreview.GPS{Latitude: draft.Baseline.Latitude, Longitude: draft.Baseline.Longitude}}
	if draft.Camera != nil {
		value.Camera = &writepreview.Point{Latitude: draft.Camera.Latitude, Longitude: draft.Camera.Longitude}
	}
	value.DescriptionBaseline = aiPreviewText(draft.Baseline.Description)
	if text, selected := drafts.SelectedDescription(draft); selected {
		value.Description = &writepreview.DescriptionInput{Text: text, Language: draft.PrimaryLanguage, Policy: draft.DescriptionPolicy, NewLineageID: uuid.NewString()}
		value.PolicyID = s.policyIdentity()
		if draft.DescriptionPolicy == "managed_append" {
			err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
				var err error
				value.Description.Owned, err = aiReadAppendLineage(ctx, tx, owner, value.Installation, value.AssetID)
				return err
			})
			if err != nil {
				return writepreview.Snapshot{}, err
			}
		}
	}
	if draft.Mirror != nil {
		value.PolicyID = s.policyIdentity()
		value.Mirror, err = s.mirrorInput(ctx, owner, draft)
		if err != nil {
			return writepreview.Snapshot{}, err
		}
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
	value := writepreview.Metadata{ImageIdentity: baseline.ImageIdentity, GPS: writepreview.GPS{Latitude: baseline.Latitude, Longitude: baseline.Longitude}, Description: aiPreviewText(baseline.Description)}
	if draft.Mirror != nil {
		value.Mirror, err = s.readMirror(ctx, draft.Owner, draft.AssetID, authority)
		if err != nil {
			return writepreview.Metadata{}, err
		}
		s.targetAuthorities = map[string]aiImageAuthority{draft.AssetID: authority}
	}
	return value, nil
}

func aiPreviewText(value *drafts.DescriptionBaseline) *writepreview.TextObservation {
	if value == nil {
		return nil
	}
	return &writepreview.TextObservation{Presence: value.Presence, Value: value.Value}
}

func (s *aiWritePreviewSession) policyIdentity() string {
	var policy writeback.CapabilityPolicy
	profile := ""
	if writer := s.drafts.results.writer; writer != nil {
		policy, profile = writer.capabilities, writer.profile
	}
	return policy.Identity(s.drafts.results.jobs.binding, profile)
}
