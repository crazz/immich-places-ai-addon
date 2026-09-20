package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIImageFailuresAreSafeAndNeverRetried(t *testing.T) {
	for _, stage := range []string{"metadata", "preview"} {
		for _, status := range []int{200, 401, 403, 404, 429, 500} {
			t.Run(fmt.Sprintf("%s/%d", stage, status), func(t *testing.T) {
				f := newAIImageFixture(t)
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if (stage == "preview") == strings.HasSuffix(r.URL.Path, "/thumbnail") {
						w.WriteHeader(status)
						_, _ = io.WriteString(w, "private-immich-image-key https://private.example/photo GPS=private")
						return true
					}
					return false
				}
				p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
				calls, _ := f.counts()
				want := 1
				if stage == "preview" {
					want = 2
				}
				if err == nil || p != nil || calls != want {
					t.Fatalf("failure retried or published: calls=%d err=%v", calls, err)
				}
				for _, private := range []string{"private", "http", "GPS"} {
					if strings.Contains(err.Error(), private) {
						t.Fatal("upstream failure disclosed")
					}
				}
			})
		}
	}
}

type aiImageBrokenReader struct{}

func (aiImageBrokenReader) Read([]byte) (int, error) { return 0, errors.New("private upstream stream") }

func TestAIImageInterruptedStreamClosesBodyWithoutPrivateError(t *testing.T) {
	f := newAIImageFixture(t)
	body := &aiImageBody{Reader: aiImageBrokenReader{}}
	f.service.client.Transport = aiImageRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, ContentLength: -1, Body: body, Header: http.Header{}}, nil
	})
	data, _, err := f.service.fetch(context.Background(), "key", "/api/assets/"+selectionA, 100)
	if !errors.Is(err, errAIImageUpstream) || data != nil || !body.closed || strings.Contains(err.Error(), "private") {
		t.Fatalf("stream failed unsafely: %v", err)
	}
}

func TestAIImageFinalSourceReadPreservesUpstreamFailureCategory(t *testing.T) {
	f := newAIImageFixture(t)
	var metadataReads atomic.Int32
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path == "/api/assets/"+selectionA && metadataReads.Add(1) == 2 {
			w.WriteHeader(503)
			return true
		}
		return false
	}
	p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
	if p != nil || !errors.Is(err, errAIImageUpstream) {
		t.Fatalf("temporary source failure became another category: %v", err)
	}
	calls, _ := f.counts()
	if calls != 3 {
		t.Fatal("final source failure retried")
	}
}
