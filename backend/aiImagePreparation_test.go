package main

import (
	"context"
	"testing"
)

func TestAIImagePreparesExactAuthorizedPreview(t *testing.T) {
	f := newAIImageFixture(t)
	p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
	if err != nil || p == nil {
		t.Fatalf("prepare: %v", err)
	}
	defer p.Release()
	info, valid := p.Info()
	if !valid || info.Binding.Asset != selectionA || info.Binding.Owner != testUserID || info.Binding.Installation != f.store.binding || info.Binding.SourceDigest == "" || info.Width != 8 || info.Height != 4 {
		t.Fatalf("bad prepared image: %+v", info)
	}
	calls, bad := f.counts()
	if calls < 2 || bad != 0 {
		t.Fatalf("unexpected upstream requests: calls=%d bad=%d", calls, bad)
	}
}
