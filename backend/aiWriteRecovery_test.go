package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestAIWriteRecoversAppliedLostResponseAfterReopenWithoutResend(t *testing.T) {
	w := newAIWriteFixture(t)
	deny := true
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" {
			w.mu.Lock()
			w.meta["exifInfo"] = map[string]float64{"latitude": 0, "longitude": 12}
			w.mu.Unlock()
			conn, _, err := out.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return true
			}
			conn.Close()
			return true
		}
		if deny {
			w.mu.Lock()
			sent := w.sends > 0
			w.mu.Unlock()
			if sent {
				out.WriteHeader(503)
				return true
			}
		}
		return false
	}
	_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
	if op := w.status(t); op.Status != "verifying" || op.Attempts != 1 {
		t.Fatal(op)
	}
	deny = false
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.writer.sync = newSyncService(w.f.db, nil, nil)
	w.f.now = w.f.now.Add(time.Minute)
	w.run(t)
	op := w.status(t)
	if op.Status != "succeeded" || !op.Verified || !op.Refreshed || op.Attempts != 1 {
		t.Fatal("not recovered", op)
	}
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal("resent", sends)
	}
}

func TestAIWriteReadbackDistinguishesUncertaintyConflictAndTolerance(t *testing.T) {
	for _, kind := range []string{"unknown-baseline", "rejected-baseline", "third-value", "rounded", "outside-tolerance", "changed-source"} {
		t.Run(kind, func(t *testing.T) {
			w := newAIWriteFixture(t)
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method != "PATCH" {
					return false
				}
				w.mu.Lock()
				switch kind {
				case "rounded":
					w.meta["exifInfo"] = map[string]float64{"latitude": 0.00000005, "longitude": 12.00000005}
				case "outside-tolerance":
					w.meta["exifInfo"] = map[string]float64{"latitude": 0.0000002, "longitude": 12}
				case "third-value":
					w.meta["exifInfo"] = map[string]float64{"latitude": 3, "longitude": 4}
				case "changed-source":
					w.meta["checksum"] = "AgICAgICAgICAgICAgICAgICAgI="
				}
				w.mu.Unlock()
				if kind == "unknown-baseline" {
					conn, _, _ := out.(http.Hijacker).Hijack()
					conn.Close()
				} else if kind == "rejected-baseline" {
					out.WriteHeader(429)
				} else {
					out.WriteHeader(200)
				}
				return true
			}
			w.run(t)
			op := w.status(t)
			want := "conflict"
			if kind == "rounded" {
				want = "succeeded"
			}
			if kind == "unknown-baseline" {
				want = "verifying"
			}
			if kind == "rejected-baseline" {
				want = "retryable"
			}
			if op.Status != want || op.Observed == nil {
				t.Fatal(kind, op)
			}
			if kind == "unknown-baseline" && op.Code != "RECONCILIATION_REQUIRED" {
				t.Fatal(op)
			}
			if sends, _ := w.counts(); sends != 1 {
				t.Fatal(sends)
			}
		})
	}
}

func TestAIWriteReadOnlyRecoveryIsBoundedAndNeverResendsUnknownBaseline(t *testing.T) {
	w := newAIWriteFixture(t)
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		conn, _, _ := out.(http.Hijacker).Hijack()
		conn.Close()
		return true
	}
	w.run(t)
	_, before := w.counts()
	for _, delay := range []time.Duration{time.Second, 5 * time.Second, 15 * time.Second, time.Hour, time.Hour} {
		w.f.now = w.f.now.Add(delay)
		w.run(t)
	}
	sends, reads := w.counts()
	if sends != 1 || reads-before != 3 {
		t.Fatal("unbounded recovery", sends, reads-before)
	}
	op := w.status(t)
	if op.Status != "verifying" || op.Code != "RECONCILIATION_REQUIRED" || op.Attempts != 1 {
		t.Fatal(op)
	}
	var guards int
	if err := w.f.db.db.QueryRow("SELECT count(*) FROM ai_write_target_guards").Scan(&guards); err != nil || guards != 1 {
		t.Fatal("uncertain target released", guards, err)
	}
}

func TestAIWriteCheckStatusStartsOnlyBoundedReadsWhileDisabled(t *testing.T) {
	w := newAIWriteFixture(t)
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		conn, _, _ := out.(http.Hijacker).Hijack()
		conn.Close()
		return true
	}
	w.run(t)
	for range 3 {
		w.f.now = w.f.now.Add(time.Minute)
		w.run(t)
	}
	w.enabled = false
	h := newAIResultHandler(w.writer.drafts.results, w.image.service)
	rec := aiRequest(h, "POST", "/ai/write-operations/"+w.op.ID+"/reconcile", `{}`, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	_, before := w.counts()
	for range 4 {
		w.f.now = w.f.now.Add(time.Minute)
		w.run(t)
	}
	sends, reads := w.counts()
	if sends != 1 || reads-before != 3 {
		t.Fatal(sends, reads-before)
	}
}
