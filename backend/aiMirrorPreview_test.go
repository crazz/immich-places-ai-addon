package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorPreviewFreezesSelectedExportAndDisclosureAcrossReopen(t *testing.T) {
	f := standardWriteFixture(t, []string{"gps"})
	ctx := context.Background()
	f.writer.capabilities.Capabilities = []string{"description", "metadata"}
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/assets/"+f.draft.AssetID+"/metadata" {
			return false
		}
		if r.Method != "GET" {
			t.Error("preview performed mutation")
		}
		_, _ = w.Write([]byte(`[]`))
		return true
	}
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"languages":["en"]}`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}
	preview, err := writepreview.Create(ctx, session, session, session, testUserID, d.ID, d.Revision, uuid.NewString(), f.f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Plan.Version != "mirror-preview-v4" || preview.Plan.Mirror == nil || preview.Plan.Mirror.Disclosure != writepreview.MirrorDisclosure || preview.Plan.Mirror.Before.Present || len(preview.Plan.Manifest.Targets) != 1 {
		t.Fatal("selected mirror lost from exact preview")
	}
	var export writepreview.MirrorExport
	if json.Unmarshal(preview.Plan.Mirror.Value, &export) != nil || export.Descriptions["en"] != "Exact e\u0301\r\n scene  " || export.Review.DraftRevision != d.Revision {
		t.Fatal("reviewed text/revision changed")
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	retained, err := session.get(ctx, testUserID, preview.Plan.ID)
	if err != nil || !reflect.DeepEqual(preview, retained) {
		t.Fatal("v4 bytes lost across reopen", err)
	}
	if _, err := session.get(ctx, "foreign", preview.Plan.ID); err == nil {
		t.Fatal("foreign owner read mirror")
	}
}
