package main

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/review"
)

type aiResultCursor struct {
	Version, Owner, Installation string
	Query                        review.Query
	Watermark, TerminalAt        int64
	Job, Item                    string
}

func (s *aiResultStore) cursor(owner string, q review.Query) (aiResultCursor, error) {
	raw := q.Cursor
	q.Cursor = ""
	c := aiResultCursor{Version: "results-v1", Owner: owner, Installation: s.jobs.binding, Query: q}
	if raw == "" {
		return c, nil
	}
	if len(raw) > 2048 || !strings.HasPrefix(raw, encryptedPrefix) {
		return c, review.ErrInvalid
	}
	data, err := decryptValue(s.jobs.db.encryptionKey, raw)
	if err != nil || json.Unmarshal([]byte(data), &c) != nil || c.Version != "results-v1" || c.Owner != owner || c.Installation != s.jobs.binding || c.Query != q || c.Watermark < 1 {
		return c, review.ErrInvalid
	}
	for _, id := range []string{c.Job, c.Item} {
		if _, err := uuid.Parse(id); err != nil {
			return c, review.ErrInvalid
		}
	}
	return c, nil
}

func (s *aiResultStore) encodeCursor(c aiResultCursor) (string, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return "", review.ErrUnavailable
	}
	encoded, err := encryptValue(s.jobs.db.encryptionKey, string(raw))
	if err != nil {
		return "", review.ErrUnavailable
	}
	return encoded, nil
}
