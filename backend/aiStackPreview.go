package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiWritePreviewSession) createStack(ctx context.Context, owner, draftID string, revision int, reviewID string) (writepreview.Preview, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	snapshot, err := s.ReadDraft(ctx, owner, draftID)
	if err != nil {
		return writepreview.Preview{}, err
	}
	if err = writepreview.Validate(snapshot, revision); err != nil {
		return writepreview.Preview{}, err
	}
	snapshot.PolicyID = s.policyIdentity()
	review, err := s.drafts.readStackReview(ctx, owner, draftID, revision, reviewID)
	if err != nil {
		return writepreview.Preview{}, err
	}
	if len(review.Targets) == 1 {
		return writepreview.Create(ctx, s, s, s, owner, draftID, revision, uuid.NewString(), s.drafts.results.jobs.now)
	}
	primary, err := s.drafts.stackSource(ctx, owner, snapshot.AssetID, s.images)
	if err != nil || primary.stackID != review.StackID {
		return writepreview.Preview{}, drafts.ErrUnavailable
	}
	members, _, err := s.drafts.stackMembers(ctx, owner, primary, s.images)
	if err != nil {
		return writepreview.Preview{}, err
	}
	ids := make([]string, 0, len(review.Targets))
	for _, target := range review.Targets {
		ids = append(ids, target.AssetID)
	}
	ids, err = writepreview.SelectStackTargets(snapshot, ids, members)
	if err != nil {
		return writepreview.Preview{}, err
	}
	sources, err := s.drafts.readStackSources(ctx, owner, ids, review.StackID, s.images)
	if err != nil {
		return writepreview.Preview{}, err
	}
	current := make(map[string]writepreview.Metadata, len(ids))
	s.targetAuthorities = make(map[string]aiImageAuthority, len(ids))
	for id, source := range sources {
		current[id] = source.metadata
		s.targetAuthorities[id] = source.authority
	}
	s.authority = sources[snapshot.AssetID].authority
	if snapshot.Mirror != nil {
		value := current[snapshot.AssetID]
		value.Mirror, err = s.readMirror(ctx, owner, snapshot.AssetID, s.authority)
		if err != nil {
			return writepreview.Preview{}, err
		}
		current[snapshot.AssetID] = value
	}
	preview, raw, err := writepreview.BuildStack(snapshot, review, current, uuid.NewString(), s.drafts.results.jobs.now())
	if err != nil {
		return writepreview.Preview{}, err
	}
	if err = s.Publish(ctx, snapshot, preview, raw); err != nil {
		return writepreview.Preview{}, err
	}
	return preview, nil
}
