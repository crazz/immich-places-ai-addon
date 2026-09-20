package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type aiImageRoundTrip func(*http.Request) (*http.Response, error)

func (f aiImageRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type aiImageBody struct {
	io.Reader
	read   int
	closed bool
}

func (b *aiImageBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += n
	return n, err
}
func (b *aiImageBody) Close() error { b.closed = true; return nil }

func TestAIImageTransportBoundsDeclaredAndStreamedBytes(t *testing.T) {
	for _, tc := range []struct {
		name          string
		declared      int64
		size, maxRead int
		success       bool
	}{
		{"declared excess", 21, 30, 0, false},
		{"unknown length", -1, 30, 21, false},
		{"misleading length", 5, 30, 21, false},
		{"exact boundary", 20, 20, 20, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAIImageFixture(t)
			body := &aiImageBody{Reader: strings.NewReader(strings.Repeat("x", tc.size))}
			attempts := 0
			f.service.client.Transport = aiImageRoundTrip(func(*http.Request) (*http.Response, error) {
				attempts++
				return &http.Response{StatusCode: 200, ContentLength: tc.declared, Body: body, Header: http.Header{"Content-Type": []string{"image/jpeg"}}}, nil
			})
			data, _, err := f.service.fetch(context.Background(), "private-key", "/api/assets/"+selectionA, 20)
			if (err == nil) != tc.success || (!tc.success && data != nil) || body.read > tc.maxRead || !body.closed || attempts != 1 {
				t.Fatalf("unbounded fetch: read=%d closed=%v attempts=%d err=%v", body.read, body.closed, attempts, err)
			}
		})
	}
}
