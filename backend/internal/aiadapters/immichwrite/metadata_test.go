package immichwrite

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/ai/writepreview"
)

func TestMetadataTransportSendsExactlyOneApprovedNamespaceItem(t *testing.T) {
	const asset = "aaf12459-abcd-4321-8421-aaccff110033"
	const record = "bbf12459-abcd-4321-8421-aaccff110033"
	text := "Exact e\u0301\r\n text "
	choice := drafts.MirrorSelection{Languages: []string{"en"}}
	raw, err := writepreview.BuildMirrorExport(drafts.Draft{Revision: 1, FactsRevision: 1, Descriptions: []drafts.Description{{Language: "en", Text: &text, Status: "complete", FactsRevision: 1}}}, results.Document{}, choice, writepreview.MirrorProvenance{}, record)
	if err != nil {
		t.Fatal(err)
	}
	plan := writepreview.Plan{Version: "mirror-preview-v4", TargetID: asset, DraftRevision: 1, Mirror: &writepreview.MirrorPlan{Key: writepreview.MirrorNamespace, RecordID: record, Value: raw, Selection: choice, Disclosure: writepreview.MirrorDisclosure}}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPut || r.URL.Path != "/api/assets/"+asset+"/metadata" || r.Header.Get("x-api-key") != "private-test" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("wrong request authority")
		}
		var body map[string][]struct {
			Key   string          `json:"key"`
			Value json.RawMessage `json:"value"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 1 || len(body["items"]) != 1 || body["items"][0].Key != writepreview.MirrorNamespace || string(body["items"][0].Value) != string(raw) {
			t.Error("expanded or changed metadata payload")
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	out := (&Transport{Endpoint: server.URL}).SendMetadata(context.Background(), "private-test", plan)
	if !out.CompletionKnown || calls != 1 {
		t.Fatal("metadata transport did not make one bounded attempt", out, calls)
	}
}
