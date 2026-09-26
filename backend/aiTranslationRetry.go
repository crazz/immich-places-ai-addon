package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/translations"
)

func (s *aiTranslationStore) validateParent(ctx context.Context, tx *sql.Tx, owner string, req translations.Request) error {
	if req.ParentID == "" {
		return nil
	}
	parent, err := s.read(ctx, tx, owner, req.ParentID)
	if err != nil || parent.Request.DraftID != req.DraftID {
		return drafts.ErrUnavailable
	}
	eligible := map[string]bool{}
	for _, item := range parent.Items {
		eligible[item.Language] = item.State == "failed" || item.State == "unavailable" || item.State == "canceled" || item.State == "interrupted"
	}
	for _, tag := range req.Languages {
		if !eligible[tag] {
			return drafts.ErrInvalid
		}
	}
	return nil
}
