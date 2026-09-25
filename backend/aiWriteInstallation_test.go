package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAIWriteInstallationRotationFencesActiveSenderPublication(t *testing.T) {
	w := newAIWriteFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		close(entered)
		<-release
		return false
	}
	done := make(chan error, 1)
	go func() { done <- w.writer.runOne(context.Background(), testUserID, w.op.ID) }()
	<-entered
	selectionSQL(t, w.f.db, "UPDATE ai_installation_identity SET id='replacement' WHERE singleton=1")
	close(release)
	<-done
	var status string
	var verified, refreshed, guards int
	if err := w.f.db.db.QueryRow("SELECT o.status,t.verified,t.refreshed FROM ai_write_operations o JOIN ai_write_targets t ON t.operationID=o.id WHERE o.id=?", w.op.ID).Scan(&status, &verified, &refreshed); err != nil {
		t.Fatal(err)
	}
	if status != "writing" || verified != 0 || refreshed != 0 {
		t.Fatal("old installation published", status, verified, refreshed)
	}
	if err := w.f.db.db.QueryRow("SELECT count(*) FROM ai_write_target_guards").Scan(&guards); err != nil || guards != 1 {
		t.Fatal("uncertain old guard lost", guards, err)
	}
	if _, err := w.writer.get(context.Background(), testUserID, w.op.ID, false); err == nil {
		t.Fatal("old installation remained readable")
	}
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal(sends)
	}
}
