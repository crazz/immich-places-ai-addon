package immichwrite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/ai/writepreview"
)

func TestMetadataTransportNeverRetriesRedirectsOrAmbiguousFailures(t *testing.T) {
	for _, scenario := range []string{"redirect", "drop", "timeout", "oversize", "throttled", "server", "accepted", "request-timeout"} {
		t.Run(scenario, func(t *testing.T) {
			const asset = "aaf12459-abcd-4321-8421-aaccff110033"
			const record = "bbf12459-abcd-4321-8421-aaccff110033"
			text := "reviewed text"
			choice := drafts.MirrorSelection{Languages: []string{"en"}}
			raw, err := writepreview.BuildMirrorExport(drafts.Draft{Revision: 1, FactsRevision: 1, Descriptions: []drafts.Description{{Language: "en", Text: &text, Status: "complete", FactsRevision: 1}}}, results.Document{}, choice, writepreview.MirrorProvenance{}, record)
			if err != nil {
				t.Fatal(err)
			}
			plan := writepreview.Plan{Version: "mirror-preview-v4", TargetID: asset, DraftRevision: 1, Mirror: &writepreview.MirrorPlan{Key: writepreview.MirrorNamespace, RecordID: record, Value: raw, Selection: choice, Disclosure: writepreview.MirrorDisclosure}}
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				switch scenario {
				case "redirect":
					w.Header().Set("Location", "/alternate")
					w.WriteHeader(307)
				case "drop":
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Error(err)
						return
					}
					_ = conn.Close()
				case "timeout":
					select {
					case <-r.Context().Done():
					case <-time.After(500 * time.Millisecond):
					}
				case "oversize":
					_, _ = w.Write([]byte(strings.Repeat("x", (64<<10)+1)))
				case "throttled":
					w.WriteHeader(429)
				case "server":
					w.WriteHeader(500)
				case "accepted":
					w.WriteHeader(202)
				case "request-timeout":
					w.WriteHeader(408)
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			result := (&Transport{Endpoint: server.URL}).SendMetadata(ctx, "private-test", plan)
			if calls.Load() != 1 || result.CompletionKnown != (scenario == "throttled") {
				t.Fatal("hidden retry or invented completion", calls.Load(), result)
			}
		})
	}
}
