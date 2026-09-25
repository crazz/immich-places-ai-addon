package main

import (
	"net/http"
	"testing"
	"time"
)

func TestAIWriteFirstRecoveryWaitsForPersistedDeadline(t *testing.T) {
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
	w.run(t)
	if _, reads := w.counts(); reads != before {
		t.Fatal("recovery ran before one-second deadline", reads, before)
	}
	w.f.now = w.f.now.Add(time.Second)
	w.run(t)
	if sends, reads := w.counts(); sends != 1 || reads != before+1 {
		t.Fatal(sends, reads, before)
	}
}
