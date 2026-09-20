package main

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestAIImageOverallDeadlineAndCancellationReleaseRetrieval(t *testing.T) {
	for _, stage := range []string{"before", "metadata", "preview"} {
		t.Run(stage, func(t *testing.T) {
			f := newAIImageFixture(t)
			f.service.timeout = 100 * time.Millisecond
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				matches := stage == "metadata" && r.URL.Path == "/api/assets/"+selectionA || stage == "preview" && r.URL.Path == "/api/assets/"+selectionA+"/thumbnail"
				if matches {
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
					<-r.Context().Done()
					return true
				}
				return false
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			expected := context.DeadlineExceeded
			if stage == "before" {
				cancel()
				expected = context.Canceled
			}
			p, err := f.service.prepare(ctx, testUserID, f.store.binding, selectionA)
			if p != nil || !errors.Is(err, expected) {
				t.Fatalf("wrong interruption result: %v", err)
			}
			if stage != "before" && ctx.Err() != nil {
				t.Fatal("preparer had no independent overall deadline")
			}
			calls, _ := f.counts()
			want := 1
			if stage == "before" {
				want = 0
			}
			if stage == "preview" {
				want = 2
			}
			if calls != want {
				t.Fatalf("retried after interruption: %d calls", calls)
			}
		})
	}
}

func TestAIImageCancellationOfFinalMetadataReadPreventsPublication(t *testing.T) {
	f := newAIImageFixture(t)
	f.service.timeout = 100 * time.Millisecond
	var reads atomic.Int32
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path == "/api/assets/"+selectionA && reads.Add(1) == 2 {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return true
		}
		return false
	}
	p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
	if p != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("final interruption was not preserved: %v", err)
	}
	calls, _ := f.counts()
	if calls != 3 {
		t.Fatalf("unexpected calls: %d", calls)
	}
}
