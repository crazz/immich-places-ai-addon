package main

import (
	"bytes"
	"context"
	"image/jpeg"
	"testing"
)

func TestAIResultThumbnailRequiresFreshReadAuthorityWithExecutionDisabled(t *testing.T) {
	f, p, req := productionFixture(t)
	productionSession(t, f)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	p.store.enabled = false
	f.image.store.enabled = false
	h := newAIResultHandler(&aiResultStore{jobs: p.store}, f.image.service)
	before, _ := f.image.counts()
	rec := aiRequest(h, "GET", "/ai/jobs/"+job.ID+"/items/"+job.Items[0].ID+"/thumbnail", "", "", true)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/jpeg" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("current thumbnail unavailable", rec.Code)
	}
	image, err := jpeg.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil || image.Bounds().Dx() != 8 {
		t.Fatal("invalid thumbnail", err)
	}
	reads, bad := f.image.counts()
	if reads-before != 3 || bad != 0 || f.hits.Load() != 0 {
		t.Fatal("unexpected image authority calls", reads, bad)
	}
}
