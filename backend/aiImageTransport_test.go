package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIImageRejectsEveryRedirectWithoutForwardingCredentials(t *testing.T) {
	for _, stage := range []string{"metadata", "preview"} {
		for _, crossHost := range []bool{false, true} {
			t.Run(stage+map[bool]string{false: "/same-host", true: "/cross-host"}[crossHost], func(t *testing.T) {
				f := newAIImageFixture(t)
				var followed atomic.Int32
				target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { followed.Add(1); w.WriteHeader(200) }))
				defer target.Close()
				destination := f.server.URL + "/redirect-target"
				if crossHost {
					destination = target.URL
				}
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if r.URL.Path == "/redirect-target" {
						followed.Add(1)
						return true
					}
					isPreview := strings.HasSuffix(r.URL.Path, "/thumbnail")
					if (stage == "preview") == isPreview {
						http.Redirect(w, r, destination, 302)
						return true
					}
					return false
				}
				p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
				if err == nil || p != nil || followed.Load() != 0 {
					t.Fatalf("redirect followed: %d, %v", followed.Load(), err)
				}
			})
		}
	}
}

func TestAIImageTransportRejectsCompressedResponsesAndUnsafeConfiguration(t *testing.T) {
	f := newAIImageFixture(t)
	for _, endpoint := range []string{"file:///private/image", "http://user:secret@example.com", "https://example.com?secret=value", "https://example.com/#private", "http:///missing-host"} {
		if service, err := newAIImagePreparer(f.db, f.store, endpoint); err == nil || service != nil {
			t.Fatal("unsafe configured endpoint accepted")
		}
	}
	body := &aiImageBody{Reader: strings.NewReader("private-image")}
	f.service.client.Transport = aiImageRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, ContentLength: -1, Header: http.Header{"Content-Encoding": []string{"gzip"}}}, nil
	})
	data, _, err := f.service.fetch(context.Background(), "private-key", "/api/assets/"+selectionA, 100)
	if err == nil || data != nil || body.read != 0 || !body.closed {
		t.Fatal("compressed body accepted or retained")
	}
}
