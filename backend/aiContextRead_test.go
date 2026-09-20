package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIContextMetadataReadRejectsMalformedAndUnboundedBodies(t *testing.T) {
	for _, failure := range []string{"length", "stream", "compression", "failure", "utf8"} {
		t.Run(failure, func(t *testing.T) {
			f := newAIImageFixture(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch failure {
				case "length":
					w.Header().Set("Content-Length", "1048577")
				case "stream":
					w.(http.Flusher).Flush()
					_, _ = w.Write([]byte(strings.Repeat("a", (1<<20)+1)))
				case "compression":
					w.Header().Set("Content-Encoding", "gzip")
				case "failure":
					http.Error(w, "private error", 503)
				case "utf8":
					_, _ = w.Write([]byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'})
				}
			}))
			defer server.Close()
			f.service.endpoint = server.URL
			data, available, err := (&aiContextPreparer{images: f.service}).read(context.Background(), "private-key", "/api/albums/"+selectionID(90), nil)
			if err == nil || available || len(data) != 0 {
				t.Fatal("malformed or excessive metadata admitted", failure)
			}
		})
	}
}
