package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

type aiWriteSummary struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Revision   int    `json:"draftRevision"`
	ApprovedAt string `json:"approvedAt"`
}

type aiWriteHistory struct {
	Items      []aiWriteSummary `json:"items"`
	NextCursor string           `json:"nextCursor"`
}

func registerAIWriteHistoryRoute(mux *http.ServeMux, store *aiWriteStore) {
	mux.HandleFunc("GET /ai/write-operations", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		draftID, cursor := q.Get("draftId"), q.Get("before")
		_, draftErr := uuid.Parse(draftID)
		_, cursorErr := uuid.Parse(cursor)
		if draftErr != nil || (cursor != "" && cursorErr != nil) || len(q["draftId"]) != 1 || len(q["before"]) > 1 || len(q) > 2 || (len(q) == 2 && len(q["before"]) != 1) {
			writeAIWriteFailure(w, writeback.Failure("INVALID_WRITE"))
			return
		}
		page, err := store.history(r.Context(), getUserFromContext(r).ID, draftID, cursor)
		if err != nil {
			writeAIWriteFailure(w, err)
			return
		}
		writeJSON(w, 200, page)
	})
}

func (s *aiWriteStore) history(ctx context.Context, owner, draftID, cursor string) (aiWriteHistory, error) {
	page := aiWriteHistory{Items: []aiWriteSummary{}}
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := s.drafts.read(ctx, tx, owner, draftID); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT id,status,revision,approvedAt FROM ai_all_write_operations WHERE userID=? AND installationID=? AND draftID=? AND (?='' OR id<?) ORDER BY id DESC LIMIT 101`, owner, s.drafts.results.jobs.binding, draftID, cursor, cursor)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item aiWriteSummary
			var at int64
			if err = rows.Scan(&item.ID, &item.Status, &item.Revision, &at); err != nil {
				return err
			}
			if len(page.Items) == 100 {
				page.NextCursor = page.Items[99].ID
				break
			}
			item.ApprovedAt = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
			page.Items = append(page.Items, item)
		}
		return rows.Err()
	})
	return page, err
}
