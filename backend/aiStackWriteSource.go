package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (a *aiStackAttempt) Read(ctx context.Context, op writeback.Operation) (writepreview.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if op.Status == "queued" && a.assetID != a.parent.Plan.TargetID {
		analyzed, _, authority, err := a.readAsset(ctx, a.parent.Plan.TargetID, false)
		if err != nil || analyzed.ImageIdentity != a.parent.Plan.ImageIdentity {
			return writepreview.Metadata{}, writeback.Failure("SOURCE_CHANGED")
		}
		a.analyzedAuthority = authority
	}
	fresh, stackID, authority, err := a.readAsset(ctx, a.assetID, op.Status != "queued")
	if err != nil {
		return fresh, err
	}
	if op.Status == "queued" && a.parent.Plan.Manifest.StackID != "" && stackID != a.parent.Plan.Manifest.StackID {
		return fresh, writeback.Failure("STACK_MEMBERSHIP_CHANGED")
	}
	a.authority = authority
	if a.assetID == a.parent.Plan.TargetID {
		a.analyzedAuthority = authority
	}
	return fresh, nil
}

func (a *aiStackAttempt) readAsset(ctx context.Context, asset string, allowMissing bool) (writepreview.Metadata, string, aiImageAuthority, error) {
	var authority aiImageAuthority
	var empty writepreview.Metadata
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		authority, err = a.assetAuthority(ctx, tx, asset, allowMissing)
		return err
	})
	if err != nil || a.store.images == nil {
		return empty, "", authority, drafts.ErrUnavailable
	}
	raw, _, err := a.store.images.fetch(ctx, authority.key, "/api/assets/"+asset, 1<<20)
	if err != nil {
		return empty, "", authority, drafts.ErrUnavailable
	}
	defer clear(raw)
	fresh, stackID, err := aiParseWriteMetadata(raw, asset, true)
	if err != nil {
		return empty, "", authority, err
	}
	err = a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := a.assetAuthority(ctx, tx, asset, allowMissing)
		if err != nil || current != authority {
			return drafts.ErrUnavailable
		}
		return nil
	})
	return fresh, stackID, authority, err
}

func (a *aiStackAttempt) assetAuthority(ctx context.Context, tx *sql.Tx, asset string, allowMissing bool) (aiImageAuthority, error) {
	var authority aiImageAuthority
	key, err := a.store.credential(ctx, tx, a.parent.Plan.Owner)
	if err != nil {
		return authority, err
	}
	var hash string
	if tx.QueryRowContext(ctx, `SELECT credentialHash FROM ai_stack_write_operations WHERE userID=? AND installationID=? AND id=?`, a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID).Scan(&hash) != nil || hash != aiWriteCredentialHash(key) {
		return authority, drafts.ErrUnavailable
	}
	var facts struct {
		Type                    string
		Hidden                  bool
		Library, Stack, Primary *string
	}
	err = tx.QueryRowContext(ctx, `SELECT type,isHidden,libraryID,stackID,stackPrimaryAssetID FROM assets WHERE userID=? AND immichID=?`+hiddenLibraryFilter, a.parent.Plan.Owner, asset).Scan(&facts.Type, &facts.Hidden, &facts.Library, &facts.Stack, &facts.Primary)
	if errors.Is(err, sql.ErrNoRows) && allowMissing {
		var count int
		if tx.QueryRowContext(ctx, `SELECT count(*) FROM assets WHERE userID=? AND immichID=?`, a.parent.Plan.Owner, asset).Scan(&count) == nil && count == 0 {
			return aiImageAuthority{key: key, facts: "missing"}, nil
		}
	}
	if err != nil || facts.Type != "IMAGE" || facts.Hidden {
		return authority, drafts.ErrUnavailable
	}
	raw, err := json.Marshal(facts)
	if err != nil {
		return authority, drafts.ErrStorage
	}
	return aiImageAuthority{key: key, facts: string(raw)}, nil
}

func (a *aiStackAttempt) dispatchAuthority(ctx context.Context, tx *sql.Tx) error {
	plan := a.parent.Plan
	if !a.store.permitsPlan(plan) {
		return writeback.Failure("WRITE_DISABLED")
	}
	expires, err := time.Parse(time.RFC3339Nano, plan.ExpiresAt)
	if err != nil || !a.store.drafts.results.jobs.now().Before(expires) {
		return writeback.Failure("APPROVAL_EXPIRED")
	}
	draft, err := a.store.drafts.read(ctx, tx, plan.Owner, plan.DraftID)
	if err != nil || draft.Revision != plan.DraftRevision || draft.State != "staged" {
		return writeback.Failure("DRAFT_CONFLICT")
	}
	for asset, expected := range map[string]aiImageAuthority{plan.TargetID: a.analyzedAuthority, a.assetID: a.authority} {
		current, err := a.assetAuthority(ctx, tx, asset, false)
		if err != nil || current != expected {
			return drafts.ErrUnavailable
		}
	}
	return ctx.Err()
}
