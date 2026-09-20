package images

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func jpegFixture(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(80 + x%170), G: uint8(80 + y%170), A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func testBinding() Binding {
	return Binding{Owner: "owner", Installation: "installation", Asset: "asset", SourceDigest: "source"}
}

func TestPrepareRasterDownscalesWithoutCropping(t *testing.T) {
	limits := DefaultLimits()
	limits.LongEdge = 4
	prepared, err := PrepareRaster(context.Background(), jpegFixture(t, 8, 4), "image/jpeg", testBinding(), limits)
	if err != nil || prepared == nil {
		t.Fatalf("prepare: %v", err)
	}
	defer prepared.Release()
	payload, err := prepared.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(bytes.NewReader(payload))
	if err != nil || img.Bounds().Dx() != 4 || img.Bounds().Dy() != 2 {
		t.Fatalf("expected 4x2 JPEG: %v", err)
	}
	info, valid := prepared.Info()
	if !valid || info.Width != 4 || info.Height != 2 || info.MIME != "image/jpeg" || info.Binding != testBinding() || info.Digest == "" || info.Policy == "" {
		t.Fatalf("invalid metadata: %+v", info)
	}
}

func TestPrepareRasterRejectsDecodedAndSourceBudgetOverflow(t *testing.T) {
	input := jpegFixture(t, 8, 4)
	for _, tc := range []struct {
		name  string
		alter func(*Limits)
	}{
		{"pixels", func(l *Limits) { l.MaxPixels = 31 }},
		{"dimension", func(l *Limits) { l.MaxDimension = 7 }},
		{"source", func(l *Limits) { l.MaxSourceBytes = len(input) - 1 }},
		{"invalid policy", func(l *Limits) { l.LongEdge = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			limits := DefaultLimits()
			tc.alter(&limits)
			p, err := PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
			if err == nil || p != nil {
				t.Fatal("over-budget input accepted")
			}
		})
	}
	limits := DefaultLimits()
	limits.MaxPixels = 32
	limits.MaxDimension = 8
	limits.MaxSourceBytes = len(input)
	p, err := PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
	if err != nil {
		t.Fatalf("exact boundary rejected: %v", err)
	}
	defer p.Release()
	info, _ := p.Info()
	if info.Width != 8 || info.Height != 4 {
		t.Fatalf("small image enlarged: %+v", info)
	}
}

func TestPrepareRasterBudgetIncludesBase64AndPrefix(t *testing.T) {
	input := jpegFixture(t, 8, 4)
	limits := DefaultLimits()
	p, err := PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := p.Bytes()
	p.Release()
	budget := len("data:image/jpeg;base64,") + 4*((len(data)+2)/3)
	limits.MaxRequestBytes = budget - 1
	p, err = PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
	if err == nil || p != nil {
		t.Fatal("transmission expansion was not bounded")
	}
	limits.MaxRequestBytes = budget
	p, err = PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
	if err != nil {
		t.Fatalf("exact transmission limit rejected: %v", err)
	}
	p.Release()
}
