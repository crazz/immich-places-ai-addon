package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteRetryReplayAfterSuccessIsLocal(t *testing.T) {
	w := newAIWriteFixture(t)
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		sends, _ := w.counts()
		if sends == 1 {
			out.WriteHeader(429)
			return true
		}
		return false
	}
	w.run(t)
	op := w.status(t)
	if op.Status != "retryable" {
		t.Fatal(op.Status)
	}
	h := newAIResultHandler(w.writer.drafts.results, w.image.service)
	body := fmt.Sprintf(`{"generation":%d}`, op.Generation)
	path := "/ai/write-operations/" + op.ID + "/retry"
	first := aiRequest(h, "POST", path, body, aiTestOrigin, true)
	if first.Code != 200 {
		t.Fatal(first.Code, first.Body.String())
	}
	w.run(t)
	op = w.status(t)
	if op.Status != "succeeded" || op.Attempts != 2 {
		t.Fatal(op)
	}
	sends, reads := w.counts()
	repeated := aiRequest(h, "POST", path, body, aiTestOrigin, true)
	if repeated.Code != 200 {
		t.Fatal(repeated.Code, repeated.Body.String())
	}
	var recovered writeback.Operation
	if err := json.Unmarshal(repeated.Body.Bytes(), &recovered); err != nil {
		t.Fatal(err)
	}
	if recovered.ID != op.ID || recovered.Status != op.Status || recovered.Generation != op.Generation || recovered.Attempts != op.Attempts {
		t.Fatal("did not recover current operation", recovered)
	}
	if afterSends, afterReads := w.counts(); sends != 2 || afterSends != sends || afterReads != reads {
		t.Fatal("retry replay contacted Immich", sends, reads, afterSends, afterReads)
	}
}

func TestAIWriteRetryConcurrentReplayRecoversAfterReadLosesEligibility(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		t.Run(fmt.Sprint(unavailable), func(t *testing.T) {
			w := newAIWriteFixture(t)
			var holdNext atomic.Bool
			entered, release := make(chan struct{}), make(chan struct{})
			releaseRead := sync.OnceFunc(func() { close(release) })
			defer releaseRead()
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PATCH" {
					if sends, _ := w.counts(); sends == 1 {
						out.WriteHeader(429)
						return true
					}
				}
				if r.Method == "GET" && holdNext.CompareAndSwap(true, false) {
					close(entered)
					<-release
					if unavailable {
						out.WriteHeader(503)
						return true
					}
				}
				return false
			}
			w.run(t)
			op := w.status(t)
			h := newAIResultHandler(w.writer.drafts.results, w.image.service)
			body := fmt.Sprintf(`{"generation":%d}`, op.Generation)
			path := "/ai/write-operations/" + op.ID + "/retry"
			holdNext.Store(true)
			result := make(chan int, 1)
			go func() { result <- aiRequest(h, "POST", path, body, aiTestOrigin, true).Code }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("retry metadata read did not start")
			}
			accepted := aiRequest(h, "POST", path, body, aiTestOrigin, true)
			if accepted.Code != 200 {
				t.Fatal(accepted.Code, accepted.Body.String())
			}
			w.run(t)
			if current := w.status(t); current.Status != "succeeded" || current.Attempts != 2 {
				t.Fatal(current)
			}
			releaseRead()
			select {
			case status := <-result:
				if status != 200 {
					t.Fatal("concurrent accepted retry was not recovered", status)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("retry replay did not finish")
			}
			if sends, _ := w.counts(); sends != 2 {
				t.Fatal("duplicate retry allocated another send", sends)
			}
		})
	}
}

func TestAIWriteRetryReplaySurvivesReopenAndAuthorityLoss(t *testing.T) {
	w := newAIWriteFixture(t)
	var offline atomic.Bool
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" {
			out.WriteHeader(429)
			return true
		}
		if offline.Load() {
			out.WriteHeader(503)
			return true
		}
		return false
	}
	w.run(t)
	op := w.status(t)
	accepted, err := w.writer.retry(context.Background(), testUserID, op.ID, op.Generation)
	if err != nil || accepted.Status != "queued" {
		t.Fatal(accepted, err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.writer.sync = newSyncService(w.f.db, nil, nil)
	w.f.now = w.f.now.Add(6 * time.Minute)
	w.mu.Lock()
	w.enabled = false
	w.mu.Unlock()
	offline.Store(true)
	sends, reads := w.counts()
	recovered, err := w.writer.retry(context.Background(), testUserID, op.ID, op.Generation)
	if err != nil || recovered.ID != accepted.ID || recovered.Generation != accepted.Generation || recovered.Status != accepted.Status || recovered.Attempts != 1 {
		t.Fatal("accepted retry was not recovered after restart", recovered, err)
	}
	if afterSends, afterReads := w.counts(); afterSends != sends || afterReads != reads {
		t.Fatal("local recovery contacted unavailable Immich", sends, reads, afterSends, afterReads)
	}
}

func TestAIWriteRetryReplayRetainsOwnerInstallationAndGenerationScope(t *testing.T) {
	for _, scope := range []string{"owner", "installation", "generation"} {
		t.Run(scope, func(t *testing.T) {
			w := newAIWriteFixture(t)
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PATCH" {
					out.WriteHeader(429)
					return true
				}
				return false
			}
			w.run(t)
			op := w.status(t)
			accepted, err := w.writer.retry(context.Background(), testUserID, op.ID, op.Generation)
			if err != nil {
				t.Fatal(err)
			}
			owner, generation := testUserID, op.Generation
			switch scope {
			case "owner":
				owner = "another-owner"
			case "installation":
				selectionSQL(t, w.f.db, "UPDATE ai_installation_identity SET id='replacement' WHERE singleton=1")
			case "generation":
				generation = accepted.Generation
			}
			sends, reads := w.counts()
			_, err = w.writer.retry(context.Background(), owner, op.ID, generation)
			if err == nil {
				t.Fatal("out-of-scope retry received recovery authority", scope)
			}
			if afterSends, afterReads := w.counts(); afterSends != sends || (scope != "generation" && afterReads != reads) {
				t.Fatal("out-of-scope recovery contacted Immich", sends, reads, afterSends, afterReads)
			}
			var attempts, currentGeneration int
			if err := w.f.db.db.QueryRow("SELECT attempts,generation FROM ai_write_targets WHERE operationID=?", op.ID).Scan(&attempts, &currentGeneration); err != nil {
				t.Fatal(err)
			}
			if attempts != accepted.Attempts || currentGeneration != accepted.Generation {
				t.Fatal("rejected recovery changed durable attempt state", attempts, currentGeneration)
			}
		})
	}
}

func TestAIWriteUnacceptedRetryStillRequiresFreshEligibility(t *testing.T) {
	for _, changed := range []string{"GPS", "disabled", "expired", "unavailable"} {
		t.Run(changed, func(t *testing.T) {
			w := newAIWriteFixture(t)
			var offline atomic.Bool
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PATCH" {
					out.WriteHeader(429)
					return true
				}
				if offline.Load() {
					out.WriteHeader(503)
					return true
				}
				return false
			}
			w.run(t)
			op := w.status(t)
			switch changed {
			case "GPS":
				w.mu.Lock()
				w.meta["exifInfo"] = map[string]float64{"latitude": 42, "longitude": 13}
				w.mu.Unlock()
			case "disabled":
				w.mu.Lock()
				w.enabled = false
				w.mu.Unlock()
			case "expired":
				w.f.now = w.f.now.Add(6 * time.Minute)
			case "unavailable":
				offline.Store(true)
			}
			if _, err := w.writer.retry(context.Background(), testUserID, op.ID, op.Generation); err == nil {
				t.Fatal("new retry bypassed fresh eligibility", changed)
			}
			current := w.status(t)
			if current.Generation != op.Generation || current.Attempts != 1 || current.Status != "retryable" {
				t.Fatal("rejected retry changed durable attempt state", current)
			}
			if sends, _ := w.counts(); sends != 1 {
				t.Fatal("rejected retry sent a mutation", sends)
			}
		})
	}
}
