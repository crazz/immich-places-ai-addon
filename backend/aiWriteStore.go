package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"immich-places-backend/internal/ai/writepreview"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

type aiWriteStore struct {
	capabilities writeback.CapabilityPolicy
	mu           sync.Mutex
	active       map[string]string
	images       *aiImagePreparer
	sync         *SyncService
	drafts       *aiDraftStore
	enabled      func() bool
	profile      string
}

func (s *aiWriteStore) available() bool {
	return s != nil && s.enabled != nil && s.enabled() && s.profile == "immich-v3.2.2"
}

func (s *aiWriteStore) confirm(ctx context.Context, owner string, input writeback.Confirmation) (writeback.Operation, error) {
	var result writeback.Operation
	if _, err := uuid.Parse(input.PreviewID); err != nil || len(input.Digest) != 64 || len(input.Key) < 1 || len(input.Key) > 128 || strings.TrimSpace(input.Key) != input.Key {
		return result, writeback.Failure("INVALID_WRITE")
	}
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		old, err := s.read(ctx, tx, owner, input.Key, true)
		if err == nil {
			if old.Plan.ID != input.PreviewID || old.Digest != input.Digest {
				return writeback.Failure("IDEMPOTENCY_CONFLICT")
			}
			result = old
			return nil
		}
		if !errors.Is(err, drafts.ErrUnavailable) {
			return err
		}
		if !s.available() {
			return writeback.Failure("WRITE_DISABLED")
		}
		raw, plan, digest, err := s.approvalPlan(ctx, tx, owner, input)
		if err != nil {
			return err
		}
		if !s.permitsPlan(plan) {
			return writeback.Failure("WRITE_DISABLED")
		}
		if plan.Version == "stack-preview-v3" || plan.Version == "mirror-preview-v4" {
			result, err = s.confirmStack(ctx, tx, owner, input, plan, raw, digest)
			return err
		}
		result.Plan = plan

		result.ID = uuid.NewString()
		result.Digest = digest
		result.Status = "queued"
		result.Settled = true
		now := s.drafts.results.jobs.now()
		result.ApprovedAt = now.UTC().Format(time.RFC3339Nano)
		key, err := s.credential(ctx, tx, owner)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_write_target_guards(installationID,assetID,token) VALUES(?,?,?)`, result.Plan.Installation, result.Plan.TargetID, result.ID)
		if err != nil {
			return writeback.Failure("TARGET_BUSY")
		}
		_, err = tx.ExecContext(ctx, aiWriteStatement(plan, `INSERT INTO ai_write_operations(userID,installationID,id,previewID,draftID,revision,assetID,idempotencyKey,payload,digest,approvedAt,credentialHash) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`), owner, result.Plan.Installation, result.ID, input.PreviewID, result.Plan.DraftID, result.Plan.DraftRevision, result.Plan.TargetID, input.Key, string(raw), digest, now.UnixNano(), aiWriteCredentialHash(key))
		if err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, aiWriteStatement(plan, `INSERT INTO ai_write_targets(userID,installationID,operationID) VALUES(?,?,?)`), owner, result.Plan.Installation, result.ID); err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, aiWriteStatement(plan, `UPDATE ai_write_previews SET protected=1 WHERE userID=? AND installationID=? AND id=?`), owner, result.Plan.Installation, input.PreviewID); err != nil {
			return drafts.ErrStorage
		}
		if err = s.event(ctx, tx, result, "approved"); err != nil {
			return err
		}
		result.Events = []writeback.Event{{Code: "approved", At: result.ApprovedAt, Attempt: 0}}
		if plan.Version == "standard-preview-v2" {
			for _, field := range plan.Fields {
				result.Fields = append(result.Fields, writeback.FieldOutcome{Field: field, Status: "pending"})
			}
		}
		return nil
	})
	return result, err
}

func aiWriteCredentialHash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func (s *aiWriteStore) credential(ctx context.Context, tx *sql.Tx, owner string) (string, error) {
	var user UserRow
	if tx.QueryRowContext(ctx, `SELECT immichAPIKey FROM users WHERE ID=?`, owner).Scan(&user.ImmichAPIKey) != nil || user.ImmichAPIKey == nil || s.drafts.results.jobs.db.decryptUserSecrets(&user) != nil || *user.ImmichAPIKey == "" {
		return "", drafts.ErrUnavailable
	}
	return *user.ImmichAPIKey, nil
}

func (s *aiWriteStore) event(ctx context.Context, tx *sql.Tx, op writeback.Operation, code string) error {
	_, err := tx.ExecContext(ctx, aiWriteStatement(op.Plan, `INSERT INTO ai_write_events(userID,installationID,operationID,code,at,attempt) SELECT ?,?,?,?,?,? WHERE (SELECT count(*) FROM ai_write_events WHERE userID=? AND installationID=? AND operationID=?)<100`), op.Plan.Owner, op.Plan.Installation, op.ID, code, s.drafts.results.jobs.now().UnixNano(), op.Attempts, op.Plan.Owner, op.Plan.Installation, op.ID)
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}

func (s *aiWriteStore) approvalPlan(ctx context.Context, tx *sql.Tx, owner string, input writeback.Confirmation) ([]byte, writepreview.Plan, string, error) {
	var raw []byte
	var digest, draftID string
	var revision int
	var expires int64
	var invalidated, protected bool
	var plan writepreview.Plan
	err := tx.QueryRowContext(ctx, `SELECT payload,digest,draftID,revision,expiresAt,invalidated,protected FROM ai_all_write_previews WHERE userID=? AND installationID=? AND id=?`, owner, s.drafts.results.jobs.binding, input.PreviewID).Scan(&raw, &digest, &draftID, &revision, &expires, &invalidated, &protected)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, plan, "", drafts.ErrUnavailable
	}
	if err != nil {
		return nil, plan, "", drafts.ErrStorage
	}
	plan, err = writeback.DecodePlan(raw, digest, owner, s.drafts.results.jobs.binding, input.PreviewID)
	if err != nil || plan.DraftID != draftID || plan.DraftRevision != revision {
		return nil, plan, "", drafts.ErrStorage
	}
	if input.Digest != digest {
		return nil, plan, "", writeback.Failure("PLAN_CONFLICT")
	}
	if protected {
		return nil, plan, "", writeback.Failure("PREVIEW_CONSUMED")
	}
	if expires <= s.drafts.results.jobs.now().UnixNano() {
		return nil, plan, "", writeback.Failure("PREVIEW_EXPIRED")
	}
	draft, err := s.drafts.read(ctx, tx, owner, draftID)
	if err != nil {
		return nil, plan, "", err
	}
	if invalidated || draft.State != "staged" || draft.Revision != revision {
		return nil, plan, "", writeback.Failure("DRAFT_CONFLICT")
	}
	if draft.AssetID != plan.TargetID || draft.AnalysisID != plan.AnalysisID || draft.Baseline.ImageIdentity != plan.ImageIdentity || !slices.Equal(draft.Fields, plan.Fields) {
		return nil, plan, "", drafts.ErrStorage
	}
	if slices.Contains(plan.Fields, "gps") && (draft.Camera == nil || draft.Camera.Latitude != plan.Intended.Latitude || draft.Camera.Longitude != plan.Intended.Longitude || !writepreview.EqualGPS(plan.Before, writepreview.GPS{Latitude: draft.Baseline.Latitude, Longitude: draft.Baseline.Longitude})) {
		return nil, plan, "", drafts.ErrStorage
	}
	if plan.Description != nil {
		owned, err := aiReadAppendLineage(ctx, tx, owner, plan.Installation, plan.TargetID)
		if err != nil || !aiApprovedDescriptionMatches(draft, *plan.Description, owned) {
			return nil, plan, "", drafts.ErrStorage
		}
	}
	if plan.Mirror != nil {
		if err := s.approvedMirrorMatches(ctx, tx, draft, plan); err != nil {
			return nil, plan, "", err
		}
	} else if draft.Mirror != nil {
		return nil, plan, "", drafts.ErrStorage
	}
	return raw, plan, digest, nil
}

func aiApprovedDescriptionMatches(draft drafts.Draft, plan writepreview.DescriptionPlan, owned *writepreview.AppendLineage) bool {
	text, ready := drafts.SelectedDescription(draft)
	if !ready || draft.PrimaryLanguage != plan.Language || draft.DescriptionPolicy != plan.Policy || draft.Baseline.Description == nil || draft.Baseline.Description.Value != plan.Before.Value {
		return false
	}
	input := writepreview.DescriptionInput{Text: text, Language: plan.Language, Policy: plan.Policy, Owned: owned}
	if plan.Lineage != nil {
		input.NewLineageID = plan.Lineage.ID
	}
	expected, err := writepreview.PlanDescription(plan.Before, input)
	return err == nil && reflect.DeepEqual(expected, &plan)
}
