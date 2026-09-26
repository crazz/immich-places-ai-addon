package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestAIStandardUpgradePreservesPopulatedV1ApprovalsAndGuards(t *testing.T) {
	for _, state := range []string{"queued", "verified", "unknown"} {
		t.Run(state, func(t *testing.T) {
			w := newAIWriteFixture(t)
			ctx := context.Background()
			if state == "verified" {
				w.run(t)
			} else if state == "unknown" {
				attempt := &aiWriteAttempt{store: w.writer}
				if _, err := attempt.Read(ctx, w.op); err != nil {
					t.Fatal(err)
				}
				if err := attempt.Reserve(ctx, w.op, false); err != nil {
					t.Fatal(err)
				}
			}
			beforeState, err := json.Marshal(w.status(t))
			if err != nil {
				t.Fatal(err)
			}
			if err := goose.DownTo(w.f.db.db, "migrations", 28); err != nil {
				t.Fatal(err)
			}
			var raw, digest string
			if err := w.f.db.db.QueryRow(`SELECT payload,digest FROM ai_write_operations WHERE id=?`, w.op.ID).Scan(&raw, &digest); err != nil {
				t.Fatal(err)
			}
			var attempts, generation, guards int
			if err := w.f.db.db.QueryRow(`SELECT attempts,generation FROM ai_write_targets WHERE operationID=?`, w.op.ID).Scan(&attempts, &generation); err != nil {
				t.Fatal(err)
			}
			if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE token=?`, w.op.ID).Scan(&guards); err != nil {
				t.Fatal(err)
			}
			var audit string
			if err := w.f.db.db.QueryRow(`SELECT json_group_array(json_array(code,at,attempt)) FROM (SELECT code,at,attempt FROM ai_write_events WHERE operationID=? ORDER BY sequence)`, w.op.ID).Scan(&audit); err != nil {
				t.Fatal(err)
			}
			if err := runMigrations(w.f.db.db); err != nil {
				t.Fatal(err)
			}
			w.f.reopen(t)
			w.writer.drafts.results.jobs = w.f.store
			current := w.status(t)
			afterState, err := json.Marshal(current)
			if err != nil || string(afterState) != string(beforeState) {
				t.Fatal("retained outcome or recovery state changed", err)
			}
			after, err := json.Marshal(current.Plan)
			if err != nil || string(after) != raw || current.Digest != digest || current.Plan.TargetID != w.op.Plan.TargetID || current.Attempts != attempts || current.Generation != generation || current.Plan.Version != "gps-preview-v1" || current.Plan.Description != nil || len(current.Fields) != 0 {
				t.Fatal("legacy approval changed during upgrade", err)
			}
			var afterGuards int
			if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE token=?`, w.op.ID).Scan(&afterGuards); err != nil || afterGuards != guards {
				t.Fatal("legacy exclusion changed", err)
			}
			var afterAudit string
			if err := w.f.db.db.QueryRow(`SELECT json_group_array(json_array(code,at,attempt)) FROM (SELECT code,at,attempt FROM ai_write_events WHERE operationID=? ORDER BY sequence)`, w.op.ID).Scan(&afterAudit); err != nil || afterAudit != audit {
				t.Fatal("legacy audit changed", err)
			}
			if w.writer.capabilities.Allows(w.f.store.binding, w.writer.profile, "description") {
				t.Fatal("upgrade enabled description")
			}
			if state == "queued" {
				w.run(t)
				if w.status(t).Status != "succeeded" {
					t.Fatal("legacy queue cannot dispatch")
				}
			}
			if state == "unknown" && (current.Status != "writing" || guards != 1) {
				t.Fatal("ambiguous legacy operation lost protection")
			}
		})
	}
}
