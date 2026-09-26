package main

import (
	"testing"

	"github.com/pressly/goose/v3"
)

func TestAIMirrorStepUpgradeNeverCreatesOrAdmitsStepsForOldApprovals(t *testing.T) {
	f := stackWriteFixture(t)
	if err := goose.DownTo(f.f.db.db, "migrations", 43); err != nil {
		t.Fatal(err)
	}
	op := approveStackFixture(t, f, f.members)
	if err := runMigrations(f.f.db.db); err != nil {
		t.Fatal(err)
	}
	f.f.reopen(t)
	for _, table := range []string{"ai_mirror_write_steps", "ai_mirror_write_events", "ai_mirror_records"} {
		var count int
		if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("new step storage missing or old approval gained metadata state", table, err)
		}
	}
	if _, err := f.f.db.db.Exec(`INSERT INTO ai_mirror_write_steps(userID,installationID,operationID,assetID) VALUES(?,?,?,?)`, testUserID, f.f.store.binding, op.ID, op.Plan.TargetID); err == nil {
		t.Fatal("old immutable approval admitted a metadata step")
	}
}
