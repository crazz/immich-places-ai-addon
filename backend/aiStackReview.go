package main

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiDraftStore) observeStack(ctx context.Context, owner, id string, revision int, requested []string, reader *aiImagePreparer) (writepreview.StackReview, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var review writepreview.StackReview
	session := &aiWritePreviewSession{drafts: s, images: reader}
	snapshot, err := session.ReadDraft(ctx, owner, id)
	if err != nil {
		return review, err
	}
	if err = writepreview.Validate(snapshot, revision); err != nil {
		return review, err
	}
	primary, err := s.stackSource(ctx, owner, snapshot.AssetID, reader)
	if err != nil {
		return review, err
	}
	members, candidates, err := s.stackMembers(ctx, owner, primary, reader)
	if err != nil {
		return review, err
	}
	ids, err := writepreview.SelectStackTargets(snapshot, requested, members)
	if err != nil {
		return review, err
	}
	sources, err := s.readStackSources(ctx, owner, ids, primary.stackID, reader)
	if err != nil {
		return review, err
	}
	if err = writepreview.Compare(snapshot, sources[snapshot.AssetID].metadata); err != nil {
		return review, err
	}
	now := s.results.jobs.now()
	review = writepreview.StackReview{ID: uuid.NewString(), Owner: owner, Installation: snapshot.Installation, DraftID: id, DraftRevision: revision, AnalyzedID: snapshot.AssetID, ImageIdentity: snapshot.ImageIdentity, StackID: primary.stackID, ExpiresAt: now.Add(5 * time.Minute).UTC().Format(time.RFC3339Nano), Candidates: candidates, Targets: make([]writepreview.ReviewedTarget, 0, len(ids))}
	for _, id := range ids {
		source := sources[id]
		review.Targets = append(review.Targets, writepreview.ReviewedTarget{AssetID: id, ImageIdentity: source.metadata.ImageIdentity, Before: source.metadata.GPS})
	}
	if err = s.saveStackReview(ctx, review, sources); err != nil {
		return writepreview.StackReview{}, err
	}
	return review, nil
}

func (s *aiDraftStore) readStackSources(ctx context.Context, owner string, ids []string, stackID string, reader *aiImagePreparer) (map[string]aiStackSource, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		id     string
		source aiStackSource
		err    error
	}
	input := make(chan string, len(ids))
	output := make(chan result, len(ids))
	for _, id := range ids {
		input <- id
	}
	close(input)
	var group sync.WaitGroup
	for range min(4, len(ids)) {
		group.Add(1)
		go func() {
			defer group.Done()
			for id := range input {
				source, err := s.stackSource(ctx, owner, id, reader)
				if err == nil && source.stackID != stackID {
					err = drafts.ErrUnavailable
				}
				output <- result{id, source, err}
				if err != nil {
					cancel()
				}
			}
		}()
	}
	group.Wait()
	close(output)
	sources := make(map[string]aiStackSource, len(ids))
	for item := range output {
		if item.err != nil {
			return nil, drafts.ErrUnavailable
		}
		sources[item.id] = item.source
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return sources, nil
}
