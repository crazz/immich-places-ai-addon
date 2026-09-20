package images

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestPrepareRasterCompositesTransparencyOverWhite(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 20; x < 40; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	p, err := PrepareRaster(context.Background(), b.Bytes(), "image/png", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Release()
	data, _ := p.Bytes()
	result, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	r, g, blue, _ := result.At(5, 10).RGBA()
	if r < 64000 || g < 64000 || blue < 64000 {
		t.Fatal("transparent pixels did not become white")
	}
	r, g, blue, _ = result.At(35, 10).RGBA()
	if r < 60000 || g > 5000 || blue > 5000 {
		t.Fatal("opaque content moved or changed")
	}
}
