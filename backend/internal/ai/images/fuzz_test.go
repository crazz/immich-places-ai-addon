package images

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func FuzzPrepareRaster(f *testing.F) {
	var jpegData bytes.Buffer
	if err := jpeg.Encode(&jpegData, image.NewRGBA(image.Rect(0, 0, 8, 4)), nil); err != nil {
		f.Fatal(err)
	}
	f.Add(jpegData.Bytes(), uint8(0))
	webp, err := os.ReadFile("testdata/static.webp")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(webp, uint8(2))
	f.Add([]byte("not an image"), uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, format uint8) {
		if len(data) > 65536 {
			return
		}
		limits := Limits{LongEdge: 32, MaxPixels: 4096, MaxDimension: 256, MaxSourceBytes: 65536, MaxRequestBytes: 8192}
		p, err := PrepareRaster(context.Background(), data, []string{"image/jpeg", "image/png", "image/webp"}[format%3], testBinding(), limits)
		if err != nil {
			if p != nil {
				t.Fatal("failure exposed a partial result")
			}
			return
		}
		defer p.Release()
		payload, err := p.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		img, err := jpeg.Decode(bytes.NewReader(payload))
		if err != nil {
			t.Fatal("result is not a JPEG")
		}
		if img.Bounds().Dx() > 32 || img.Bounds().Dy() > 32 || 22+4*((len(payload)+2)/3) > 8192 {
			t.Fatal("result violates output bounds")
		}
	})
}
