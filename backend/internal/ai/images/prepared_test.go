package images

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestPreparedOwnsCopiesAndReleaseInvalidatesAliases(t *testing.T) {
	input := jpegFixture(t, 4, 2)
	p, err := PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	initial, _ := p.Bytes()
	original := bytes.Clone(initial)
	returned, _ := p.Bytes()
	returned[0] ^= 255
	input[0] ^= 255
	again, _ := p.Bytes()
	if !bytes.Equal(again, original) {
		t.Fatal("returned slice mutated retained image")
	}
	alias := p
	p.Release()
	p.Release()
	if data, err := alias.Bytes(); err == nil || len(data) != 0 {
		t.Fatal("released image is usable")
	}
	if _, ok := alias.Info(); ok {
		t.Fatal("released metadata remains valid")
	}
	var zero Prepared
	if _, err := zero.Bytes(); err == nil {
		t.Fatal("zero value is usable")
	}
	if _, ok := zero.Info(); ok {
		t.Fatal("zero metadata is valid")
	}
	var missing *Prepared
	if _, err := missing.Bytes(); err == nil {
		t.Fatal("nil image is usable")
	}
	missing.Release()
}

func TestPreparedDefaultFormattingDoesNotDisclosePayload(t *testing.T) {
	p, err := PrepareRaster(context.Background(), jpegFixture(t, 8, 4), "image/jpeg", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Release()
	for _, value := range []any{p, *p, struct{ Image *Prepared }{p}} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		for _, rendered := range []string{string(encoded), fmt.Sprintf("%v", value), fmt.Sprintf("%+v", value), fmt.Sprintf("%#v", value)} {
			if strings.Contains(rendered, "owner") || strings.Contains(rendered, "source") || strings.Contains(rendered, "255, 216") || strings.Contains(rendered, "255 216") {
				t.Fatalf("private data disclosed: %.120s", rendered)
			}
		}
	}
}

func TestPreparedRequiresCompleteBindingAndIdentifiesPolicy(t *testing.T) {
	input := jpegFixture(t, 8, 4)
	binding := testBinding()
	for _, alter := range []func(*Binding){func(b *Binding) { b.Owner = "" }, func(b *Binding) { b.Installation = "" }, func(b *Binding) { b.Asset = "" }, func(b *Binding) { b.SourceDigest = "" }} {
		invalid := binding
		alter(&invalid)
		p, err := PrepareRaster(context.Background(), input, "image/jpeg", invalid, DefaultLimits())
		if err == nil || p != nil {
			t.Fatal("incomplete binding accepted")
		}
	}
	limits := DefaultLimits()
	first, err := PrepareRaster(context.Background(), input, "image/jpeg", binding, limits)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	a, _ := first.Info()
	limits.LongEdge = 1024
	second, err := PrepareRaster(context.Background(), input, "image/jpeg", binding, limits)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Release()
	b, _ := second.Info()
	if a.Policy == b.Policy {
		t.Fatal("different preparation policies share a binding")
	}
}

func TestPreparedConcurrentReadersAndRelease(t *testing.T) {
	p, err := PrepareRaster(context.Background(), jpegFixture(t, 8, 4), "image/jpeg", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	copied := *p
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				data, err := copied.Bytes()
				if err == nil && (len(data) < 2 || data[0] != 255 || data[1] != 216) {
					t.Error("partial JPEG published")
				}
				copied.Info()
			}
		}()
	}
	p.Release()
	wg.Wait()
	if _, err := copied.Bytes(); err == nil {
		t.Fatal("copied handle survived release")
	}
}
