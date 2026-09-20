package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/selection"
)

func (s *aiSelectionStore) bind(ctx context.Context, endpoint, epoch string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || epoch == "" {
		return selection.ErrInvalid
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.ForceQuery = false
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Scheme == "https" && parsed.Port() == "443" || parsed.Scheme == "http" && parsed.Port() == "80" {
		parsed.Host = parsed.Hostname()
		if strings.Contains(parsed.Host, ":") {
			parsed.Host = "[" + parsed.Host + "]"
		}
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	sum := sha256.Sum256([]byte(parsed.String() + "\n" + epoch))
	fingerprint := hex.EncodeToString(sum[:])
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id, previous string
	err = tx.QueryRowContext(ctx, "SELECT id,fingerprint FROM ai_installation_identity WHERE singleton=1").Scan(&id, &previous)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if previous != fingerprint {
		if err = invalidateAIJobInstallation(ctx, tx, id, s.now()); err != nil {
			return err
		}
		id = uuid.NewString()
		if _, err = tx.ExecContext(ctx, "DELETE FROM ai_selection_snapshots"); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO ai_installation_identity (singleton,id,fingerprint) VALUES (1,?,?) ON CONFLICT(singleton) DO UPDATE SET id=excluded.id,fingerprint=excluded.fingerprint", id, fingerprint); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	s.binding = id
	return nil
}

func (s *aiSelectionStore) checkBinding(ctx context.Context, tx *sql.Tx) error {
	var current string
	if err := tx.QueryRowContext(ctx, "SELECT id FROM ai_installation_identity WHERE singleton=1").Scan(&current); err != nil {
		return err
	}
	if s.binding == "" || s.binding != current {
		return errAISelectionStale
	}
	return nil
}
