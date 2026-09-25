package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestAIWriteExplicitRetryIsIdempotentBoundedAndRequiresKnownCompletion(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			w := newAIWriteFixture(t)
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method != "PATCH" {
					return false
				}
				if unknown {
					conn, _, _ := out.(http.Hijacker).Hijack()
					conn.Close()
				} else {
					out.WriteHeader(429)
				}
				return true
			}
			w.run(t)
			op := w.status(t)
			h := newAIResultHandler(w.writer.drafts.results, w.image.service)
			body := fmt.Sprintf(`{"generation":%d}`, op.Generation)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					rec := aiRequest(h, "POST", "/ai/write-operations/"+op.ID+"/retry", body, aiTestOrigin, true)
					want := 200
					if unknown {
						want = 409
					}
					if rec.Code != want {
						t.Error(rec.Code, rec.Body.String())
					}
				}()
			}
			close(start)
			wg.Wait()
			if unknown {
				if op = w.status(t); op.Status != "verifying" || op.Attempts != 1 {
					t.Fatal(op)
				}
				return
			}
			w.run(t)
			op = w.status(t)
			if op.Status != "failed" || op.Attempts != 2 {
				t.Fatal(op)
			}
			rec := aiRequest(h, "POST", "/ai/write-operations/"+op.ID+"/retry", fmt.Sprintf(`{"generation":%d}`, op.Generation), aiTestOrigin, true)
			if rec.Code != 409 {
				t.Fatal(rec.Code, rec.Body.String())
			}
			if sends, _ := w.counts(); sends != 2 {
				t.Fatal(sends)
			}
		})
	}
}
