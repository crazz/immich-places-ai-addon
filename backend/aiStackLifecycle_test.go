package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAIStackInstallationRotationFencesActiveAndUnstartedTargets(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	m := stackMutationServer(t, f)
	entered, release := make(chan struct{}), make(chan struct{})
	m.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" {
			close(entered)
			<-release
		}
		return false
	}
	done := make(chan error, 1)
	go func() { done <- f.writer.runOne(context.Background(), testUserID, op.ID) }()
	<-entered
	selectionSQL(t, f.f.db, `UPDATE ai_installation_identity SET id='replacement' WHERE singleton=1`)
	close(release)
	if err := <-done; err == nil {
		t.Fatal("obsolete writer published completion")
	}
	var active, verified, guards int
	if err := f.f.db.db.QueryRow(`SELECT count(*),sum(verified) FROM ai_stack_write_targets WHERE operationID=? AND status='writing'`, op.ID).Scan(&active, &verified); err != nil || active != 1 || verified != 0 {
		t.Fatal("old installation published target", err)
	}
	if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 3 {
		t.Fatal("old unresolved exclusions were discarded", err)
	}
	if _, err := f.writer.get(context.Background(), testUserID, op.ID, false); err == nil {
		t.Fatal("old installation remained readable")
	}
	if err := f.writer.runOne(context.Background(), testUserID, op.ID); err == nil {
		t.Fatal("old installation continued queued work")
	}
	for i, target := range op.Plan.Manifest.Targets {
		want := 0
		if i == 0 {
			want = 1
		}
		if m.sent(target.AssetID) != want {
			t.Fatal("obsolete installation sent another target")
		}
	}
}
