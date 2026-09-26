package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"reflect"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiWriteStore) approvedMirrorMatches(ctx context.Context, tx *sql.Tx, draft drafts.Draft, plan writepreview.Plan) error {
	if draft.Mirror == nil || !reflect.DeepEqual(*draft.Mirror, plan.Mirror.Selection) {
		return drafts.ErrStorage
	}
	var payload, metadata []byte
	var jobID, outcome string
	err := tx.QueryRowContext(ctx, `SELECT a.payload,a.metadata,a.jobID,a.outcome FROM ai_analyses a JOIN ai_jobs j ON j.userID=a.userID AND j.id=a.jobID WHERE a.userID=? AND a.id=? AND j.installationID=?`, plan.Owner, plan.AnalysisID, plan.Installation).Scan(&payload, &metadata, &jobID, &outcome)
	if err != nil || len(payload) > 1<<20 || len(metadata) > 128<<10 {
		return drafts.ErrStorage
	}
	var provenance jobs.ResultMetadata
	if json.Unmarshal(metadata, &provenance) != nil {
		return drafts.ErrStorage
	}
	job, err := s.drafts.results.jobs.readJob(ctx, tx, plan.Owner, jobID)
	if err != nil {
		return drafts.ErrStorage
	}
	document, err := review.ValidateRecord(payload, provenance, job, plan.TargetID, outcome)
	if err != nil {
		return drafts.ErrStorage
	}
	expected, err := writepreview.BuildMirrorExport(draft, document, *draft.Mirror, writepreview.MirrorProvenance{Mode: string(provenance.Mode), Model: provenance.Model}, plan.Mirror.RecordID)
	if err != nil || !bytes.Equal(expected, plan.Mirror.Value) {
		return drafts.ErrStorage
	}
	return nil
}
